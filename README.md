# FIPS Checker

A tool to verify FIPS compliance of Go binaries in container images.

## What It Does

- Scans container images for Go binaries
- Checks if binaries are built with FIPS-enabled Go (`GOEXPERIMENT=systemcrypto`)
- Tests runtime FIPS compliance by executing binaries with `GOFIPS=1`
- Detects distroless images (which cannot be FIPS compliant)

## Requirements

- Docker
- Bash shell

## Usage

```bash
./build-and-check.sh <runtime-image>
```

**Examples:**
```bash
./build-and-check.sh mcr.microsoft.com/oss/v2/azure/ip-masq-agent-v2:v0.1.15
./build-and-check.sh mcr.microsoft.com/oss/kubernetes-csi/blob-csi:v1.26.6
```

## How It Works

1. **Detects Build Image**: Determines the appropriate FIPS-enabled Go build image
2. **Builds Checker**: Compiles the FIPS checker tool using a FIPS-enabled Go compiler
3. **Scans Runtime Image**: Runs the checker inside the target image to scan all Go binaries
4. **Reports Results**: Shows detailed FIPS compliance status for each binary found

## FIPS Compliance Checking Algorithm

The tool uses a two-phase approach to determine FIPS compliance:

### Phase 1: Static Analysis
Extracts build information from Go binaries using `debug/buildinfo`:

1. **Build Settings Check**: Inspects build settings for:
   - `CGO_ENABLED=1` - Required for OpenSSL integration
   - `GOEXPERIMENT=systemcrypto` - Enables system crypto backend

2. **Systemcrypto Flag**: A binary is marked as using systemcrypto if:
   - Build settings contain `GOEXPERIMENT=systemcrypto`

### Phase 2: Runtime Verification
Tests actual FIPS capability by executing the binary:

1. **Environment Setup**: Runs binary with `GOFIPS=1` environment variable
   - This forces FIPS mode enforcement at runtime

2. **Panic Detection**: Monitors stderr for FIPS-related panic messages:
   ```
   panic: opensslcrypto: FIPS mode requested (system FIPS mode) 
   but not available in OpenSSL X.X.X
   ```

3. **Timeout Handling**: Gives binary 2 seconds to start and potentially panic
   - If timeout occurs without panic: **MIGHT BE COMPLIANT**
   - If FIPS panic detected: **NOT COMPLIANT**
   - If exits normally without FIPS errors: **MIGHT BE COMPLIANT**

### Final Status Determination

| Condition | Status |
|-----------|--------|
| No systemcrypto | ❌ **NOT COMPLIANT (systemcrypto not in use)** |
| Systemcrypto enabled + Runtime check failed | ❌ **NOT COMPLIANT (runtime check fails)** |
| Systemcrypto enabled + Runtime check passed + Host not FIPS capable | ❌ **NOT COMPLIANT (host not FIPS capable)** |
| Systemcrypto enabled + Runtime check passed + Host FIPS capable | ✅ **COMPLIANT** |

**Note**: Full compliance requires systemcrypto enabled, passing runtime checks, and a FIPS-capable OpenSSL on the host system.

## What It Checks

### Static Analysis
- CGO enabled
- `GOEXPERIMENT=systemcrypto` build setting

### Runtime Verification
- Executes each binary with `GOFIPS=1` environment variable
- Detects OpenSSL FIPS mode panic messages
- Confirms OpenSSL FIPS capability on the host system

## Report Output

For each Go binary found:
- **✅ COMPLIANT**: Binary has systemcrypto enabled, passes runtime check, and host is FIPS capable
- **❌ NOT COMPLIANT (systemcrypto not in use)**: Binary doesn't use systemcrypto
- **❌ NOT COMPLIANT (runtime check fails)**: Binary has systemcrypto but fails runtime FIPS check
- **❌ NOT COMPLIANT (host not FIPS capable)**: Binary passes checks but host OpenSSL is not FIPS capable

### Sample Output

**FIPS Compliant Binary:**
```
=== Host FIPS Environment Check ===
OpenSSL Version: OpenSSL 3.0.8 7 Feb 2023
FIPS Capable: true
✅ Status: Host is FIPS capable

=== Binary FIPS Check Report ===
Total binaries scanned: 1

Binaries with systemcrypto: 1
Binaries that fail FIPS check: 0

─────────────────────────────────────────────────────
[1] Binary: usr/local/bin/ip-masq-agent-v2
    Type: gobinary
    Go Version: go1.24.4 X:systemcrypto
    Module: github.com/Azure/ip-masq-agent-v2
    CGO Enabled: true
    Uses Systemcrypto: true
    Fails on FIPS Check: false
    ✅ FIPS Status: COMPLIANT

─────────────────────────────────────────────────────
Summary:
  Total: 1 | Systemcrypto: 1 | Failed FIPS: 0
```

**Non-FIPS Binary:**
```
=== Host FIPS Environment Check ===
OpenSSL Version: OpenSSL 3.0.8 7 Feb 2023
FIPS Capable: true
✅ Status: Host is FIPS capable

=== Binary FIPS Check Report ===
Total binaries scanned: 1

Binaries with systemcrypto: 0
Binaries that fail FIPS check: 1

─────────────────────────────────────────────────────
[1] Binary: usr/local/bin/blobplugin
    Type: gobinary
    Go Version: go1.23.2
    Module: sigs.k8s.io/blob-csi-driver
    CGO Enabled: false
    Uses Systemcrypto: false
    Fails on FIPS Check: true
    ❌ FIPS Status: NOT COMPLIANT (systemcrypto not in use)

─────────────────────────────────────────────────────
Summary:
  Total: 1 | Systemcrypto: 0 | Failed FIPS: 1
```

## Distroless Images

The tool automatically detects distroless images and exits with an error, as they:
- Lack necessary libraries for FIPS verification
- Cannot support FIPS-enabled OpenSSL
