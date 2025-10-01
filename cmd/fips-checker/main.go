//go:build cgo

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/golang-fips/openssl/v2"

	"github.com/bahe-msft/fips-check/internal/binarychecker"
	_ "github.com/bahe-msft/fips-check/internal/opensslsetup"
)

func main() {
	checkHost()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	reports, err := binarychecker.Check(ctx, "/")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printReports(reports)
}

func printReports(reports []binarychecker.BinaryReport) {
	fmt.Printf("\n=== Binary FIPS Check Report ===\n")
	fmt.Printf("Total binaries scanned: %d\n\n", len(reports))

	if len(reports) == 0 {
		fmt.Println("No Go binaries found.")
		return
	}

	// Count statistics
	systemcryptoCount := 0
	failedCount := 0
	for _, report := range reports {
		if report.GoBinaryDetails.UseSystemcrypto {
			systemcryptoCount++
		}
		if report.GoBinaryDetails.FailsOnFIPSCheck {
			failedCount++
		}
	}

	fmt.Printf("Binaries with systemcrypto: %d\n", systemcryptoCount)
	fmt.Printf("Binaries that fail FIPS check: %d\n\n", failedCount)

	// Print detailed report for each binary
	for i, report := range reports {
		fmt.Printf("─────────────────────────────────────────────────────\n")
		fmt.Printf("[%d] Binary: %s\n", i+1, report.RelativePath)
		fmt.Printf("    Type: %s\n", report.Type)

		details := report.GoBinaryDetails
		fmt.Printf("    Go Version: %s\n", details.GoVersion)
		if details.Module != "" {
			fmt.Printf("    Module: %s\n", details.Module)
		}
		fmt.Printf("    CGO Enabled: %t\n", details.CGOEnabled)
		fmt.Printf("    Uses Systemcrypto: %t\n", details.UseSystemcrypto)

		// Always report FIPS status
		if !details.UseSystemcrypto {
			fmt.Printf("    Fails on FIPS Check: %t\n", details.FailsOnFIPSCheck)
			fmt.Printf("    ⚠️  FIPS Status: NOT FIPS ENABLED (no systemcrypto)\n")
			printRuntimeOutput(details.RuntimePanicLog)
		} else {
			fmt.Printf("    Fails on FIPS Check: %t\n", details.FailsOnFIPSCheck)

			if details.FailsOnFIPSCheck {
				fmt.Printf("    ❌ FIPS Status: NOT COMPLIANT\n")
			} else {
				fmt.Printf("    ✅ FIPS Status: MIGHT BE COMPLIANT\n")
			}

			printRuntimeOutput(details.RuntimePanicLog)
		}

		if report.Error != nil {
			fmt.Printf("    ⚠️  Error: %v\n", report.Error)
		}

		fmt.Println()
	}

	fmt.Printf("─────────────────────────────────────────────────────\n")
	fmt.Printf("Summary:\n")
	fmt.Printf("  Total: %d | Systemcrypto: %d | Failed FIPS: %d\n",
		len(reports), systemcryptoCount, failedCount)
}

// printRuntimeOutput prints the runtime panic log with indentation
func printRuntimeOutput(log string) {
	if log != "" {
		fmt.Printf("    Runtime Output:\n")
		lines := strings.SplitSeq(log, "\n")
		for line := range lines {
			if line != "" {
				fmt.Printf("        %s\n", line)
			}
		}
	}
}

func checkHost() {
	fmt.Printf("Host:\n")
	fmt.Printf("- OpenSSL version: %s\n", openssl.VersionText())
	fmt.Printf("- OpenSSL FIPS capable: %t\n", openssl.FIPSCapable())
}
