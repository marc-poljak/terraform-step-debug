# Terraform Step Debugger

A CLI tool that intercepts Terraform apply operations and executes them step-by-step with user approval for each resource operation, similar to a code debugger.

## Features

- Generate a Terraform plan
- Parse the plan to identify individual resource operations
- Present each operation to the user for review/approval
- Execute operations one-by-one using targeted apply
- Allow stepping, skipping, or aborting the process
- Dependency-aware execution order
- Support for variable files (tfvars)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/marc-poljak/terraform-step-debug.git
cd terraform-step-debug

# Build and install
make install
```

### Using Go

```bash
go install github.com/marc-poljak/terraform-step-debug/cmd/terraform-step-debug@latest
```

## Usage

```bash
# Basic usage (in a Terraform directory)
terraform-step-debug

# Specify a Terraform directory
terraform-step-debug --dir /path/to/terraform/project

# Use an existing plan file
terraform-step-debug --plan /path/to/terraform.tfplan

# Use with a variable file (e.g., prod.tfvars)
terraform-step-debug --var-file prod.tfvars

# Dry run mode (don't apply changes)
terraform-step-debug --dry-run

# Target a specific resource
terraform-step-debug --target aws_instance.example
```

### Environment-Specific Deployments

For different environments, you can use variable files:

```bash
# Development environment
terraform-step-debug

# Staging environment
terraform-step-debug --var-file staging.tfvars

# Production environment 
terraform-step-debug --var-file prod.tfvars
```

This ensures that all operations use the correct variable values for each environment, maintaining consistency between planning and execution.

### Commands During Execution

During the step-by-step execution, you can use the following commands:

- `a` or `apply` - Apply the current resource
- `s` or `skip` - Skip the current resource
- `d` or `detail` - Show detailed information about the current resource
- `x` or `abort` - Abort the execution

## Example

The repository includes a local demo in `examples/local-demo` that you can use to try the tool without requiring any cloud provider access:

```bash
cd examples/local-demo
terraform init
terraform-step-debug --var-file prod.tfvars
```

The demo includes different variable files for development, staging, and production environments, showing how the tool can be used with environment-specific configurations.

## Requirements

- Go 1.21 or higher
- Terraform 0.12 or higher

## Development

```bash
# Clone the repository
git clone https://github.com/marc-poljak/terraform-step-debug.git
cd terraform-step-debug

# Run tests
make test

# Run linter
make lint

# Build the binary
make build

# Build and run
make run
```

## Project Structure

```
terraform-step-debug/
├── cmd/
│   └── terraform-step-debug/    # Main command entrypoint
├── internal/
│   ├── executor/                # Apply step execution
│   ├── parser/                  # Terraform plan parsing
│   ├── model/                   # Data structures
│   ├── ui/                      # Interactive UI components
│   └── util/                    # Helper functions
├── examples/
│   └── local-demo/              # Local demo with variable files
├── build/                       # Build artifacts
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
└── README.md
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.