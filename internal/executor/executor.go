package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/marc-poljak/terraform-step-debug/internal/model"
)

// TerraformExecutor handles the execution of Terraform operations
type TerraformExecutor struct {
	terraformPath string
	terraformDir  string
	planFile      string
	varFile       string
	dryRun        bool
}

// NewTerraformExecutor creates a new TerraformExecutor
func NewTerraformExecutor(terraformPath, terraformDir, planFile, varFile string, dryRun bool) *TerraformExecutor {
	if terraformPath == "" {
		terraformPath = "terraform" // Default to using terraform from PATH
	}

	return &TerraformExecutor{
		terraformPath: terraformPath,
		terraformDir:  terraformDir,
		planFile:      planFile,
		varFile:       varFile,
		dryRun:        dryRun,
	}
}

// ApplyResource applies a single resource from the plan
func (e *TerraformExecutor) ApplyResource(resource *model.Resource) error {
	fmt.Printf("Applying resource: %s (%s)\n", resource.Address, resource.Action)

	if e.dryRun {
		fmt.Println("[DRY RUN] Would apply this resource")
		time.Sleep(500 * time.Millisecond) // Simulate execution time
		return nil
	}

	// Build the command to apply the specific resource
	// For Terraform 1.11.x, we use -target as separate arguments
	args := []string{
		"apply",
		"-auto-approve",
		"-target",
		resource.Address,
	}

	// Add var-file if specified
	if e.varFile != "" {
		args = append(args, "-var-file", e.varFile)
	}

	cmd := exec.Command(e.terraformPath, args...)

	cmd.Dir = e.terraformDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the command
	err := cmd.Run()
	if err != nil {
		resource.Status = model.StatusFailed
		return fmt.Errorf("failed to apply resource %s: %w", resource.Address, err)
	}

	resource.Status = model.StatusComplete
	return nil
}

// GetResourceDetails retrieves detailed information about a resource
func (e *TerraformExecutor) GetResourceDetails(resource *model.Resource) (string, error) {
	// For an existing resource, we can use terraform state show
	// For a planned resource, we need to use the plan file

	var cmd *exec.Cmd
	if resource.Action == model.ActionCreate {
		// For creates, use the plan file
		cmd = exec.Command(
			e.terraformPath,
			"show",
			"-json",
			e.planFile,
		)
	} else {
		// For updates/deletes, try to get current state
		cmd = exec.Command(
			e.terraformPath,
			"state",
			"show",
			resource.Address,
		)
	}

	cmd.Dir = e.terraformDir
	output, err := cmd.Output()
	if err != nil {
		// If state show fails (resource might not exist), fall back to plan
		cmd = exec.Command(
			e.terraformPath,
			"show",
			"-json",
			e.planFile,
		)
		cmd.Dir = e.terraformDir
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get resource details: %w", err)
		}
	}

	// For JSON output, we'd need to extract the specific resource
	// For simplicity, we'll just return the raw output for now
	return string(output), nil
}

// AbortPlan aborts the current plan execution
func (e *TerraformExecutor) AbortPlan() error {
	fmt.Println("Aborting plan execution")
	return nil
}

// GetResourceDiff gets the diff for a specific resource
func (e *TerraformExecutor) GetResourceDiff(resource *model.Resource) (string, error) {
	// Use terraform plan with -target to get the diff for a specific resource
	// For Terraform 1.11.x, we use -target as separate arguments
	args := []string{
		"plan",
		"-target",
		resource.Address,
	}

	// Add var-file if specified
	if e.varFile != "" {
		args = append(args, "-var-file", e.varFile)
	}

	cmd := exec.Command(e.terraformPath, args...)
	cmd.Dir = e.terraformDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get resource diff: %w", err)
	}

	// Extract the diff section from the output
	// This is a simple approach and might need improvement
	outputStr := string(output)

	// Split the output by resource
	sections := strings.Split(outputStr, "# ")

	// Find the section for our resource
	for _, section := range sections {
		if strings.HasPrefix(section, resource.Address) {
			return "# " + section, nil
		}
	}

	return outputStr, nil
}

// ExecuteStepAction executes a step action on a resource
func (e *TerraformExecutor) ExecuteStepAction(action model.StepAction, resource *model.Resource) error {
	switch action {
	case model.StepApply:
		// Mark the resource as approved
		resource.Status = model.StatusApproved
		// Apply the resource
		return e.ApplyResource(resource)

	case model.StepSkip:
		// Mark the resource as skipped
		resource.Status = model.StatusSkipped
		fmt.Printf("Skipping resource: %s\n", resource.Address)
		return nil

	case model.StepAbort:
		// Abort the plan
		return e.AbortPlan()

	case model.StepDetail:
		// Show resource details
		details, err := e.GetResourceDiff(resource)
		if err != nil {
			return fmt.Errorf("failed to get resource details: %w", err)
		}
		fmt.Println("\nResource Details:")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println(details)
		fmt.Println(strings.Repeat("-", 80))
		return nil

	default:
		return fmt.Errorf("unknown step action: %s", action)
	}
}
