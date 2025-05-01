# Terraform Step Debugger

A CLI tool that intercepts Terraform apply operations and executes them step-by-step with user approval for each resource operation, similar to a code debugger.

## Features

- Generate a Terraform plan
- Parse the plan to identify individual resource operations
- Present each operation to the user for review/approval
- Execute operations one-by-one using targeted apply
- Allow stepping, skipping, or aborting the process
- Dependency-aware execution order

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/terraform-step-debug.git
cd terraform-step-debug

# Build and install
make install
```

### Using Go

```bash
go install github.com/yourusername/terraform-step-debug/cmd/terraform-step-debug@latest
```

## Usage

```bash
# Basic usage (in a Terraform directory)
terraform-step-debug

# Specify a Terraform directory
terraform-step-debug --dir /path/to/terraform/project

# Use an existing plan file
terraform-step-debug --plan /path/to/terraform.tfplan

# Dry run mode (don't apply changes)
terraform-step-debug --dry-run

# Target a specific resource
terraform-step-debug --target aws_instance.example
```

### Commands During Execution

During the step-by-step execution, you can use the following commands:

- `a` or `apply` - Apply the current resource
- `s` or `skip` - Skip the current resource
- `d` or `detail` - Show detailed information about the current resource
- `x` or `abort` - Abort the execution

## Requirements

- Go 1.18 or higher
- Terraform 0.12 or higher

## Development

```bash
# Clone the repository
git clone https://github.com/yourusername/terraform-step-debug.git
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