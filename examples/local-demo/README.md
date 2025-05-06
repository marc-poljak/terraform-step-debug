# 🧪 Terraform Local Demo

This is a demo Terraform configuration for testing the `terraform-step-debug` tool locally without requiring any cloud provider access. It uses local resources and random generators to create a realistic dependency graph of resources.

## ✨ What's Included

1. 🎲 Random resources (pet names, UUIDs, passwords, integers)
2. 📄 Local file resources with templated content
3. ⏱️ Time-based resources to simulate longer-running operations
4. 🔄 Various dependencies between resources
5. 📊 Output values with resource information
6. 🔧 Variable files to demonstrate different configurations

## 📊 Resource Dependency Graph

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
local_file.config, local_file.infrastructure, local_file.secrets, local_file.environment_info (file outputs)
↓
time_sleep.wait_for_delay (configurable, may be skipped)
↓
local_file.delayed_report_with_delay or local_file.delayed_report_no_delay (depending on enable_delay)
```

## 🚀 Usage

1. Initialize the Terraform configuration:
   ```
   terraform init
   ```

2. Run your terraform-step-debug tool with default variables:
   ```
   terraform-step-debug
   ```

3. Or run with a specific variable file:
   ```
   terraform-step-debug --var-file=demo.tfvars
   ```
   
   ```
   terraform-step-debug --var-file=prod.tfvars
   ```

4. Follow the prompts to apply each resource step by step.

5. To clean up when you're done:
   ```
   terraform destroy
   ```

## 🌐 Variable Files

This demo includes multiple variable files to demonstrate how terraform-step-debug handles different configurations:

- **🔧 Default**: Uses the default values in variables.tf
  - app_name_prefix: "app"
  - environment: "dev"
  - 30-second delay enabled

- **🚧 demo.tfvars**: Staging environment settings
  - app_name_prefix: "demo"
  - environment: "staging"
  - 10-second delay enabled
  - Custom tags

- **🏭 prod.tfvars**: Production environment settings
  - app_name_prefix: "prod"
  - environment: "production"
  - Delay disabled
  - Higher security settings (longer passwords)
  - Additional tags

## 🔄 Idempotent Configuration

This demo is designed to be idempotent, meaning that after the initial apply, subsequent runs of `terraform apply` or `terraform-step-debug` will show "No changes. Your infrastructure matches the configuration." This is achieved by:

1. Avoiding dynamic content like timestamps in file resources
2. Using deterministic resource names and values

This behavior demonstrates a best practice in Terraform configurations and makes the demo more predictable and stable while testing the step debugger.

## 📂 Expected Outputs

After applying this configuration, you'll find the following files in the `output` directory:

- `app-config.json`: A simulated application configuration in JSON format
- `infrastructure.txt`: A simulated infrastructure report  
- `secrets.txt`: A simulated secrets file
- `environment-[env].txt`: Information about the environment used
- `delayed-report.txt`: A file that is created after a delay (if enabled)

## 📝 Notes

- The `time_sleep` resource introduces a configurable delay to help demonstrate the step-by-step debugging capability of your tool.
- All files are created locally in the `output` directory.
- No connections to external services or cloud providers are required.
- Try running with different var-files to see how the tool handles different configurations.