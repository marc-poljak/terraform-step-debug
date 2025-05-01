package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// FindTerraformBinary tries to find the terraform binary in PATH
func FindTerraformBinary() (string, error) {
	// Try to find terraform in PATH
	binary, err := exec.LookPath("terraform")
	if err == nil {
		return binary, nil
	}

	// For macOS with Homebrew, check common locations
	homebrewPaths := []string{
		"/usr/local/bin/terraform",
		"/opt/homebrew/bin/terraform",
	}

	for _, path := range homebrewPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("terraform binary not found in PATH or common locations")
}

// FindTerraformDir finds the directory containing Terraform files
func FindTerraformDir(startDir string) (string, error) {
	// If startDir is empty, use the current directory
	if startDir == "" {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Check if the directory contains Terraform files
	files, err := os.ReadDir(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	// Look for .tf files
	hasTfFiles := false
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tf") {
			hasTfFiles = true
			break
		}
	}

	if hasTfFiles {
		return startDir, nil
	}

	// If we didn't find any .tf files and we're not at the root,
	// try the parent directory
	parentDir := filepath.Dir(startDir)
	if parentDir != startDir {
		return FindTerraformDir(parentDir)
	}

	return "", fmt.Errorf("no Terraform files found in the directory hierarchy")
}

// CreateTempPlanFile creates a temporary plan file
func CreateTempPlanFile() (string, error) {
	// Create a temporary file for the plan
	tmpFile, err := os.CreateTemp("", "terraform-step-debug-*.tfplan")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary plan file: %w", err)
	}

	// Close the file and return its path
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temporary plan file: %w", err)
	}

	return tmpFile.Name(), nil
}

// CheckTerraformVersion checks if the Terraform version is compatible
func CheckTerraformVersion(terraformPath string) error {
	cmd := exec.Command(terraformPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get Terraform version: %w", err)
	}

	versionStr := string(output)

	// Check if the version is supported
	if !strings.Contains(versionStr, "Terraform v") {
		return fmt.Errorf("unexpected Terraform version output: %s", versionStr)
	}

	// Extract version number
	versionMatch := strings.SplitN(versionStr, "v", 2)
	if len(versionMatch) < 2 {
		return fmt.Errorf("unable to parse Terraform version: %s", versionStr)
	}

	versionNumber := strings.Split(versionMatch[1], " ")[0]

	// Check for minimum version requirement (0.12.0)
	parts := strings.Split(versionNumber, ".")
	if len(parts) < 2 {
		return fmt.Errorf("unable to parse Terraform version number: %s", versionNumber)
	}

	majorVersion, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("unable to parse Terraform major version: %w", err)
	}

	minorVersion, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("unable to parse Terraform minor version: %w", err)
	}

	// Terraform 0.12.0 or higher is required
	if majorVersion == 0 && minorVersion < 12 {
		return fmt.Errorf("unsupported Terraform version %s, version 0.12.0 or higher is required", versionNumber)
	}

	// Informative message for newer versions
	if majorVersion >= 1 && minorVersion >= 11 {
		fmt.Printf("Using Terraform v%s\n", versionNumber)
	}

	return nil
}

// CleanupFiles removes temporary files
func CleanupFiles(files ...string) {
	for _, file := range files {
		_ = os.Remove(file)
	}
}

// ValidateTargetResource validates if a target resource exists in the plan
func ValidateTargetResource(target string, resourceMap map[string]*struct{}) error {
	if target == "" {
		return nil // No target specified, validation passes
	}

	// Create a new map to track valid targets
	validTargets := make(map[string]bool)

	// Populate valid targets from resourceMap
	for addr := range resourceMap {
		validTargets[addr] = true
	}

	// Check if the target is valid
	if !validTargets[target] {
		return fmt.Errorf("target resource '%s' not found in plan", target)
	}

	return nil
}

// CalculateResourceCount counts resources by action type
func CalculateResourceCount(resourceMap map[string]string) map[string]int {
	counts := make(map[string]int)

	for _, action := range resourceMap {
		counts[action]++
	}

	return counts
}

// FormatAction returns a formatted string representation of an action
func FormatAction(action string) string {
	switch action {
	case "create":
		return "Create"
	case "update":
		return "Update"
	case "delete":
		return "Delete"
	case "read":
		return "Read"
	case "no-op":
		return "No-op"
	default:
		return action
	}
}

// FormatAddress formats a resource address for display
func FormatAddress(address string) string {
	// Check if it's a data source
	if strings.HasPrefix(address, "data.") {
		parts := strings.SplitN(address[5:], ".", 2)
		if len(parts) == 2 {
			return fmt.Sprintf("data \"%s\" \"%s\"", parts[0], parts[1])
		}
	}

	// Regular resource
	parts := strings.SplitN(address, ".", 2)
	if len(parts) == 2 {
		return fmt.Sprintf("\"%s\" \"%s\"", parts[0], parts[1])
	}

	return address
}
