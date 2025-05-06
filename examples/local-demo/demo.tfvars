app_name_prefix = "demo"
environment     = "staging"
min_port        = 5000
max_port        = 5999
password_length = 20
enable_delay    = true
delay_seconds   = 10
resource_tags = {
  Project     = "Terraform Step Debug Demo"
  ManagedBy   = "Terraform"
  Environment = "Staging"
  Owner       = "DevOps Team"
  Purpose     = "Variable File Demonstration"
  Region      = "eu-west-1"
}