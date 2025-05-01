# Terraform Local Demo

This is a demo Terraform configuration for testing the `terraform-step-debug` tool locally without requiring any cloud provider access. It uses local resources and random generators to create a realistic dependency graph of resources.

## What's Included

1. Random resources (pet names, UUIDs, passwords, integers)
2. Local file resources with templated content
3. Time-based resources to simulate longer-running operations
4. Various dependencies between resources
5. Output values with resource information

## Resource Dependency Graph

This configuration creates the following dependency structure:

```
random_pet.server
↓
random_password.db_password, random_password.api_key (independent)
↓
random_uuid.bucket_suffix, random_uuid.api_token (independent)
↓
random_integer.priority, random_integer.port (independent)
↓
local_file.output_directory (creates output directory)
↓
local_file.config, local_file.infrastructure, local_file.secrets (file outputs)
↓
time_sleep.wait_30_seconds (waits 30 seconds)
↓
local_file.delayed_report (created after delay)
```

## Usage

1. Initialize the Terraform configuration:
   ```
   terraform init
   ```

2. Run your terraform-step-debug tool:
   ```
   terraform-step-debug
   ```

3. Follow the prompts to apply each resource step by step.

4. To clean up when you're done:
   ```
   terraform destroy
   ```

## Expected Outputs

After applying this configuration, you'll find the following files in the `output` directory:

- `app-config.json`: A simulated application configuration in JSON format
- `infrastructure.txt`: A simulated infrastructure report  
- `secrets.txt`: A simulated secrets file
- `delayed-report.txt`: A file that is created after a 30-second delay

## Notes

- The `time_sleep` resource introduces a 30-second delay to help demonstrate the step-by-step debugging capability of your tool.
- All files are created locally in the `output` directory.
- No connections to external services or cloud providers are required.