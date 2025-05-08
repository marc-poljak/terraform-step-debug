#!/bin/zsh

# Get the version from git tag
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
echo "Creating release $VERSION..."

# Build the release binaries
make dist

# Create a GitHub release using gh CLI tool
gh release create $VERSION \
  --title "Terraform Step Debug $VERSION" \
  --notes "Release of Terraform Step Debug tool, enabling step-by-step debugging of Terraform operations. Now with RISC-V support!" \
  ./build/dist/terraform-step-debug_darwin_amd64.gz \
  ./build/dist/terraform-step-debug_darwin_arm64.gz \
  ./build/dist/terraform-step-debug_linux_amd64.gz \
  ./build/dist/terraform-step-debug_linux_arm64.gz \
  ./build/dist/terraform-step-debug_linux_riscv64.gz \
  ./build/dist/terraform-step-debug_windows_amd64.zip