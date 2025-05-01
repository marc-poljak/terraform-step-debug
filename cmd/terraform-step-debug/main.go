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

	// Print version information if requested
	if *version {
		fmt.Printf("terraform-step-debug version %s\n", Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Commit: %s\n", Commit)
		os.Exit(0)
	}

	// Find Terraform binary if not specified
	if *terraformPath == "" {
		var err error
		*terraformPath, err = util.FindTerraformBinary()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	}

	// Check Terraform version
	if err := util.CheckTerraformVersion(*terraformPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Find Terraform directory if not specified
	if *terraformDir == "" {
		var err error
		*terraformDir, err = util.FindTerraformDir("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	}

	// Create a new UI
	ui := ui.NewUI()

	// Create a new parser
	planParser := parser.NewTerraformPlanParser(*terraformPath)

	// If no plan file is specified, generate one
	cleanup := false
	if *planFile == "" {
		var err error
		*planFile, err = util.CreateTempPlanFile()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		cleanup = true

		fmt.Printf("Generating Terraform plan to %s...\n", *planFile)
		if err := planParser.GeneratePlan(*terraformDir, *planFile, *varFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating plan: %s\n", err)
			os.Exit(1)
		}
	}

	// Clean up the plan file when we're done if we generated it
	if cleanup {
		defer util.CleanupFiles(*planFile)
	}

	// Parse the plan
	plan, err := planParser.ParsePlan(*planFile, *terraformDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing plan: %s\n", err)
		os.Exit(1)
	}

	// Check if there are any changes
	if !plan.HasChanges {
		fmt.Println("No changes to apply.")
		os.Exit(0)
	}

	// If a target resource is specified, validate it
	if *targetAddr != "" {
		emptyResourceMap := make(map[string]*struct{})
		for addr := range plan.ResourcesMap {
			emptyResourceMap[addr] = nil
		}

		if err := util.ValidateTargetResource(*targetAddr, emptyResourceMap); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	}

	// Build the execution graph
	executionGraph := planParser.BuildExecutionGraph(plan)

	// Create an executor
	executer := executor.NewTerraformExecutor(*terraformPath, *terraformDir, *planFile, *varFile, *dryRun)

	// Display the plan summary
	ui.DisplayPlanSummary(plan)

	// Execute the plan step by step
	var executedResources []*model.Resource

	// Iterate through each layer of the execution graph
	for layerIndex, layer := range executionGraph.Layers {
		fmt.Printf("Executing layer %d of %d\n", layerIndex+1, len(executionGraph.Layers))

		// Process each resource in the layer
		for _, resource := range layer {
			// Skip resources that are not targeted, if a target is specified
			if *targetAddr != "" && resource.Address != *targetAddr {
				continue
			}

			// Display resource information
			totalResources := len(plan.Resources)
			currentIndex := len(executedResources) + 1
			ui.DisplayResourceInfo(resource, currentIndex, totalResources)

			// Handle the resource
			for {
				// Get the user action
				action, err := ui.GetUserAction()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting user action: %s\n", err)
					continue
				}

				// If the action is to abort, confirm and exit
				if action == model.StepAbort {
					fmt.Print("Are you sure you want to abort? [y/n]: ")
					var confirm string
					if _, err := fmt.Scanln(&confirm); err != nil {
						fmt.Fprintf(os.Stderr, "Error reading confirmation: %s\n", err)
						continue
					}
					if confirm == "y" || confirm == "Y" {
						ui.DisplaySummary(executedResources)
						fmt.Println("Execution aborted.")
						os.Exit(0)
					}
					continue
				}

				// If the action is to show details, display them and continue
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
						ui.DisplaySummary(executedResources)
						fmt.Println("Execution aborted due to errors.")
						os.Exit(1)
					}
				}

				// Add the resource to the executed resources
				executedResources = append(executedResources, resource)
				break
			}
		}
	}

	// Display the summary
	ui.DisplaySummary(executedResources)
	fmt.Println("Execution complete.")
}
