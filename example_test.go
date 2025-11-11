//go:build cgo

package fipscheck

import (
	"context"
	"testing"
	"time"
)

// TestExampleUsage demonstrates how external users would use the SDK
func TestExampleUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping image tests in short mode")
	}

	t.Run("example_fips_compliant_image", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Test FIPS compliant image
		result, err := CheckImageFIPSCompliance(ctx, "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1")
		if err != nil {
			t.Fatalf("CheckImageFIPSCompliance failed: %v", err)
		}

		t.Logf("Example FIPS Compliant Image Check:")
		t.Logf("  Image: %s", result.ImageRef)
		t.Logf("  FIPS Compliant: %t", result.IsCompliant)
		t.Logf("  Reason: %s", result.Reason)
		t.Logf("  Binaries Found: %d", result.BinariesFound)
		t.Logf("  Compliant Binaries: %d", result.CompliantBinaries)

		if !result.IsCompliant {
			t.Error("Expected FIPS-enabled image to be compliant")
		}

		// Show detailed report for compliant case too
		t.Logf("Detailed compliance analysis:")
		t.Logf("%s", result.DetailedReport)
	})

	t.Run("example_non_fips_image", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Test non-FIPS image
		result, err := CheckImageFIPSCompliance(ctx, "mcr.microsoft.com/oss/kubernetes-csi/blob-csi:v1.26.6")
		if err != nil {
			t.Fatalf("CheckImageFIPSCompliance failed: %v", err)
		}

		t.Logf("Example Non-FIPS Image Check:")
		t.Logf("  Image: %s", result.ImageRef)
		t.Logf("  FIPS Compliant: %t", result.IsCompliant)
		t.Logf("  Reason: %s", result.Reason)
		t.Logf("  Binaries Found: %d", result.BinariesFound)
		t.Logf("  Compliant Binaries: %d", result.CompliantBinaries)

		if result.IsCompliant {
			t.Error("Expected non-FIPS image to be non-compliant")
		}

		// Show detailed report for non-compliant case
		t.Logf("Detailed compliance analysis:")
		t.Logf("%s", result.DetailedReport)
	})
}
