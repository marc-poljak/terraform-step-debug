package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/marc-poljak/terraform-step-debug/internal/model"
)

// TerraformPlanParser is responsible for parsing Terraform plan files
type TerraformPlanParser struct {
	terraformPath string
}

// NewTerraformPlanParser creates a new TerraformPlanParser
func NewTerraformPlanParser(terraformPath string) *TerraformPlanParser {
	if terraformPath == "" {
		terraformPath = "terraform" // Default to using terraform from PATH
	}
	return &TerraformPlanParser{
		terraformPath: terraformPath,
	}
}

// GeneratePlan generates a new Terraform plan file
func (p *TerraformPlanParser) GeneratePlan(terraformDir, outFile, varFile string) error {
	// Build the command
	args := []string{"plan", "-out", outFile}

	// Add var-file if specified
	if varFile != "" {
		args = append(args, "-var-file", varFile)
	}

	// For Terraform 1.11.x, we use proper argument separation
	cmd := exec.Command(p.terraformPath, args...)
	cmd.Dir = terraformDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ParsePlan parses a Terraform plan file and returns a model.Plan
func (p *TerraformPlanParser) ParsePlan(planFile, terraformDir string) (*model.Plan, error) {
	// Convert the plan to JSON format
	jsonData, err := p.convertPlanToJSON(planFile)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plan to JSON: %w", err)
	}

	// Parse the JSON data
	var planData map[string]interface{}
	if err := json.Unmarshal(jsonData, &planData); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	// Create a new plan
	plan := model.NewPlan(planFile, terraformDir)

	// Extract resources from the plan
	if err := p.extractResources(planData, plan); err != nil {
		return nil, fmt.Errorf("failed to extract resources: %w", err)
	}

	// Calculate plan statistics
	p.calculatePlanStats(plan)

	// Resolve dependencies
	p.resolveDependencies(plan, planData)

	return plan, nil
}

// convertPlanToJSON runs 'terraform show -json' on the plan file
func (p *TerraformPlanParser) convertPlanToJSON(planFile string) ([]byte, error) {
	cmd := exec.Command(p.terraformPath, "show", "-json", planFile)
	return cmd.Output()
}

// extractResources extracts resources from the plan data
func (p *TerraformPlanParser) extractResources(planData map[string]interface{}, plan *model.Plan) error {
	// Check if the plan has changes
	if planChanges, ok := planData["resource_changes"].([]interface{}); ok {
		for _, change := range planChanges {
			changeMap, ok := change.(map[string]interface{})
			if !ok {
				continue
			}

			// Skip resources with no actions
			actions, ok := changeMap["change"].(map[string]interface{})["actions"].([]interface{})
			if !ok || len(actions) == 0 {
				continue
			}

			// Get the primary action (create, update, delete)
			var action model.Action
			switch actions[0].(string) {
			case "create":
				action = model.ActionCreate
				plan.Stats.Create++
			case "update":
				action = model.ActionUpdate
				plan.Stats.Update++
			case "delete":
				action = model.ActionDelete
				plan.Stats.Delete++
			case "read":
				action = model.ActionRead
			case "no-op":
				action = model.ActionNoop
				plan.Stats.Noop++
				continue // Skip no-op resources
			default:
				continue // Skip unknown actions
			}

			// Create a new resource
			address := changeMap["address"].(string)
			parts := strings.Split(address, ".")
			resourceType := parts[0]
			resourceName := strings.Join(parts[1:], ".")

			resource := &model.Resource{
				Address:      address,
				Type:         resourceType,
				Name:         resourceName,
				Action:       action,
				Dependencies: []string{},
				Attributes:   extractAttributes(changeMap),
				Status:       model.StatusPending,
				Warnings:     extractWarnings(changeMap),
			}

			// Add the resource to the plan
			plan.Resources = append(plan.Resources, resource)
			plan.ResourcesMap[address] = resource
		}
	}

	plan.HasChanges = len(plan.Resources) > 0
	return nil
}

// extractAttributes extracts attributes from a resource change
func extractAttributes(changeMap map[string]interface{}) map[string]any {
	attributes := make(map[string]any)

	// Get the "after" state for creates/updates, or "before" state for deletes
	change, ok := changeMap["change"].(map[string]interface{})
	if !ok {
		return attributes
	}

	var values map[string]interface{}

	if after, ok := change["after"].(map[string]interface{}); ok && after != nil {
		values = after
	} else if before, ok := change["before"].(map[string]interface{}); ok && before != nil {
		values = before
	}

	// Copy values to attributes
	for k, v := range values {
		attributes[k] = v
	}

	return attributes
}

// extractWarnings extracts warnings from a resource change
func extractWarnings(changeMap map[string]interface{}) []string {
	warnings := []string{}

	if warningsData, ok := changeMap["change"].(map[string]interface{})["warnings"].([]interface{}); ok {
		for _, warning := range warningsData {
			if w, ok := warning.(string); ok {
				warnings = append(warnings, w)
			}
		}
	}

	return warnings
}

// calculatePlanStats calculates statistics for the plan
func (p *TerraformPlanParser) calculatePlanStats(plan *model.Plan) {
	// Stats are already calculated during resource extraction
	// This function is a placeholder for any additional statistics
}

// resolveDependencies resolves dependencies between resources
func (p *TerraformPlanParser) resolveDependencies(plan *model.Plan, planData map[string]interface{}) {
	// Extract configuration data which contains dependency information
	configResources, ok := planData["configuration"].(map[string]interface{})["root_module"].(map[string]interface{})["resources"].([]interface{})
	if !ok {
		return
	}

	// Create a map of resource addresses to their dependencies
	depMap := make(map[string][]string)
	for _, res := range configResources {
		resMap, ok := res.(map[string]interface{})
		if !ok {
			continue
		}

		// Get the resource address
		addrMode := resMap["mode"].(string)
		addrType := resMap["type"].(string)
		addrName := resMap["name"].(string)
		address := fmt.Sprintf("%s.%s", addrType, addrName)
		if addrMode == "data" {
			address = fmt.Sprintf("data.%s", address)
		}

		// Get the dependencies
		var deps []string
		if depsData, ok := resMap["depends_on"].([]interface{}); ok {
			for _, dep := range depsData {
				if depStr, ok := dep.(string); ok {
					deps = append(deps, depStr)
				}
			}
		}

		// Extract any implicit dependencies from expressions
		if expressions, ok := resMap["expressions"].(map[string]interface{}); ok {
			for _, expr := range expressions {
				if exprMap, ok := expr.(map[string]interface{}); ok {
					if refs, ok := exprMap["references"].([]interface{}); ok {
						for _, ref := range refs {
							if refStr, ok := ref.(string); ok {
								// Only add if it's a resource reference (not a variable)
								if strings.Contains(refStr, ".") && !strings.HasPrefix(refStr, "var.") {
									deps = append(deps, refStr)
								}
							}
						}
					}
				}
			}
		}

		depMap[address] = deps
	}

	// Assign dependencies to resources
	for _, resource := range plan.Resources {
		if deps, ok := depMap[resource.Address]; ok {
			resource.Dependencies = deps
		}
	}
}

// BuildExecutionGraph builds an execution graph based on resource dependencies
func (p *TerraformPlanParser) BuildExecutionGraph(plan *model.Plan) *model.ExecutionGraph {
	graph := &model.ExecutionGraph{
		Layers: [][]*model.Resource{},
	}

	// Copy the resources to avoid modifying the original plan
	pendingResources := make(map[string]*model.Resource)

	// Copy resources from the plan to pendingResources map
	for addr := range plan.ResourcesMap {
		pendingResources[addr] = plan.ResourcesMap[addr]
	}

	// Process resources until none are left
	for len(pendingResources) > 0 {
		// Create a new layer for resources that have no pending dependencies
		currentLayer := []*model.Resource{}

		// Find resources with no pending dependencies
		var readyAddresses []string
		for addr, resource := range pendingResources {
			// Check if this resource has any pending dependencies
			hasPendingDeps := false
			for _, depAddr := range resource.Dependencies {
				if _, exists := pendingResources[depAddr]; exists {
					hasPendingDeps = true
					break
				}
			}

			// If no pending dependencies, add to current layer
			if !hasPendingDeps {
				currentLayer = append(currentLayer, resource)
				readyAddresses = append(readyAddresses, addr)
			}
		}

		// If we found no resources but there are still pending ones,
		// we have a circular dependency
		if len(readyAddresses) == 0 && len(pendingResources) > 0 {
			// Break the circular dependency by picking the resource with
			// the least pending dependencies
			var bestAddr string
			minDeps := -1

			for addr, resource := range pendingResources {
				// Count pending dependencies
				depCount := 0
				for _, depAddr := range resource.Dependencies {
					if _, exists := pendingResources[depAddr]; exists {
						depCount++
					}
				}

				// Update if this resource has fewer dependencies
				if minDeps == -1 || depCount < minDeps {
					minDeps = depCount
					bestAddr = addr
				}
			}

			// Add the selected resource to the current layer
			if bestAddr != "" {
				currentLayer = append(currentLayer, pendingResources[bestAddr])
				readyAddresses = append(readyAddresses, bestAddr)
			}
		}

		// Remove processed resources from pending list
		for _, addr := range readyAddresses {
			delete(pendingResources, addr)
		}

		// Add the current layer to the graph
		if len(currentLayer) > 0 {
			graph.Layers = append(graph.Layers, currentLayer)
		}
	}

	return graph
}
