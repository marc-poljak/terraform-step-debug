package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/marc-poljak/terraform-step-debug/internal/executor"
	"github.com/marc-poljak/terraform-step-debug/internal/model"
	"github.com/marc-poljak/terraform-step-debug/internal/parser"
	"github.com/marc-poljak/terraform-step-debug/internal/ui"
	"github.com/marc-poljak/terraform-step-debug/internal/util"
)

var (
	// Command line flags
	terraformDir  = flag.String("dir", "", "Path to the Terraform directory (default: current directory)")
	planFile      = flag.String("plan", "", "Path to the Terraform plan file (default: generate new plan)")
	terraformPath = flag.String("terraform", "", "Path to the Terraform binary (default: use from PATH)")
	dryRun        = flag.Bool("dry-run", false, "Perform a dry run without actually applying changes")
	targetAddr    = flag.String("target", "", "Target a specific resource (default: all resources)")
	version       = flag.Bool("version", false, "Print version information and exit")
	varFile       = flag.String("var-file", "", "Path to the Terraform variable file (e.g., prod.tfvars)")
)

// Version information, to be set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Print version if requested
	if handleVersionFlag() {
		return
	}

	// Setup environment
	if err := setupEnvironment(); err != nil {
		exitWithError(err)
	}

	// Setup UI and parser
	ui := ui.NewUI()
	planParser := parser.NewTerraformPlanParser(*terraformPath)

	// Handle plan file
	cleanup, err := handlePlanFile(planParser)
	if err != nil {
		exitWithError(err)
	}
	if cleanup {
		defer util.CleanupFiles(*planFile)
	}

	// Parse the plan
	plan, err := planParser.ParsePlan(*planFile, *terraformDir)
	if err != nil {
		exitWithError(fmt.Errorf("error parsing plan: %w", err))
	}

	// Check if there are changes and handle target resource
	if !handlePlanChanges(plan) {
		return
	}

	// Build execution graph and run the executor
	executionGraph := planParser.BuildExecutionGraph(plan)
	executer := executor.NewTerraformExecutor(*terraformPath, *terraformDir, *planFile, *varFile, *dryRun)

	// Display the plan summary
	ui.DisplayPlanSummary(plan)

	// Execute the plan
	executedResources := executeResources(ui, executer, executionGraph, plan, *targetAddr)

	// Display summary and exit
	ui.DisplaySummary(executedResources)
	fmt.Println("Execution complete.")
}

// handleVersionFlag handles the version flag and returns true if the program should exit
func handleVersionFlag() bool {
	if *version {
		fmt.Printf("terraform-step-debug version %s\n", Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Commit: %s\n", Commit)
		os.Exit(0)
		return true
	}
	return false
}

// setupEnvironment sets up the terraform path and directory
func setupEnvironment() error {
	var err error

	// Find Terraform binary if not specified
	if *terraformPath == "" {
		*terraformPath, err = util.FindTerraformBinary()
		if err != nil {
			return err
		}
	}

	// Check Terraform version
	if err := util.CheckTerraformVersion(*terraformPath); err != nil {
		return err
	}

	// Find Terraform directory if not specified
	if *terraformDir == "" {
		*terraformDir, err = util.FindTerraformDir("")
		if err != nil {
			return err
		}
	}

	return nil
}

// handlePlanFile generates a plan file if needed and returns whether cleanup is needed
func handlePlanFile(planParser *parser.TerraformPlanParser) (bool, error) {
	cleanup := false

	// If no plan file is specified, generate one
	if *planFile == "" {
		var err error
		*planFile, err = util.CreateTempPlanFile()
		if err != nil {
			return false, err
		}
		cleanup = true

		fmt.Printf("Generating Terraform plan to %s...\n", *planFile)
		if err := planParser.GeneratePlan(*terraformDir, *planFile, *varFile); err != nil {
			return cleanup, fmt.Errorf("error generating plan: %w", err)
		}
	}

	return cleanup, nil
}

// handlePlanChanges checks if there are changes and validates target resources
// Returns false if the program should exit
func handlePlanChanges(plan *model.Plan) bool {
	// Check if there are any changes
	if !plan.HasChanges {
		fmt.Println("No changes to apply.")
		return false
	}

	// If a target resource is specified, validate it
	if *targetAddr != "" {
		emptyResourceMap := make(map[string]*struct{})
		for addr := range plan.ResourcesMap {
			emptyResourceMap[addr] = nil
		}

		if err := util.ValidateTargetResource(*targetAddr, emptyResourceMap); err != nil {
			exitWithError(err)
			return false
		}
	}

	return true
}

// executeResources executes the planned resources
func executeResources(ui *ui.UI, executer *executor.TerraformExecutor,
	executionGraph *model.ExecutionGraph, plan *model.Plan, targetAddr string) []*model.Resource {

	var executedResources []*model.Resource

	// Iterate through each layer of the execution graph
	for layerIndex, layer := range executionGraph.Layers {
		fmt.Printf("Executing layer %d of %d\n", layerIndex+1, len(executionGraph.Layers))

		// Process each resource in the layer
		for _, resource := range layer {
			// Skip resources that are not targeted, if a target is specified
			if targetAddr != "" && resource.Address != targetAddr {
				continue
			}

			// Display resource information
			totalResources := len(plan.Resources)
			currentIndex := len(executedResources) + 1
			ui.DisplayResourceInfo(resource, currentIndex, totalResources)

			// Process the user's action for this resource
			if processResourceAction(ui, executer, resource) {
				executedResources = append(executedResources, resource)
			}
		}
	}

	return executedResources
}

// processResourceAction handles user actions for a resource
// Returns true if the resource was processed and added to executed resources
func processResourceAction(ui *ui.UI, executer *executor.TerraformExecutor, resource *model.Resource) bool {
	for {
		// Get the user action
		action, err := ui.GetUserAction()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting user action: %s\n", err)
			continue
		}

		// Handle abort action
		if action == model.StepAbort {
			if confirmAbort() {
				fmt.Println("Execution aborted.")
				os.Exit(0)
			}
			continue
		}

		// Handle detail action
		if action == model.StepDetail {
			if err := executer.ExecuteStepAction(action, resource); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
			continue
		}

		// Execute the action
		startTime := time.Now()
		err = executer.ExecuteStepAction(action, resource)
		elapsed := time.Since(startTime)

		// Display the result
		ui.DisplayExecutionResult(resource, err == nil, elapsed)

		// If there was an error, ask if the user wants to continue
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			if !ui.ConfirmContinue() {
				fmt.Println("Execution aborted due to errors.")
				os.Exit(1)
			}
		}

		return true
	}
}

// confirmAbort asks the user to confirm aborting the execution
func confirmAbort() bool {
	fmt.Print("Are you sure you want to abort? [y/n]: ")
	var confirm string
	if _, err := fmt.Scanln(&confirm); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading confirmation: %s\n", err)
		return false
	}
	return confirm == "y" || confirm == "Y"
}

// exitWithError prints an error message and exits with code 1
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}
