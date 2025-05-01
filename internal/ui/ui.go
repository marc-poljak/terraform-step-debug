package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/marc-poljak/terraform-step-debug/internal/model"
)

// Colors for the terminal output
const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
)

// UI handles user interaction during the debugging process
type UI struct {
	reader *bufio.Reader
}

// NewUI creates a new UI
func NewUI() *UI {
	return &UI{
		reader: bufio.NewReader(os.Stdin),
	}
}

// DisplayPlanSummary displays a summary of the plan
func (u *UI) DisplayPlanSummary(plan *model.Plan) {
	fmt.Println(colorBold + "Terraform Step Debugger" + colorReset)
	fmt.Println("Plan file:", plan.PlanFile)
	fmt.Println("Directory:", plan.TerraformDir)
	fmt.Println()

	fmt.Println(colorBold + "Plan Summary:" + colorReset)
	fmt.Printf("  %sCreates:%s %d\n", colorGreen, colorReset, plan.Stats.Create)
	fmt.Printf("  %sUpdates:%s %d\n", colorYellow, colorReset, plan.Stats.Update)
	fmt.Printf("  %sDeletes:%s %d\n", colorRed, colorReset, plan.Stats.Delete)
	fmt.Printf("  %sNoops:%s %d\n", colorBlue, colorReset, plan.Stats.Noop)
	fmt.Println()
}

// DisplayResourceInfo displays information about a resource
func (u *UI) DisplayResourceInfo(resource *model.Resource, index, total int) {
	// Calculate the percentage complete
	percent := float64(index) / float64(total) * 100

	// Display the progress
	fmt.Printf("[%d/%d] %.1f%% complete\n", index, total, percent)

	// Display the resource information with appropriate color
	var color string
	switch resource.Action {
	case model.ActionCreate:
		color = colorGreen
	case model.ActionUpdate:
		color = colorYellow
	case model.ActionDelete:
		color = colorRed
	default:
		color = colorReset
	}

	// Print the resource information
	fmt.Printf("\n%sResource: %s%s\n", colorBold, resource.Address, colorReset)
	fmt.Printf("  %sAction:%s %s%s%s\n", colorBold, colorReset, color, resource.Action, colorReset)
	fmt.Printf("  %sType:%s %s\n", colorBold, colorReset, resource.Type)

	// Display dependencies if any
	if len(resource.Dependencies) > 0 {
		fmt.Printf("  %sDependencies:%s\n", colorBold, colorReset)
		for _, dep := range resource.Dependencies {
			fmt.Printf("    - %s\n", dep)
		}
	}

	// Display warnings if any
	if len(resource.Warnings) > 0 {
		fmt.Printf("  %sWarnings:%s\n", colorBold, colorReset)
		for _, warning := range resource.Warnings {
			fmt.Printf("    - %s%s%s\n", colorYellow, warning, colorReset)
		}
	}

	fmt.Println()
}

// GetUserAction gets the action to take for the current step
func (u *UI) GetUserAction() (model.StepAction, error) {
	for {
		fmt.Print(colorBold + "Action" + colorReset + " [a=apply, s=skip, d=detail, x=abort]: ")
		input, err := u.reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "a", "apply":
			return model.StepApply, nil
		case "s", "skip":
			return model.StepSkip, nil
		case "d", "detail":
			return model.StepDetail, nil
		case "x", "abort":
			return model.StepAbort, nil
		default:
			fmt.Println("Invalid action. Please try again.")
		}
	}
}

// DisplayExecutionResult displays the result of executing a resource
func (u *UI) DisplayExecutionResult(resource *model.Resource, success bool, elapsed time.Duration) {
	if success {
		fmt.Printf("%sSuccess:%s Applied %s in %.2f seconds\n\n",
			colorGreen, colorReset, resource.Address, elapsed.Seconds())
	} else {
		fmt.Printf("%sFailure:%s Could not apply %s (%.2f seconds)\n\n",
			colorRed, colorReset, resource.Address, elapsed.Seconds())
	}
}

// DisplaySummary displays a summary of the execution
func (u *UI) DisplaySummary(executedResources []*model.Resource) {
	fmt.Println(colorBold + "Execution Summary:" + colorReset)

	// Count resources by status
	completed := 0
	skipped := 0
	failed := 0

	for _, res := range executedResources {
		switch res.Status {
		case model.StatusComplete:
			completed++
		case model.StatusSkipped:
			skipped++
		case model.StatusFailed:
			failed++
		}
	}

	// Display the counts
	fmt.Printf("  %sCompleted:%s %d\n", colorGreen, colorReset, completed)
	fmt.Printf("  %sSkipped:%s %d\n", colorYellow, colorReset, skipped)
	fmt.Printf("  %sFailed:%s %d\n", colorRed, colorReset, failed)

	// Display detailed resource status
	fmt.Println("\nResource Status:")
	for _, res := range executedResources {
		var statusColor string
		switch res.Status {
		case model.StatusComplete:
			statusColor = colorGreen
		case model.StatusSkipped:
			statusColor = colorYellow
		case model.StatusFailed:
			statusColor = colorRed
		default:
			statusColor = colorReset
		}

		fmt.Printf("  %s: %s%s%s\n", res.Address, statusColor, res.Status, colorReset)
	}

	fmt.Println()
}

// ConfirmContinue asks the user if they want to continue after an error
func (u *UI) ConfirmContinue() bool {
	fmt.Print(colorBold + "Continue" + colorReset + " despite errors? [y/n]: ")
	input, err := u.reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// WaitForEnter waits for the user to press Enter
func (u *UI) WaitForEnter() {
	fmt.Print("Press Enter to continue...")
	_, _ = u.reader.ReadString('\n')
}
