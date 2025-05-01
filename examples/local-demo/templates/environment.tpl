ENVIRONMENT: ${environment}
=======================

Application: ${app_name}

Resource Tags:
%{ for key, value in tags ~}
  - ${key}: ${value}
%{ endfor ~}

This file demonstrates using different environment configurations
through variable files. When you run terraform-step-debug with
different var-files, this file will change.

For instance:
  terraform-step-debug --var-file=demo.tfvars
  terraform-step-debug --var-file=prod.tfvars

Timestamp: ${timestamp()}