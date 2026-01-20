# Yealink YMCS API Client Library - Technical Manual

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Package Overview](#package-overview)
4. [Client Initialization](#client-initialization)
5. [Authentication](#authentication)
6. [Device Management](#device-management)
7. [Data Types](#data-types)
8. [Error Handling](#error-handling)
9. [Advanced Usage](#advanced-usage)
10. [Code Examples](#code-examples)

---

## Introduction

The Yealink YMCS API Client Library provides a comprehensive Go interface for interacting with the Yealink Management Cloud Service (YMCS) API version 2. This library implements OAuth 2.0 authentication with automatic token management and offers a complete set of device management operations.

This library is currently concentrated on the 5.1 RPS Device Management portion of the YMCS specifications, offering robust support for provisioning and managing devices through the RPS endpoints. It implements the OAuth 2.0 client credentials flow with automatic token management.

### Key Features

- Full OAuth 2.0 client credentials flow implementation
- Automatic token caching and refresh
- Device lifecycle management (search, add, delete)
- Device PIN retrieval for provisioning
- Batch operations support
- MAC address normalization (supports multiple formats)
- Thread-safe implementation for concurrent use
- Comprehensive type definitions
- Built-in debug capabilities


### API Endpoints Reference

The library abstracts the following YMCS API v2 endpoints:

| Operation | Endpoint | HTTP Method | Library Method |
|-----------|----------|-------------|----------------|
| Get Token | `/v2/token` | POST | `GetAccessToken()` |
| Search Devices | `/v2/rps/listDevices` | POST | `SearchDevices()` |
| Get Device Details | `/v2/rps/devices/{id}` | GET | `GetDeviceDetails()` |
| Get Device PINs | `/v2/rps/listDevicePins` | POST | `GetDevicePINs()` |
| Add Device | `/v2/rps/devices` | POST | `AddDevice()` |
| Add by MAC (Batch) | `/v2/rps/addDevicesByMac` | POST | `AddDevicesByMac()` |
| Delete Devices | `/v2/rps/delDevices` | POST | `DeleteDevices()` |

### Package Information

- **Package Name**: `ymcs`
- **Import Path (this repository)**: `github.com/nethesis/falconieri/libs/ymcs`
- **Go Version**: 1.17 or higher
- **API Version**: v2
- **No External Dependencies**: Uses only Go standard library

Note: the client uses `time.Time` millisecond helpers such as `time.Now().UnixMilli()` / `time.UnixMilli()`, which require Go 1.17+.

---

## Installation

### Using Go Modules

Add the library to your project:

```bash
go get github.com/nethesis/falconieri@latest
```

### Import Statement

```go
import "github.com/nethesis/falconieri/libs/ymcs"
```

---

## Package Overview

The library is organized into the following components:

### Core Components

- **Client**: Main client structure for API interactions
- **Authentication**: OAuth 2.0 token management
- **Device Operations**: CRUD operations for device management
- **Types**: Comprehensive data structures for requests and responses

### File Structure

```
libs/ymcs/
├── client.go        # Core client implementation and HTTP request handling
├── auth.go          # OAuth 2.0 authentication and token management
├── devices.go       # Device management operations
├── errors.go        # YMCS error parsing and structured error type
└── types.go         # Data type definitions
```

---

## Client Initialization

### Creating a New Client

The `NewClient` function creates a new instance of the Yealink YMCS API client.

#### Function Signature

```go
func NewClient(baseURL, clientID, clientSecret string) *Client
```

#### Parameters

- **baseURL** (string): The YMCS API base URL
  - Europe: `https://eu-api.ymcs.yealink.com`
  - US: `https://us-api.ymcs.yealink.com`
  - Custom: Your on-premises YMCS server URL
  
- **clientID** (string): OAuth 2.0 client ID provided by Yealink

- **clientSecret** (string): OAuth 2.0 client secret provided by Yealink

#### Returns

- **\*Client**: Pointer to a new Client instance

#### Example

```go
client := ymcs.NewClient(
    "https://eu-api.ymcs.yealink.com",
    "your-client-id",
    "your-client-secret",
)
```

### Client Structure

The Client structure contains the following fields:

```go
type Client struct {
    BaseURL      string        // API base URL
    ClientID     string        // OAuth client ID
    ClientSecret string        // OAuth client secret
    HTTPClient   *http.Client  // Underlying HTTP client
    Debug        bool          // Enable debug logging
    
    // Internal fields (managed automatically)
    accessToken  string        // Cached OAuth token
    tokenExpiry  time.Time     // Token expiration time
    
    // Last request/response (overwritten on each call)
    LastRequest     *http.Request
    LastRequestBody string
    LastResponse    *http.Response
    LastRespBody    string
}
```

### Client Configuration

#### Setting Debug Mode

Enable debug mode to capture detailed request/response information:

```go
client := ymcs.NewClient(baseURL, clientID, clientSecret)
client.Debug = true
```

The client always stores the last request/response details in `LastRequest`, `LastRequestBody`, `LastResponse`, and `LastRespBody` (overwritten on every request). When `Debug` is enabled it also emits log messages for token usage and refresh.

**Note on concurrent access**: These debug fields are protected by an internal mutex for thread-safe writes. If you need to read these fields while the client may be processing concurrent requests, you should copy the values immediately after the operation completes to avoid potential races.

#### Custom HTTP Client

Replace the default HTTP client with a custom one:

```go
client := ymcs.NewClient(baseURL, clientID, clientSecret)
client.HTTPClient = &http.Client{
    Timeout: 60 * time.Second,
    Transport: customTransport,
}
```

---

## Authentication

### Overview

The library implements OAuth 2.0 client credentials flow with automatic token management. Tokens are cached and automatically refreshed when expired.

### GetAccessToken

Retrieves and caches an OAuth 2.0 access token.

#### Function Signature

```go
func (c *Client) GetAccessToken() (string, error)
```

#### Returns

- **string**: The access token
- **error**: Error if authentication fails

#### Behavior

1. Returns cached token if still valid (with 60-second buffer)
2. Automatically requests new token when expired
3. Caches token for subsequent requests
4. Thread-safe token management

#### Example

```go
token, err := client.GetAccessToken()
if err != nil {
    log.Fatalf("Authentication failed: %v", err)
}
fmt.Printf("Access token: %s\n", token)
```

#### Error Conditions

- Invalid credentials (401 Unauthorized)
- Network connectivity issues
- Malformed response from server
- Invalid client ID/secret format

#### Notes

- You typically don't need to call this method directly
- All device management methods automatically handle authentication
- Token caching improves performance by reducing authentication requests

---

## Device Management

### SearchDevices

Searches for devices with optional MAC address filtering and pagination support.

#### Function Signature

```go
func (c *Client) SearchDevices(mac string, skip, limit int, autoCount bool) (*DeviceSearchResponse, error)
```

#### Parameters

- **mac** (string): MAC address filter (optional, empty string for no filter)
  - Supports formats: `aa:bb:cc:dd:ee:ff`, `AA-BB-CC-DD-EE-FF`, `aabbccddeeff`
  - Automatically normalized to lowercase without separators
  
- **skip** (int): Number of records to skip (for pagination)

- **limit** (int): Maximum number of records to return

- **autoCount** (bool): Whether to include total count in response

#### Returns

- **\*DeviceSearchResponse**: Search results with device list and metadata
- **error**: Error if search fails

#### Example

```go
// Search for specific device
devices, err := client.SearchDevices("aa:bb:cc:dd:ee:ff", 0, 10, true)
if err != nil {
    log.Fatalf("Search failed: %v", err)
}

fmt.Printf("Found %d devices\n", devices.Total)
for _, device := range devices.Data {
    fmt.Printf("MAC: %s, ID: %s\n", device.MAC, device.ID)
}

// List all devices with pagination
devices, err := client.SearchDevices("", 0, 100, true)
if err != nil {
    log.Fatalf("Search failed: %v", err)
}
```

#### Response Structure

The `DeviceSearchResponse` contains:

```go
type DeviceSearchResponse struct {
    Skip  int      `json:"skip"`   // Number of skipped records
    Limit int      `json:"limit"`  // Maximum records returned
    Total int      `json:"total"`  // Total matching records
    Data  []Device `json:"data"`   // Array of device objects
}
```

---

### GetDeviceDetails

Retrieves detailed information about a specific device by its ID.

#### Function Signature

```go
func (c *Client) GetDeviceDetails(deviceID string) (*DeviceDetails, error)
```

#### Parameters

- **deviceID** (string): The unique device identifier (obtained from search results)

#### Returns

- **\*DeviceDetails**: Detailed device information
- **error**: Error if retrieval fails

#### Example

```go
details, err := client.GetDeviceDetails("device-id-123")
if err != nil {
    log.Fatalf("Failed to get details: %v", err)
}

fmt.Printf("Device MAC: %s\n", details.MAC)
if details.SN != nil {
    fmt.Printf("Serial Number: %s\n", *details.SN)
}
if details.UniqueServerURL != nil {
    fmt.Printf("Server URL: %s\n", *details.UniqueServerURL)
}
```

---

### GetDevicePINs

Retrieves PINs for multiple devices in a single request.

#### Function Signature

```go
func (c *Client) GetDevicePINs(macs []string) ([]DevicePIN, error)
```

#### Parameters

- **macs** ([]string): Array of MAC addresses
  - Supports any MAC format (automatically normalized)
  - No limit on array size specified in API

#### Returns

- **[]DevicePIN**: Array of MAC/PIN pairs
- **error**: Error if retrieval fails

#### Example

```go
macs := []string{
    "aa:bb:cc:dd:ee:01",
    "aa:bb:cc:dd:ee:02",
    "aa:bb:cc:dd:ee:03",
}

pins, err := client.GetDevicePINs(macs)
if err != nil {
    log.Fatalf("Failed to get PINs: %v", err)
}

for _, pin := range pins {
    fmt.Printf("MAC: %s, PIN: %s\n", pin.MAC, pin.PIN)
}
```

---

### GetSingleDevicePIN

Convenience method to retrieve PIN for a single device.

#### Function Signature

```go
func (c *Client) GetSingleDevicePIN(mac string) (string, error)
```

#### Parameters

- **mac** (string): MAC address of the device

#### Returns

- **string**: The device PIN (may be an empty string if no PIN is available)
- **error**: Error if the request fails (authentication, network, or API error)

#### Example

```go
pin, err := client.GetSingleDevicePIN("aa:bb:cc:dd:ee:ff")
if err != nil {
    log.Fatalf("Failed to get PIN: %v", err)
}

fmt.Printf("Device PIN: %s\n", pin)
```

#### Error Conditions
- MAC address invalid
- Network or authentication errors

---

### AddDevice

Adds a new device to the YMCS system with full metadata support.

#### Function Signature

```go
func (c *Client) AddDevice(req AddDeviceRequest) (*AddDeviceResponse, error)
```

#### Parameters

- **req** (AddDeviceRequest): Device information structure

#### AddDeviceRequest Structure

```go
type AddDeviceRequest struct {
    MAC             string   `json:"mac"`                      // Required
    SN              string   `json:"sn"`                       // Required
    ServerID        *string  `json:"serverId,omitempty"`       // Optional
    UniqueServerURL *string  `json:"uniqueServerUrl,omitempty"` // Optional
    AuthName        *string  `json:"authName,omitempty"`       // Optional
    Password        *string  `json:"password,omitempty"`       // Optional
    Remark          *string  `json:"remark,omitempty"`         // Optional
}
```

#### Returns

- **\*AddDeviceResponse**: Information about the added device
- **error**: Error if addition fails

#### Example

```go
// Basic device addition
req := ymcs.AddDeviceRequest{
    MAC: "aa:bb:cc:dd:ee:ff",
    SN:  "SN1234567890",
}

device, err := client.AddDevice(req)
if err != nil {
    log.Fatalf("Failed to add device: %v", err)
}

fmt.Printf("Device added with ID: %s\n", device.ID)

// Device with provisioning URL
serverURL := "https://provisioning.example.com"
remark := "Conference room phone"

req := ymcs.AddDeviceRequest{
    MAC:             "aa:bb:cc:dd:ee:ff",
    SN:              "SN1234567890",
    UniqueServerURL: &serverURL,
    Remark:          &remark,
}

device, err := client.AddDevice(req)
if err != nil {
    log.Fatalf("Failed to add device: %v", err)
}
```

#### Error Conditions

- Device already exists (duplicate MAC or SN)
- Invalid MAC address format
- Invalid serial number
- Missing required fields
- Authentication or permission errors

---

### AddDevicesByMac

Adds one or more devices by MAC address only, without requiring serial numbers. Supports batch operations.

#### Function Signature

```go
func (c *Client) AddDevicesByMac(devices []AddDeviceByMacRequest) (*AddDevicesByMacResponse, error)
```

#### Parameters

- **devices** ([]AddDeviceByMacRequest): Array of device requests (maximum 100)

#### AddDeviceByMacRequest Structure

```go
type AddDeviceByMacRequest struct {
    MAC             string   `json:"mac"`                      // Required
    ServerID        *string  `json:"serverId,omitempty"`       // Optional
    UniqueServerURL *string  `json:"uniqueServerUrl,omitempty"` // Optional
    AuthName        *string  `json:"authName,omitempty"`       // Optional
    Password        *string  `json:"password,omitempty"`       // Optional
    Remark          *string  `json:"remark,omitempty"`         // Optional
}
```

#### Returns

- **\*AddDevicesByMacResponse**: Batch operation results with success/failure counts
- **error**: Error if request fails

#### Example

```go
// Batch add multiple devices
serverURL := "https://provisioning.example.com"

devices := []ymcs.AddDeviceByMacRequest{
    {
        MAC:             "aa:bb:cc:dd:ee:01",
        UniqueServerURL: &serverURL,
    },
    {
        MAC:             "aa:bb:cc:dd:ee:02",
        UniqueServerURL: &serverURL,
    },
    {
        MAC:             "aa:bb:cc:dd:ee:03",
        UniqueServerURL: &serverURL,
    },
}

result, err := client.AddDevicesByMac(devices)
if err != nil {
    log.Fatalf("Batch add failed: %v", err)
}

fmt.Printf("Total: %d, Success: %d, Failed: %d\n",
    result.Total, result.SuccessCount, result.FailureCount)

// Handle partial failures
if len(result.Errors) > 0 {
    for _, addErr := range result.Errors {
        fmt.Printf("Failed to add %s: %s\n", addErr.MAC, addErr.ErrorInfo)
    }
}
```

#### Batch Operation Limits

- Maximum 100 devices per request
- Partial success supported (some devices may succeed while others fail)
- Individual errors returned in response

---

### AddDeviceByMacSingle

Convenience method to add a single device by MAC address without serial number.

#### Function Signature

```go
func (c *Client) AddDeviceByMacSingle(mac string, uniqueServerURL *string) (*AddDevicesByMacResponse, error)
```

#### Parameters

- **mac** (string): Device MAC address
- **uniqueServerURL** (*string): Provisioning server URL (optional, can be nil)

#### Returns

- **\*AddDevicesByMacResponse**: Operation result
- **error**: Error if addition fails

#### Example

```go
serverURL := "https://provisioning.example.com"
result, err := client.AddDeviceByMacSingle("aa:bb:cc:dd:ee:ff", &serverURL)
if err != nil {
    log.Fatalf("Failed to add device: %v", err)
}

if result.SuccessCount > 0 {
    fmt.Println("Device added successfully")
}
```

---

### DeleteDevices

Deletes one or more devices from the YMCS system.

#### Function Signature

```go
func (c *Client) DeleteDevices(deviceIdType string, deviceIds []string) (*DeleteDevicesResponse, error)
```

#### Parameters

- **deviceIdType** (string): Type of identifier used
  - `"mac"`: Delete by MAC address
  - `"id"`: Delete by device ID
  
- **deviceIds** ([]string): Array of device identifiers

#### Returns

- **\*DeleteDevicesResponse**: Deletion results with success/failure counts
- **error**: Error if request fails

#### Example

```go
// Delete by MAC addresses
macs := []string{
    "aa:bb:cc:dd:ee:01",
    "aa:bb:cc:dd:ee:02",
    "aa:bb:cc:dd:ee:03",
}

result, err := client.DeleteDevices("mac", macs)
if err != nil {
    log.Fatalf("Delete failed: %v", err)
}

fmt.Printf("Deleted %d of %d devices\n", result.SuccessCount, result.Total)

// Handle errors
if len(result.Errors) > 0 {
    for _, delErr := range result.Errors {
        fmt.Printf("Error on %s: %s\n", delErr.Field, delErr.Msg)
    }
}

// Delete by device IDs
ids := []string{"device-id-1", "device-id-2"}
result, err := client.DeleteDevices("id", ids)
```

---

### DeleteDeviceByMAC

Convenience method to delete a single device by MAC address.

#### Function Signature

```go
func (c *Client) DeleteDeviceByMAC(mac string) (*DeleteDevicesResponse, error)
```

#### Parameters

- **mac** (string): Device MAC address

#### Returns

- **\*DeleteDevicesResponse**: Deletion result
- **error**: Error if deletion fails

#### Example

```go
result, err := client.DeleteDeviceByMAC("aa:bb:cc:dd:ee:ff")
if err != nil {
    log.Fatalf("Delete failed: %v", err)
}

if result.SuccessCount > 0 {
    fmt.Println("Device deleted successfully")
}
```

---

### DeleteDeviceByID

Convenience method to delete a single device by its ID.

#### Function Signature

```go
func (c *Client) DeleteDeviceByID(deviceID string) (*DeleteDevicesResponse, error)
```

#### Parameters

- **deviceID** (string): Device identifier

#### Returns

- **\*DeleteDevicesResponse**: Deletion result
- **error**: Error if deletion fails

#### Example

```go
result, err := client.DeleteDeviceByID("device-id-123")
if err != nil {
    log.Fatalf("Delete failed: %v", err)
}

if result.SuccessCount > 0 {
    fmt.Println("Device deleted successfully")
}
```

---

## Data Types

### Device

Represents a device in the YMCS system.

```go
type Device struct {
    ID              string  `json:"id"`
    MAC             string  `json:"mac"`
    SN              *string `json:"sn"`
    ServerID        *string `json:"serverId"`
    ServerName      *string `json:"serverName"`
    ServerURL       *string `json:"serverUrl"`
    UniqueServerURL *string `json:"uniqueServerUrl"`
    IPAddress       *string `json:"ipAddress"`
    Remark          *string `json:"remark"`
    DateRegistered  *int64  `json:"dateRegistered"`   // Unix timestamp (milliseconds)
    LastConnected   *int64  `json:"lastConnected"`    // Unix timestamp (milliseconds)
}
```

#### Field Descriptions

- **ID**: Unique device identifier assigned by YMCS
- **MAC**: Device MAC address (normalized format: lowercase, no separators)
- **SN**: Serial number (nil if not provided)
- **ServerID**: Provisioning server identifier
- **ServerName**: Provisioning server name
- **ServerURL**: Provisioning server URL
- **UniqueServerURL**: Device-specific provisioning URL
- **IPAddress**: Last known IP address
- **Remark**: User-defined note or description
- **DateRegistered**: Registration timestamp in milliseconds since Unix epoch
- **LastConnected**: Last connection timestamp in milliseconds since Unix epoch

#### Usage Example

```go
devices, _ := client.SearchDevices("", 0, 10, true)
for _, device := range devices.Data {
    fmt.Printf("Device ID: %s\n", device.ID)
    fmt.Printf("MAC: %s\n", device.MAC)
    
    // Check optional fields
    if device.SN != nil {
        fmt.Printf("SN: %s\n", *device.SN)
    }
    
    if device.DateRegistered != nil {
        regTime := time.UnixMilli(*device.DateRegistered)
        fmt.Printf("Registered: %s\n", regTime.Format("2006-01-02 15:04:05"))
    }
}
```

---

### DeviceDetails

Extended device information including authentication details.

```go
type DeviceDetails struct {
    Device                // Embedded Device struct
    AuthName *string `json:"authName"`
}
```

#### Additional Fields

- **AuthName**: Authentication username for device provisioning

---

### DevicePIN

Device PIN information for provisioning.

```go
type DevicePIN struct {
    MAC string `json:"mac"`
    PIN string `json:"pin"`
}
```

---

### TokenResponse

OAuth 2.0 token response structure.

```go
type TokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`  // Seconds until expiration
}
```

---

### AddDeviceResponse

Response from adding a device.

```go
type AddDeviceResponse struct {
    ID              string  `json:"id"`
    MAC             string  `json:"mac"`
    SN              string  `json:"sn"`
    ServerID        *string `json:"serverId,omitempty"`
    UniqueServerURL *string `json:"uniqueServerUrl,omitempty"`
    AuthName        *string `json:"authName,omitempty"`
    Remark          *string `json:"remark,omitempty"`
}
```

---

### AddDevicesByMacResponse

Response from batch add operations.

```go
type AddDevicesByMacResponse struct {
    Total        int        `json:"total"`         // Total devices in request
    SuccessCount int        `json:"successCount"`  // Number successfully added
    FailureCount int        `json:"failureCount"`  // Number that failed
    Errors       []AddError `json:"errors"`        // Details of failures
}
```

#### AddError Structure

```go
type AddError struct {
    MAC       string `json:"mac"`
    SN        string `json:"sn"`
    ErrorInfo string `json:"errorInfo"`
}
```

---

### DeleteDevicesResponse

Response from delete operations.

```go
type DeleteDevicesResponse struct {
    Total        int       `json:"total"`         // Total devices in request
    SuccessCount int       `json:"successCount"`  // Number successfully deleted
    FailureCount int       `json:"failureCount"`  // Number that failed
    Errors       []OpError `json:"errors"`        // Details of failures
}
```

#### OpError Structure

```go
type OpError struct {
    Field string `json:"field"`
    Msg   string `json:"msg"`
}
```

---

## Error Handling

### Error Types

All methods return standard Go errors.

When the YMCS API returns a JSON error payload, the client returns an `APIError` (see `errors.go`) enriched with:

- HTTP status code (`HTTPStatus`)
- YMCS `code` (when provided)
- OAuth `error` (when provided)
- `requestId` (when provided)
- a best-effort `FriendlyMessage` for known error codes

When the response body is not a YMCS JSON error payload, the client falls back to a generic error that still includes the HTTP status code and raw body.

#### Common Error Pattern (with inspection)

```go
import (
    "errors"
    "log"

    "github.com/nethesis/falconieri/libs/ymcs"
)

devices, err := client.SearchDevices("aa:bb:cc:dd:ee:ff", 0, 10, true)
if err != nil {
    var apiErr ymcs.APIError
    if errors.As(err, &apiErr) {
        log.Printf("YMCS error: status=%d code=%s requestId=%s message=%s",
            apiErr.HTTPStatus, apiErr.Code, apiErr.RequestID, apiErr.Message)
        return
    }
    log.Printf("request failed: %v", err)
    return
}
_ = devices
```

### Batch Operation Error Handling

Batch operations support partial success:

```go
result, err := client.AddDevicesByMac(devices)
if err != nil {
    log.Fatalf("Request failed: %v", err)
}

// Even if err is nil, check for partial failures
if result.FailureCount > 0 {
    for _, addErr := range result.Errors {
        log.Printf("Failed to add %s: %s", addErr.MAC, addErr.ErrorInfo)
    }
}

// Process successful additions
if result.SuccessCount > 0 {
    log.Printf("Successfully added %d devices", result.SuccessCount)
}
```

### Network Error Handling

Network-related errors are wrapped with context:

```go
devices, err := client.SearchDevices("", 0, 10, true)
if err != nil {
    // Could be network timeout, DNS failure, connection refused, etc.
    log.Printf("Network error: %v", err)
    // Implement retry logic if needed
}
```

---

## Advanced Usage

### MAC Address Normalization

The library automatically normalizes MAC addresses to lowercase without separators.

#### Supported Input Formats

```go
// All these formats are equivalent:
client.SearchDevices("AA:BB:CC:DD:EE:FF", 0, 10, true)
client.SearchDevices("aa:bb:cc:dd:ee:ff", 0, 10, true)
client.SearchDevices("AA-BB-CC-DD-EE-FF", 0, 10, true)
client.SearchDevices("aabbccddeeff", 0, 10, true)

// All normalized to: "aabbccddeeff"
```

### Pagination

Implement pagination for large device lists:

```go
func getAllDevices(client *ymcs.Client) ([]ymcs.Device, error) {
    var allDevices []ymcs.Device
    skip := 0
    limit := 100

    for {
        result, err := client.SearchDevices("", skip, limit, true)
        if err != nil {
            return nil, err
        }

        allDevices = append(allDevices, result.Data...)

        // Check if we've retrieved all devices
        if skip+len(result.Data) >= result.Total {
            break
        }

        skip += limit
    }

    return allDevices, nil
}
```

### Concurrent Operations

The client is safe for concurrent use:

```go
func processDevicesConcurrently(client *ymcs.Client, macs []string) {
    var wg sync.WaitGroup
    results := make(chan string, len(macs))

    for _, mac := range macs {
        wg.Add(1)
        go func(m string) {
            defer wg.Done()

            pin, err := client.GetSingleDevicePIN(m)
            if err != nil {
                results <- fmt.Sprintf("Error for %s: %v", m, err)
                return
            }

            results <- fmt.Sprintf("%s: %s", m, pin)
        }(mac)
    }

    wg.Wait()
    close(results)

    for result := range results {
        fmt.Println(result)
    }
}
```

### Custom HTTP Client Configuration

Configure timeouts, proxies, and TLS settings:

```go
import (
    "crypto/tls"
    "net/http"
    "net/url"
    "time"
)

// Create custom transport
proxyURL, _ := url.Parse("http://proxy.example.com:8080")
transport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
    TLSClientConfig: &tls.Config{
        InsecureSkipVerify: false,
    },
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}

// Create custom HTTP client
httpClient := &http.Client{
    Timeout:   60 * time.Second,
    Transport: transport,
}

// Use with YMCS client
client := ymcs.NewClient(baseURL, clientID, clientSecret)
client.HTTPClient = httpClient
```

### Debug Mode for Troubleshooting

Enable debug mode to capture request/response details:

```go
client := ymcs.NewClient(baseURL, clientID, clientSecret)
client.Debug = true

_, err := client.SearchDevices("aa:bb:cc:dd:ee:ff", 0, 10, true)
if err != nil {
    // Inspect last request
    if client.LastRequest != nil {
        fmt.Printf("Request URL: %s\n", client.LastRequest.URL)
        fmt.Printf("Request Method: %s\n", client.LastRequest.Method)
        fmt.Printf("Request Body: %s\n", client.LastRequestBody)
    }

    // Inspect last response
    if client.LastResponse != nil {
        fmt.Printf("Response Status: %s\n", client.LastResponse.Status)
        fmt.Printf("Response Body: %s\n", client.LastRespBody)
    }
}
```

### Device Lifecycle Management

Complete device lifecycle example:

```go
func manageDeviceLifecycle(client *ymcs.Client, mac, sn string) error {
    // 1. Check if device exists
    devices, err := client.SearchDevices(mac, 0, 1, true)
    if err != nil {
        return fmt.Errorf("search failed: %w", err)
    }

    // 2. Delete if exists
    if devices.Total > 0 {
        fmt.Println("Device exists, deleting...")
        _, err := client.DeleteDeviceByMAC(mac)
        if err != nil {
            return fmt.Errorf("delete failed: %w", err)
        }
        time.Sleep(1 * time.Second) // Allow propagation
    }

    // 3. Add device
    serverURL := "https://provisioning.example.com"
    req := ymcs.AddDeviceRequest{
        MAC:             mac,
        SN:              sn,
        UniqueServerURL: &serverURL,
    }

    device, err := client.AddDevice(req)
    if err != nil {
        return fmt.Errorf("add failed: %w", err)
    }

    fmt.Printf("Device added: %s\n", device.ID)

    // 4. Get device PIN
    pin, err := client.GetSingleDevicePIN(mac)
    if err != nil {
        return fmt.Errorf("get PIN failed: %w", err)
    }

    fmt.Printf("Device PIN: %s\n", pin)

    // 5. Verify device details
    details, err := client.GetDeviceDetails(device.ID)
    if err != nil {
        return fmt.Errorf("get details failed: %w", err)
    }

    fmt.Printf("Device configured: %+v\n", details)

    return nil
}
```

---

## Code Examples

### Example 1: Basic Device Search

```go
package main

import (
    "fmt"
    "log"

    "github.com/nethesis/falconieri/libs/ymcs"
)

func main() {
    // Create client
    client := ymcs.NewClient(
        "https://eu-api.ymcs.yealink.com",
        "your-client-id",
        "your-client-secret",
    )

    // Search for devices
    devices, err := client.SearchDevices("", 0, 10, true)
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }

    fmt.Printf("Found %d devices\n", devices.Total)
    for i, device := range devices.Data {
        fmt.Printf("%d. MAC: %s, ID: %s\n", i+1, device.MAC, device.ID)
    }
}
```

### Example 2: Add Device with Provisioning

```go
package main

import (
    "fmt"
    "log"

    "github.com/nethesis/falconieri/libs/ymcs"
)

func main() {
    client := ymcs.NewClient(
        "https://eu-api.ymcs.yealink.com",
        "your-client-id",
        "your-client-secret",
    )

    // Prepare device configuration
    serverURL := "https://provisioning.example.com"
    remark := "Reception desk phone"
    authName := "admin"
    password := "secret123"

    req := ymcs.AddDeviceRequest{
        MAC:             "aa:bb:cc:dd:ee:ff",
        SN:              "SN1234567890",
        UniqueServerURL: &serverURL,
        Remark:          &remark,
        AuthName:        &authName,
        Password:        &password,
    }

    // Add device
    device, err := client.AddDevice(req)
    if err != nil {
        log.Fatalf("Failed to add device: %v", err)
    }

    fmt.Printf("Device added successfully!\n")
    fmt.Printf("ID: %s\n", device.ID)
    fmt.Printf("MAC: %s\n", device.MAC)
    fmt.Printf("SN: %s\n", device.SN)

    // Get device PIN
    pin, err := client.GetSingleDevicePIN(device.MAC)
    if err != nil {
        log.Fatalf("Failed to get PIN: %v", err)
    }

    fmt.Printf("PIN: %s\n", pin)
}
```

### Example 3: Batch Device Addition

```go
package main

import (
    "fmt"
    "log"

    "github.com/nethesis/falconieri/libs/ymcs"
)

func main() {
    client := ymcs.NewClient(
        "https://eu-api.ymcs.yealink.com",
        "your-client-id",
        "your-client-secret",
    )

    serverURL := "https://provisioning.example.com"

    // Prepare batch of devices
    devices := []ymcs.AddDeviceByMacRequest{
        {MAC: "aa:bb:cc:dd:ee:01", UniqueServerURL: &serverURL},
        {MAC: "aa:bb:cc:dd:ee:02", UniqueServerURL: &serverURL},
        {MAC: "aa:bb:cc:dd:ee:03", UniqueServerURL: &serverURL},
        {MAC: "aa:bb:cc:dd:ee:04", UniqueServerURL: &serverURL},
        {MAC: "aa:bb:cc:dd:ee:05", UniqueServerURL: &serverURL},
    }

    // Add devices
    result, err := client.AddDevicesByMac(devices)
    if err != nil {
        log.Fatalf("Batch add failed: %v", err)
    }

    fmt.Printf("Batch Operation Results:\n")
    fmt.Printf("Total: %d\n", result.Total)
    fmt.Printf("Success: %d\n", result.SuccessCount)
    fmt.Printf("Failed: %d\n", result.FailureCount)

    // Report errors
    if len(result.Errors) > 0 {
        fmt.Println("\nErrors:")
        for _, addErr := range result.Errors {
            fmt.Printf("  %s: %s\n", addErr.MAC, addErr.ErrorInfo)
        }
    }
}
```

### Example 4: Complete Device Management

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/nethesis/falconieri/libs/ymcs"
)

func main() {
    client := ymcs.NewClient(
        "https://eu-api.ymcs.yealink.com",
        "your-client-id",
        "your-client-secret",
    )

    mac := "aa:bb:cc:dd:ee:ff"
    sn := "SN1234567890"

    // Step 1: Search for existing device
    fmt.Println("Searching for existing device...")
    devices, err := client.SearchDevices(mac, 0, 1, true)
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }

    // Step 2: Delete if exists
    if devices.Total > 0 {
        fmt.Printf("Found existing device (ID: %s), deleting...\n", devices.Data[0].ID)
        _, err := client.DeleteDeviceByMAC(mac)
        if err != nil {
            log.Printf("Warning: Delete failed: %v", err)
        } else {
            fmt.Println("Device deleted, waiting for propagation...")
            time.Sleep(2 * time.Second)
        }
    }

    // Step 3: Add device
    fmt.Println("Adding device...")
    serverURL := "https://provisioning.example.com"
    req := ymcs.AddDeviceRequest{
        MAC:             mac,
        SN:              sn,
        UniqueServerURL: &serverURL,
    }

    device, err := client.AddDevice(req)
    if err != nil {
        log.Fatalf("Add failed: %v", err)
    }
    fmt.Printf("Device added with ID: %s\n", device.ID)

    // Step 4: Get device details
    fmt.Println("Retrieving device details...")
    details, err := client.GetDeviceDetails(device.ID)
    if err != nil {
        log.Fatalf("Get details failed: %v", err)
    }

    fmt.Printf("Device Details:\n")
    fmt.Printf("  MAC: %s\n", details.MAC)
    if details.SN != nil {
        fmt.Printf("  SN: %s\n", *details.SN)
    }
    if details.UniqueServerURL != nil {
        fmt.Printf("  Server URL: %s\n", *details.UniqueServerURL)
    }

    // Step 5: Get device PIN
    fmt.Println("Retrieving device PIN...")
    pin, err := client.GetSingleDevicePIN(mac)
    if err != nil {
        log.Fatalf("Get PIN failed: %v", err)
    }
    fmt.Printf("Device PIN: %s\n", pin)

    fmt.Println("\nDevice management completed successfully!")
}
```

### Example 5: Error Handling and Retry Logic

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/nethesis/falconieri/libs/ymcs"
)

func searchWithRetry(client *ymcs.Client, mac string, maxRetries int) (*ymcs.DeviceSearchResponse, error) {
    var lastErr error

    for attempt := 1; attempt <= maxRetries; attempt++ {
        devices, err := client.SearchDevices(mac, 0, 10, true)
        if err == nil {
            return devices, nil
        }

        lastErr = err
        if attempt < maxRetries {
            waitTime := time.Duration(attempt) * time.Second
            fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", attempt, err, waitTime)
            time.Sleep(waitTime)
        }
    }

    return nil, fmt.Errorf("all %d attempts failed, last error: %w", maxRetries, lastErr)
}

func main() {
    client := ymcs.NewClient(
        "https://eu-api.ymcs.yealink.com",
        "your-client-id",
        "your-client-secret",
    )

    devices, err := searchWithRetry(client, "aa:bb:cc:dd:ee:ff", 3)
    if err != nil {
        log.Fatalf("Search failed after retries: %v", err)
    }

    fmt.Printf("Found %d devices\n", devices.Total)
}
```

### Example 6: Working with Timestamps

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/nethesis/falconieri/libs/ymcs"
)

func main() {
    client := ymcs.NewClient(
        "https://eu-api.ymcs.yealink.com",
        "your-client-id",
        "your-client-secret",
    )

    devices, err := client.SearchDevices("", 0, 10, true)
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }

    for _, device := range devices.Data {
        fmt.Printf("Device: %s\n", device.MAC)

        if device.DateRegistered != nil {
            regTime := time.UnixMilli(*device.DateRegistered)
            fmt.Printf("  Registered: %s\n", regTime.Format("2006-01-02 15:04:05"))

            // Calculate age
            age := time.Since(regTime)
            fmt.Printf("  Age: %.0f days\n", age.Hours()/24)
        }

        if device.LastConnected != nil {
            lastTime := time.UnixMilli(*device.LastConnected)
            fmt.Printf("  Last seen: %s\n", lastTime.Format("2006-01-02 15:04:05"))

            // Check if device is active
            if time.Since(lastTime) < 24*time.Hour {
                fmt.Println("  Status: Active (connected within 24h)")
            } else {
                fmt.Println("  Status: Inactive")
            }
        }

        fmt.Println()
    }
}
```

---

## Best Practices

### Token Management

- Do not manually manage tokens; the client handles this automatically
- Token caching reduces authentication overhead
- Token refresh happens automatically with 60-second safety buffer

### Error Handling

- Always check for errors from API calls
- Use error wrapping for additional context
- Handle partial failures in batch operations
- Implement retry logic for transient network errors

### MAC Address Handling

- Use any convenient MAC format; normalization is automatic
- Store MACs in your preferred format
- The library ensures consistency with the API

### Resource Management

- Reuse client instances when possible
- The client is thread-safe for concurrent operations
- HTTP connections are pooled and reused automatically

### Batch Operations

- Use batch operations for multiple devices to reduce API calls
- Respect the 100-device limit per batch request
- Always check FailureCount and Errors in batch responses
- Implement partial retry logic for failed items

### Production Considerations

- Configure appropriate HTTP timeouts for your network environment
- Implement logging for audit trails
- Use debug mode during development and troubleshooting
- Consider rate limiting if making many concurrent requests
- Store credentials securely (environment variables, secret managers)

---

## Thread Safety

The client is designed for concurrent use:

- Token management is handled internally with proper synchronization
- Multiple goroutines can safely call client methods simultaneously
- HTTP client connection pooling is managed automatically
- No external synchronization required when using the client

---

## Performance Considerations

### Connection Pooling

The default HTTP client uses connection pooling:
- Default timeout: 30 seconds
- Connections are reused across requests
- Customize via `client.HTTPClient` if needed

### Token Caching

- Tokens are cached until expiration
- Automatic refresh reduces authentication overhead
- No token refresh on every request

### Batch Operations

- Use batch operations to reduce API calls
- Network round-trips are minimized
- Partial failures are handled gracefully
