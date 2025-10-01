//go:build cgo

package main

import (
	"fmt"

	"github.com/golang-fips/openssl/v2"

	_ "github.com/bahe-msft/fips-check/internal/opensslsetup"
)

func main() {
	checkHost()
}

func checkHost() {
	fmt.Printf("Host:\n")
	fmt.Printf("- OpenSSL version: %s\n", openssl.VersionText())
	fmt.Printf("- OpenSSL FIPS capable: %t\n", openssl.FIPSCapable())
}
