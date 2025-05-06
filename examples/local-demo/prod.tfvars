app_name_prefix = "prod"
environment     = "production"
min_port        = 10000
max_port        = 10999
password_length = 32
enable_delay    = false
delay_seconds   = 0
resource_tags = {
  Project     = "Terraform Step Debug Demo"
  ManagedBy   = "Terraform"
  Environment = "Production"
  Owner       = "Production Team"
  Purpose     = "Variable File Demonstration"
  Sensitivity = "High"
  Region      = "eu-central-1"
  Backup      = "Daily"
  Compliance  = "ISO27001"
}