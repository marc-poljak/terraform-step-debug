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
}

# Create random passwords with different lengths
resource "random_password" "db_password" {
  length  = 16
  special = true
}

resource "random_password" "api_key" {
  length  = 24
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
  min = 8000
  max = 9000
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

# This resource intentionally waits before creating to simulate a longer-running resource
resource "time_sleep" "wait_30_seconds" {
  depends_on = [random_pet.server]

  create_duration = "30s"
}

resource "local_file" "delayed_report" {
  depends_on = [time_sleep.wait_30_seconds]

  content         = "This file was created after a 30-second delay to simulate a longer-running resource. Timestamp: ${timestamp()}"
  filename        = "${path.module}/output/delayed-report.txt"
  file_permission = "0600"
}

# Outputs
output "app_name" {
  value = random_pet.server.id
}

output "bucket_name" {
  value = "data-${random_uuid.bucket_suffix.result}"
}

output "api_port" {
  value = random_integer.port.result
}

output "files_created" {
  value = [
    local_file.config.filename,
    local_file.infrastructure.filename,
    local_file.secrets.filename,
    local_file.delayed_report.filename
  ]
}