//go:build cgo

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bahe-msft/fips-check"
)

func main() {
	ctx := context.Background()

	// Test cases: one FIPS-compliant and one non-compliant image
	testImages := []string{
		"mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1",         // Should be compliant
		"mcr.microsoft.com/oss/kubernetes-csi/blob-csi:v1.26.6",         // Should be non-compliant
	}

	for i, imageRef := range testImages {
		fmt.Printf("=== Test Case %d ===\n", i+1)
		fmt.Printf("Checking FIPS compliance for: %s\n\n", imageRef)

		result, err := fipscheck.CheckImageFIPSCompliance(ctx, imageRef)
		if err != nil {
			log.Fatalf("Error checking image: %v", err)
		}

		// Show summary
		fmt.Printf("=== Summary ===\n")
		fmt.Printf("Image: %s\n", result.ImageRef)
		fmt.Printf("FIPS Compliant: %t\n", result.IsCompliant)
		fmt.Printf("Reason: %s\n", result.Reason)
		fmt.Printf("Binaries Found: %d\n", result.BinariesFound)
		fmt.Printf("Compliant Binaries: %d\n\n", result.CompliantBinaries)

		// Show detailed report
		fmt.Printf("=== Detailed Report ===\n")
		fmt.Printf("%s\n", result.DetailedReport)

		fmt.Printf("\n" + strings.Repeat("=", 80) + "\n\n")
	}
}