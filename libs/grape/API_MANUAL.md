# Grape Provisioning API Client Library - Technical Manual

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Package Overview](#package-overview)
4. [Client Initialization](#client-initialization)
5. [Authentication](#authentication)
6. [Device Management](#device-management)
7. [Data Types](#data-types)
8. [Error Handling](#error-handling)
9. [Code Examples](#code-examples)
10. [Advanced Usage](#advanced-usage)

---

## Introduction

The Grape Provisioning API Client Library provides a comprehensive Go interface for interacting with the Grape provisioning service API. This library implements Hawk HTTP authentication (HMAC-based) and offers device management operations with automatic API navigation and caching.

### Key Features

- Full Hawk HTTP authentication implementation
- Automatic API navigation (HATEOAS-based discovery)
- Caching of API endpoints for improved performance
- Device provisioning and registration
- MAC address normalization (supports multiple formats)
- Thread-safe implementation for concurrent use
- Comprehensive type definitions
- Built-in debug capabilities

### API Endpoints Reference

The library abstracts the following Grape API endpoints:

| Operation | Endpoint | HTTP Method | Library Method |
|-----------|----------|-------------|----------------|
| Get Settings | `/settings/` | GET | Internal: `getProvisioningServerID()` |
| Get Token Info | `/tokens/{clientID}` | GET | Internal: `getEndpointsURL()` |
| Get Company Info | `{company_link}` | GET | Internal: `getEndpointsURL()` |
| Register Device | `{endpoints_url}{mac}` | PUT | `RegisterDevice()` |

### Package Information

- **Package Name**: `grape`
- **Import Path**: `github.com/nethesis/falconieri/libs/grape`
- **Minimum Go Version**: As specified in this repository's `go.mod` file
- **Authentication**: Hawk (HMAC-based)
- **Dependencies**: Go standard library plus `golang.org/x/sync/singleflight`

---

## Installation

### Using Go Modules

Add the library to your project:

```bash
go get github.com/nethesis/falconieri@latest
```

### Import Statement

```go
import "github.com/nethesis/falconieri/libs/grape"
```

---

## Package Overview

The library is organized into the following components:

### Core Components

- **Client**: Main client structure for API interactions with caching
- **Authentication**: Hawk HTTP authentication implementation
- **Device Operations**: Device registration and provisioning
- **Types**: Data structures for requests and responses
- **Error Handling**: Structured error parsing

### File Structure

```
libs/grape/
├── client.go        # Core client implementation and MAC normalization
├── auth.go          # Hawk authentication (nonce, MAC, signatures)
├── devices.go       # Device management operations
├── errors.go        # Grape API error parsing and structured error type
├── types.go         # Data type definitions
└── API_MANUAL.md    # This documentation file
```

---

## Client Initialization

### Creating a New Client

```go
package main

import "github.com/nethesis/falconieri/libs/grape"

func main() {
    client := grape.NewClient(
        "https://api.grape.example.com/",  // Base URL
        "your-client-id",                    // Client ID
        "your-client-secret",                // Client Secret
    )
}
```

### Client Configuration

The client is automatically configured with:
- HTTP timeout: 30 seconds
- Thread-safe API navigation caching
- Automatic MAC address normalization

### Debug Mode

Enable debug mode to inspect HTTP requests and responses:

```go
client := grape.NewClient(baseURL, clientID, clientSecret)
client.Debug = true

// After API calls, inspect debug information:
fmt.Printf("Last Request: %+v\n", client.LastRequest)
fmt.Printf("Last Request Body: %s\n", client.LastRequestBody)
fmt.Printf("Last Response: %+v\n", client.LastResponse)
fmt.Printf("Last Response Body: %s\n", client.LastRespBody)
```

---

## Authentication

### Hawk HTTP Authentication

The Grape API uses Hawk authentication, an HMAC-based HTTP authentication scheme. The library handles all authentication details automatically.

#### Hawk Authentication Flow

1. **Nonce Generation**: 16 bytes of cryptographically secure random data
2. **Payload Hashing**: SHA256 hash of request body in Hawk format
3. **MAC Calculation**: HMAC-SHA256 signature of normalized request
4. **Authorization Header**: Constructed with id, timestamp, nonce, hash, and MAC

#### Hawk Header Format

```
Hawk id="{clientID}", ts="{timestamp}", nonce="{nonce}", hash="{payloadHash}", mac="{mac}"
```

#### Authentication Components

- **ID**: Client identifier
- **Key**: Client secret for HMAC
- **Timestamp**: Unix timestamp
- **Nonce**: Cryptographically random value
- **Hash**: SHA256 payload hash
- **MAC**: HMAC-SHA256 signature

### Automatic Authentication

All API requests are automatically authenticated using Hawk. No manual token management is required.

---

## Device Management

### Register Device

Register a device with the Grape provisioning server.

#### Method Signature

```go
func (c *Client) RegisterDevice(mac, provisioningURL string) error
```

#### Parameters

- `mac`: MAC address in any format (e.g., `AA:BB:CC:DD:EE:FF`, `AA-BB-CC-DD-EE-FF`, `AABBCCDDEEFF`)
- `provisioningURL`: The URL of the provisioning server

#### Example

```go
client := grape.NewClient(baseURL, clientID, clientSecret)

err := client.RegisterDevice("00:15:65:4F:A1:2B", "https://provision.example.com/config.xml")
if err != nil {
    log.Fatalf("Failed to register device: %v", err)
}

fmt.Println("Device registered successfully")
```

#### MAC Address Formats

The library automatically normalizes MAC addresses. All these formats are accepted:

- `AA:BB:CC:DD:EE:FF` (colon-separated)
- `AA-BB-CC-DD-EE-FF` (dash-separated)
- `AA.BB.CC.DD.EE.FF` (dot-separated)
- `AABBCCDDEEFF` (no separators)
- `aabbccddeeff` (lowercase)

All formats are normalized to lowercase without separators: `aabbccddeeff`

### API Navigation Caching

The library automatically discovers and caches API endpoints:

1. **ProvisioningServer UUID**: Retrieved once from `/settings/`
2. **Endpoints URL**: Retrieved once through token and company links

These values are cached per client instance for improved performance. Subsequent `RegisterDevice()` calls reuse cached values.

---

## Data Types

### Setting

Represents a configuration setting in the Grape API.

```go
type Setting struct {
    UUID      string `json:"uuid"`
    ParamName string `json:"param_name"`
}
```

### TokenResponse

Response from the token endpoint containing navigation links.

```go
type TokenResponse struct {
    Links struct {
        Company string `json:"company"`
    } `json:"links"`
}
```

### CompanyResponse

Response from the company endpoint containing navigation links.

```go
type CompanyResponse struct {
    Links struct {
        Endpoints string `json:"endpoints"`
    } `json:"links"`
}
```

### DeviceData

Data structure for registering a device.

```go
type DeviceData struct {
    MAC                     string                            `json:"mac"`
    AutoprovisioningEnabled bool                              `json:"autoprovisioning_enabled"`
    SettingsManager         map[string]map[string]interface{} `json:"settings_manager"`
}
```

---

## Error Handling

### APIError Type

The library provides a structured error type for API errors.

```go
type APIError struct {
    StatusCode int    // HTTP status code
    Status     string // HTTP status message
    Message    string // Parsed error message
    Body       string // Raw response body
}
```

### Error Method

```go
func (e APIError) Error() string
```

Returns a formatted error message: `Grape API error (HTTP {code}): {message}`

### Checking for API Errors

```go
err := client.RegisterDevice(mac, url)
if err != nil {
    var apiErr grape.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error: HTTP %d - %s\n", apiErr.StatusCode, apiErr.Message)
        fmt.Printf("Raw Response: %s\n", apiErr.Body)
        return
    }

    fmt.Printf("Other Error: %v\n", err)
}
```

HTTP errors are returned as `APIError` values (they may be wrapped by the caller); use `errors.As` to check for them so you can still access status codes and the raw response body.
Be sure to import the standard library `errors` package when using `errors.As`.

### Common Error Scenarios

- **401 Unauthorized**: Invalid client credentials
- **404 Not Found**: Device or endpoint not found
- **400 Bad Request**: Invalid MAC address or parameters
- **500 Internal Server Error**: Grape API server error

---

## Code Examples

### Basic Device Registration

```go
package main

import (
    "fmt"
    "log"

    "github.com/nethesis/falconieri/libs/grape"
)

func main() {
    // Create client
    client := grape.NewClient(
        "https://api.grape.example.com/",
        "my-client-id",
        "my-client-secret",
    )

    // Register device
    mac := "00:15:65:4F:A1:2B"
    provURL := "https://provision.example.com/config.xml"

    err := client.RegisterDevice(mac, provURL)
    if err != nil {
        log.Fatalf("Failed to register device: %v", err)
    }

    fmt.Println("Device registered successfully!")
}
```

### Register Multiple Devices

```go
package main

import (
    "fmt"
    "log"

    "github.com/nethesis/falconieri/libs/grape"
)

func main() {
    client := grape.NewClient(baseURL, clientID, clientSecret)

    devices := []struct {
        MAC string
        URL string
    }{
        {"00:15:65:4F:A1:2B", "https://provision.example.com/device1.xml"},
        {"00:15:65:4F:A1:2C", "https://provision.example.com/device2.xml"},
        {"00:15:65:4F:A1:2D", "https://provision.example.com/device3.xml"},
    }

    for _, dev := range devices {
        err := client.RegisterDevice(dev.MAC, dev.URL)
        if err != nil {
            log.Printf("Failed to register %s: %v", dev.MAC, err)
            continue
        }
        fmt.Printf("Registered %s successfully\n", dev.MAC)
    }
}
```

### Error Handling Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/nethesis/falconieri/libs/grape"
)

func registerDevice(client *grape.Client, mac, url string) error {
    err := client.RegisterDevice(mac, url)
    if err != nil {
        // Check if it's an API error
        if apiErr, ok := err.(grape.APIError); ok {
            switch apiErr.StatusCode {
            case 401:
                return fmt.Errorf("authentication failed: check credentials")
            case 404:
                return fmt.Errorf("device or endpoint not found")
            case 400:
                return fmt.Errorf("invalid request: %s", apiErr.Message)
            default:
                return fmt.Errorf("API error %d: %s", apiErr.StatusCode, apiErr.Message)
            }
        }
        return fmt.Errorf("network or other error: %w", err)
    }
    return nil
}

func main() {
    client := grape.NewClient(baseURL, clientID, clientSecret)

    err := registerDevice(client, "00:15:65:4F:A1:2B", "https://provision.example.com/config.xml")
    if err != nil {
        log.Fatalf("Registration failed: %v", err)
    }

    fmt.Println("Device registered!")
}
```

### Debug Mode Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/nethesis/falconieri/libs/grape"
)

func main() {
    client := grape.NewClient(baseURL, clientID, clientSecret)
    client.Debug = true

    err := client.RegisterDevice("00:15:65:4F:A1:2B", "https://provision.example.com/config.xml")

    // Inspect the last request
    fmt.Println("=== Last Request ===")
    fmt.Printf("Method: %s\n", client.LastRequest.Method)
    fmt.Printf("URL: %s\n", client.LastRequest.URL)
    fmt.Printf("Headers: %+v\n", client.LastRequest.Header)
    fmt.Printf("Body: %s\n", client.LastRequestBody)

    // Inspect the last response
    fmt.Println("\n=== Last Response ===")
    fmt.Printf("Status: %d %s\n", client.LastResponse.StatusCode, client.LastResponse.Status)
    fmt.Printf("Headers: %+v\n", client.LastResponse.Header)
    fmt.Printf("Body: %s\n", client.LastRespBody)

    if err != nil {
        log.Fatalf("Error: %v", err)
    }
}
```

---

## Advanced Usage

### Singleton Pattern (Recommended)

For applications making multiple API calls, use a singleton pattern to benefit from caching:

```go
package main

import (
    "sync"

    "github.com/nethesis/falconieri/libs/grape"
)

var (
    grapeClient     *grape.Client
    grapeClientOnce sync.Once
)

func getGrapeClient() *grape.Client {
    grapeClientOnce.Do(func() {
        grapeClient = grape.NewClient(
            "https://api.grape.example.com/",
            "my-client-id",
            "my-client-secret",
        )
    })
    return grapeClient
}

func main() {
    // All calls use the same client instance (with cached API navigation)
    client := getGrapeClient()

    client.RegisterDevice("00:15:65:4F:A1:2B", "https://provision.example.com/config1.xml")
    client.RegisterDevice("00:15:65:4F:A1:2C", "https://provision.example.com/config2.xml")
    // Subsequent calls reuse cached ProvisioningServer UUID and endpoints URL
}
```

### Thread Safety

The client is thread-safe for concurrent use:

```go
package main

import (
    "sync"

    "github.com/nethesis/falconieri/libs/grape"
)

func main() {
    client := grape.NewClient(baseURL, clientID, clientSecret)

    var wg sync.WaitGroup
    devices := []string{
        "00:15:65:4F:A1:2B",
        "00:15:65:4F:A1:2C",
        "00:15:65:4F:A1:2D",
    }

    for _, mac := range devices {
        wg.Add(1)
        go func(m string) {
            defer wg.Done()
            err := client.RegisterDevice(m, "https://provision.example.com/config.xml")
            if err != nil {
                log.Printf("Failed to register %s: %v", m, err)
            }
        }(mac)
    }

    wg.Wait()
}
```

### Custom HTTP Client

To customize the HTTP client (e.g., for proxies or custom TLS):

```go
package main

import (
    "crypto/tls"
    "net/http"
    "time"

    "github.com/nethesis/falconieri/libs/grape"
)

func main() {
    client := grape.NewClient(baseURL, clientID, clientSecret)

    // Customize HTTP client
    client.HTTPClient = &http.Client{
        Timeout: 60 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: false,
            },
        },
    }

    // Use client normally
    err := client.RegisterDevice("00:15:65:4F:A1:2B", "https://provision.example.com/config.xml")
    // ...
}
```

### Cache Reset

If you need to reset cached API navigation (e.g., after configuration changes):

```go
// Create a new client instance to reset cache
client = grape.NewClient(baseURL, newClientID, newClientSecret)
```

---

## Best Practices

### 1. Use Singleton Pattern

Reuse the same client instance across your application to benefit from API navigation caching.

### 2. Error Handling

Always check for errors and handle API-specific error codes appropriately.

### 3. MAC Address Flexibility

Let the library handle MAC address normalization—accept any format from users.

### 4. Debug Mode in Development

Enable debug mode during development to troubleshoot API issues:

```go
client.Debug = true
```

Disable in production for performance.

### 5. Timeouts

The default 30-second timeout is suitable for most cases. Adjust if needed:

```go
client.HTTPClient.Timeout = 60 * time.Second
```

### 6. Thread Safety

The client is thread-safe, but create only one instance per application for optimal caching.

---

## Troubleshooting

### Authentication Failures (401)

- Verify `ClientID` and `ClientSecret` are correct
- Check that credentials have proper permissions
- Enable debug mode to inspect the Hawk authentication header

### Device Not Found (404)

- Verify the device MAC address is correct
- Ensure the device exists in the Grape system
- Check API navigation endpoints are accessible

### Invalid Request (400)

- Verify MAC address format (library normalizes automatically)
- Check provisioning URL is well-formed
- Review error message for specific validation issues

### Network Errors

- Check network connectivity to Grape API
- Verify base URL is correct and accessible
- Ensure no firewall blocking HTTP/HTTPS traffic
- Consider increasing timeout for slow networks

---

## Support

For issues, questions, or contributions:

- **Repository**: https://github.com/nethesis/falconieri
- **Issues**: https://github.com/nethesis/falconieri/issues

---

## License

Copyright (C) 2026 Nethesis S.r.l.

This library is part of the Falconieri project and is licensed under the GNU Affero General Public License v3.0 or later.

See LICENSE file for details.
