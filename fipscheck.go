//go:build cgo

// Package fipscheck provides FIPS compliance checking functionality for container images.
package fipscheck

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ImageFIPSResult contains the FIPS compliance result for a container image.
type ImageFIPSResult struct {
	ImageRef          string
	IsCompliant       bool
	Reason           string
	BinariesFound    int
	CompliantBinaries int
	DetailedReport   string // Full detailed report like the CLI output
}

// CheckImageFIPSCompliance checks if a container image is FIPS compliant.
// It accepts an image reference and returns the compliance status with all checks embedded.
func CheckImageFIPSCompliance(ctx context.Context, imageRef string) (*ImageFIPSResult, error) {
	buildImage := "mcr.microsoft.com/oss/go/microsoft/golang:1.24-fips-azurelinux3.0"

	// Generate unique tag for this check
	imageTag := fmt.Sprintf("fips-checker-%d", time.Now().Unix())

	// Build and run the FIPS checker
	buildCmd := exec.CommandContext(ctx, "docker", "build",
		"--build-arg", fmt.Sprintf("BUILD_IMAGE=%s", buildImage),
		"--build-arg", fmt.Sprintf("RUNTIME_IMAGE=%s", imageRef),
		"-t", imageTag,
		".")

	if err := buildCmd.Run(); err != nil {
		return &ImageFIPSResult{
			ImageRef:    imageRef,
			IsCompliant: false,
			Reason:      fmt.Sprintf("Failed to build checker: %v", err),
		}, nil
	}

	// Clean up image after check
	defer func() {
		exec.Command("docker", "rmi", imageTag).Run()
	}()

	// Run the checker and capture output
	runCmd := exec.CommandContext(ctx, "docker", "run", "--rm", imageTag)
	output, err := runCmd.CombinedOutput()
	if err != nil {
		return &ImageFIPSResult{
			ImageRef:    imageRef,
			IsCompliant: false,
			Reason:      fmt.Sprintf("Checker failed: %v", err),
		}, nil
	}

	// Parse the output to determine compliance
	outputStr := string(output)

	// Check for distroless/no OpenSSL case
	if strings.Contains(outputStr, "Runtime image does not contain OpenSSL binary") {
		return &ImageFIPSResult{
			ImageRef:    imageRef,
			IsCompliant: false,
			Reason:      "Image lacks OpenSSL binary (likely distroless)",
		}, nil
	}

	// Count binaries and compliant ones
	binariesFound := 0
	compliantBinaries := 0

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "✅ FIPS Status: COMPLIANT") {
			compliantBinaries++
		} else if strings.Contains(line, "❌ FIPS Status: NOT COMPLIANT") {
			binariesFound++
		}
	}
	binariesFound += compliantBinaries

	isCompliant := binariesFound > 0 && compliantBinaries == binariesFound
	reason := "All binaries are FIPS compliant"
	if !isCompliant {
		if binariesFound == 0 {
			reason = "No Go binaries found"
		} else {
			reason = fmt.Sprintf("%d of %d binaries are not FIPS compliant", binariesFound-compliantBinaries, binariesFound)
		}
	}

	return &ImageFIPSResult{
		ImageRef:          imageRef,
		IsCompliant:       isCompliant,
		Reason:           reason,
		BinariesFound:    binariesFound,
		CompliantBinaries: compliantBinaries,
		DetailedReport:   outputStr,
	}, nil
}