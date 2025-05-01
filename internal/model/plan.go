package model

// Resource represents a single Terraform resource operation
// (create, update, delete) from the plan
type Resource struct {
	Address      string         // The resource address (e.g., aws_instance.example)
	Type         string         // Resource type (e.g., aws_instance)
	Name         string         // Resource name (e.g., example)
	Action       Action         // The action (create, update, delete)
	Dependencies []string       // List of resource addresses this resource depends on
	Attributes   map[string]any // The resource attributes
	Status       ResourceStatus // Current status of the resource during execution
	Warnings     []string       // Any warnings associated with this resource
}

// Action represents the type of operation to be performed on a resource
type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionRead   Action = "read"
	ActionNoop   Action = "no-op"
)

// ResourceStatus represents the current status of a resource in the execution process
type ResourceStatus string

const (
	StatusPending  ResourceStatus = "pending"
	StatusApproved ResourceStatus = "approved"
	StatusSkipped  ResourceStatus = "skipped"
	StatusFailed   ResourceStatus = "failed"
	StatusComplete ResourceStatus = "complete"
)

// Plan represents a parsed Terraform plan
type Plan struct {
	Resources    []*Resource          // All resources in the plan
	ResourcesMap map[string]*Resource // Resources mapped by address for quick lookup
	PlanFile     string               // Path to the Terraform plan file
	TerraformDir string               // Path to the Terraform directory
	HasChanges   bool                 // Whether the plan has any changes
	Stats        PlanStats            // Statistics about the plan
}

// PlanStats contains statistics about a plan
type PlanStats struct {
	Create int // Number of resources to create
	Update int // Number of resources to update
	Delete int // Number of resources to delete
	Noop   int // Number of resources with no changes
}

// NewPlan creates a new empty Plan
func NewPlan(planFile, terraformDir string) *Plan {
	return &Plan{
		Resources:    make([]*Resource, 0),
		ResourcesMap: make(map[string]*Resource),
		PlanFile:     planFile,
		TerraformDir: terraformDir,
		HasChanges:   false,
		Stats:        PlanStats{},
	}
}

// ExecutionGraph represents the ordered list of resources to be executed
// based on their dependencies, grouped by layers that can be executed in parallel
type ExecutionGraph struct {
	Layers [][]*Resource // Resources grouped by dependency layers
}

// StepAction represents the action to take for the current step
type StepAction string

const (
	StepApply  StepAction = "apply"  // Apply the current resource
	StepSkip   StepAction = "skip"   // Skip the current resource
	StepAbort  StepAction = "abort"  // Abort the entire process
	StepDetail StepAction = "detail" // Show more details about the current resource
)
