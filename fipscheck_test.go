//go:build cgo

package fipscheck

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCheckImageFIPSCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping image test in short mode")
	}

	tests := []struct {
		name            string
		imageRef        string
		expectCompliant bool
		description     string
	}{
		{
			name:            "fips_compliant_istio_image",
			imageRef:        "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1",
			expectCompliant: true,
			description:     "FIPS-enabled Istio proxy image",
		},
		{
			name:            "non_fips_blob_csi_image",
			imageRef:        "mcr.microsoft.com/oss/kubernetes-csi/blob-csi:v1.26.6",
			expectCompliant: false,
			description:     "Non-FIPS blob CSI image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			result, err := CheckImageFIPSCompliance(ctx, tt.imageRef)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.IsCompliant != tt.expectCompliant {
				t.Errorf("Expected IsCompliant %t, got %t. Reason: %s",
					tt.expectCompliant, result.IsCompliant, result.Reason)
			}

			t.Logf("=== %s ===", tt.description)
			t.Logf("Image: %s", result.ImageRef)
			t.Logf("FIPS Compliant: %t", result.IsCompliant)
			t.Logf("Reason: %s", result.Reason)
			t.Logf("Binaries Found: %d", result.BinariesFound)
			t.Logf("Compliant Binaries: %d", result.CompliantBinaries)

			if !strings.Contains(result.DetailedReport, "=== Host FIPS Environment Check ===") {
				t.Error("DetailedReport missing host FIPS section")
			}
		})
	}
}