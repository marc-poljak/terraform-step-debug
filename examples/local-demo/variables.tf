variable "app_name_prefix" {
  description = "Prefix to use for the application name (combined with random_pet)"
  type        = string
  default     = "app"
}

variable "environment" {
  description = "Environment name (dev, test, prod)"
  type        = string
  default     = "dev"
}

variable "min_port" {
  description = "Minimum port number for random port selection"
  type        = number
  default     = 8000
}

variable "max_port" {
  description = "Maximum port number for random port selection"
  type        = number
  default     = 9000
}

variable "password_length" {
  description = "Length of generated passwords"
  type        = number
  default     = 16
}

variable "enable_delay" {
  description = "Whether to enable the 30-second delay resource"
  type        = bool
  default     = true
}

variable "delay_seconds" {
  description = "Number of seconds to delay for the time_sleep resource"
  type        = number
  default     = 30
}

variable "resource_tags" {
  description = "Tags to apply to resources (for documentation purposes in this local demo)"
  type        = map(string)
  default = {
    Project     = "Terraform Step Debug Demo"
    ManagedBy   = "Terraform"
    Environment = "Dev"
  }
}