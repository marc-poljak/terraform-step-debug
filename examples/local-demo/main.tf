terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5.1"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9.1"
    }
  }
  required_version = ">= 1.0.0"
}

# Create a random pet name for our demo
resource "random_pet" "server" {
  length    = 2
  separator = "-"
  prefix    = var.app_name_prefix
}

# Create random passwords with different lengths
resource "random_password" "db_password" {
  length  = var.password_length
  special = true
}

resource "random_password" "api_key" {
  length  = var.password_length + 8
  special = true
}

# Create UUIDs for various resources
resource "random_uuid" "bucket_suffix" {}

resource "random_uuid" "api_token" {}

# Create some random integers to simulate IDs
resource "random_integer" "priority" {
  min = 1
  max = 100
}

resource "random_integer" "port" {
  min = var.min_port
  max = var.max_port
}

# Create a local file with our "configuration"
resource "local_file" "config" {
  content = templatefile("${path.module}/templates/config.tpl", {
    app_name    = random_pet.server.id
    db_password = random_password.db_password.result
    api_key     = random_password.api_key.result
    bucket_name = "data-${random_uuid.bucket_suffix.result}"
    api_token   = random_uuid.api_token.result
    priority    = random_integer.priority.result
    port        = random_integer.port.result
    environment = var.environment
    tags        = var.resource_tags
  })
  filename        = "${path.module}/output/app-config.json"
  file_permission = "0600"

  depends_on = [local_file.output_directory]
}

# Create a file with "infrastructure details"
resource "local_file" "infrastructure" {
  content = templatefile("${path.module}/templates/infrastructure.tpl", {
    app_name    = random_pet.server.id
    bucket_name = "data-${random_uuid.bucket_suffix.result}"
    port        = random_integer.port.result
    environment = var.environment
    tags        = var.resource_tags
  })
  filename        = "${path.module}/output/infrastructure.txt"
  file_permission = "0600"

  depends_on = [local_file.output_directory]
}

# Create a file with "secrets"
resource "local_file" "secrets" {
  content = templatefile("${path.module}/templates/secrets.tpl", {
    db_password = random_password.db_password.result
    api_key     = random_password.api_key.result
    api_token   = random_uuid.api_token.result
    environment = var.environment
  })
  filename        = "${path.module}/output/secrets.txt"
  file_permission = "0600"

  depends_on = [local_file.output_directory]
}

# Create output directory
resource "local_file" "output_directory" {
  content         = ""
  filename        = "${path.module}/output/.gitkeep"
  file_permission = "0600"

  provisioner "local-exec" {
    command = "mkdir -p ${path.module}/output"
  }
}

# Create a file with environment info
resource "local_file" "environment_info" {
  content = templatefile("${path.module}/templates/environment.tpl", {
    environment = var.environment
    app_name    = random_pet.server.id
    tags        = var.resource_tags
  })
  filename        = "${path.module}/output/environment-${var.environment}.txt"
  file_permission = "0600"

  depends_on = [local_file.output_directory]
}

# This resource intentionally waits before creating to simulate a longer-running resource
# Only created if enable_delay is true
resource "time_sleep" "wait_for_delay" {
  count      = var.enable_delay ? 1 : 0
  depends_on = [random_pet.server]

  create_duration = "${var.delay_seconds}s"
}

# Delayed report for when delay is enabled
resource "local_file" "delayed_report_with_delay" {
  count      = var.enable_delay ? 1 : 0
  depends_on = [time_sleep.wait_for_delay]

  content = templatefile("${path.module}/templates/delayed.tpl", {
    environment = var.environment
    delay       = var.delay_seconds
    app_name    = random_pet.server.id
  })
  filename        = "${path.module}/output/delayed-report.txt"
  file_permission = "0600"
}

# Immediate report for when delay is disabled
resource "local_file" "delayed_report_no_delay" {
  count      = var.enable_delay ? 0 : 1
  depends_on = [local_file.output_directory]

  content = templatefile("${path.module}/templates/delayed.tpl", {
    environment = var.environment
    delay       = 0
    app_name    = random_pet.server.id
  })
  filename        = "${path.module}/output/delayed-report.txt"
  file_permission = "0600"
}

# Outputs
output "app_name" {
  value = random_pet.server.id
}

output "environment" {
  value = var.environment
}

output "bucket_name" {
  value = "data-${random_uuid.bucket_suffix.result}"
}

output "api_port" {
  value = random_integer.port.result
}

output "delay_enabled" {
  value = var.enable_delay
}

output "files_created" {
  value = [
    local_file.config.filename,
    local_file.infrastructure.filename,
    local_file.secrets.filename,
    local_file.environment_info.filename,
    var.enable_delay ? local_file.delayed_report_with_delay[0].filename : local_file.delayed_report_no_delay[0].filename
  ]
}

output "resource_tags" {
  value = var.resource_tags
}