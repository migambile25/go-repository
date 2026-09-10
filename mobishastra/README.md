# Mobishastra SMS Client for Go

A robust Go client library for Mobishastra SMS Gateway API. This package provides easy-to-use methods for sending single and bulk SMS, checking balance, tracking delivery status, and handling callbacks.

## Features

- ✅ Send single SMS messages with custom Sender ID per message
- 📱 Send bulk SMS to multiple recipients
- 💰 Check account balance
- 📊 Track delivery status (DLR)
- 🔄 Support for scheduled messages
- 🌍 Unicode support for international languages
- 📞 Callback handling for incoming SMS and DLR
- 🔐 Environment-based configuration
- 🛡️ Built-in error handling and status mapping

## Installation

```bash
go get github.com/hekimapro/mobishastra
```

## Configuration

The package uses environment variables for configuration. Create a `.env` file in your project root or set system environment variables.

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `SMS_USER_ID` | Your SMS gateway username | Yes | - |
| `SMS_PASSWORD` | Your SMS gateway password | Yes | - |
| `SMS_SEND_URL` | SMS sending endpoint | No | `https://mshastra.com/sendurl.aspx` |
| `SMS_SEND_COMM_URL` | Bulk SMS endpoint | No | `https://mshastra.com/sendurlcomma.aspx` |
| `SMS_BALANCE_URL` | Balance check endpoint | No | `https://mshastra.com/balance.aspx` |
| `SMS_DELIVERY_STATUS_URL` | DLR status endpoint | No | `http://login.telkosh.com/dlrstatus_api.aspx` |
| `SMS_CALLBACK_SECRET` | Secret key for callback authentication | No | - |
| `SMS_CALLBACK_PORT` | Port for callback server | No | - |
| `SMS_COUNTRY_CODE` | Default country code | No | `ALL` |
| `SMS_PRIORITY` | Default message priority | No | `High` |

**Note:** `SMS_SENDER_ID` has been removed from configuration. Sender ID is now passed as a parameter when sending messages, allowing different Sender IDs per message.

### Example .env file

```env
SMS_USER_ID = PROFILE_ID
SMS_PASSWORD = your_password
SMS_COUNTRY_CODE = ALL
SMS_PRIORITY = High
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/hekimapro/mobishastra"
)

func main() {
    // Load configuration
    cfg := mobishastra.LoadConfig()

    // Create SMS client
    client := mobishastra.NewSMSClient(cfg)

    // Authenticate and check balance
    balance, err := client.Authenticate()
    if err != nil {
        log.Fatal("Authentication failed:", err)
    }
    fmt.Printf("Balance: %.2f credits\n", balance.Balance)

    // Send a single SMS with custom Sender ID
    senderID := "YOURSMS"
    response, err := client.SendSMS(senderID, "1234567890", "Hello from Go!")
    if err != nil {
        log.Fatal("Failed to send SMS:", err)
    }

    if response.Success {
        fmt.Println("SMS sent successfully!")
        fmt.Printf("Message ID: %s\n", response.MessageID)
    } else {
        fmt.Printf("Failed to send: %s\n", response.Error)
    }
}
```

## API Reference

### Types

#### Config
Configuration structure containing API credentials and endpoints.

```go
type Config struct {
    UserID            string
    Password          string
    SendSMSURL        string
    SendSMSCommaURL   string
    CheckBalanceURL   string
    DeliveryStatusURL string
    CallbackSecret    string
    CallbackPort      string
    CountryCode       string
    Priority          string
}
```

#### SMSClient
Main client for interacting with the SMS gateway.

```go
type SMSClient struct {
    config     *config.Config
    httpClient *http.Client
}
```

#### Response Types

```go
// Single SMS response
type SingleSMSResponse struct {
    Success      bool
    MessageID    string
    ResponseCode string
    ResponseText string
    Error        string
}

// Bulk SMS response
type BulkSMSResponse struct {
    Success   bool
    Responses []SingleSMSResponseData
    TotalSent int
}

// Single SMS response data for bulk operations
type SingleSMSResponseData struct {
    MobileNo     string
    MessageID    string
    ResponseCode string
    ResponseText string
    Success      bool
}

// Balance response
type BalanceResponse struct {
    Success bool
    Balance float64
    Credits int
    Error   string
}

// Delivery status response
type DeliveryStatusResponse struct {
    MessageID   string
    Status      string
    Time        time.Time
    IsDelivered bool
}

// Inbound SMS callback
type InboundSMSCallback struct {
    ShortCode   string
    PhoneNumber string
    Keyword     string
    Message     string
    Timestamp   time.Time
}

// Delivery status callback
type DeliveryStatusCallback struct {
    MessageID   string
    Status      string
    Time        time.Time
    PhoneNumber string
    ErrorCode   string
}
```

### Methods

#### Client Initialization

##### `NewSMSClient(cfg *config.Config) *SMSClient`
Creates a new SMS client instance with the provided configuration.

#### Authentication & Balance

##### `Authenticate() (*BalanceResponse, error)`
Authenticates the user by checking account balance. Returns balance information or error.

##### `CheckBalance() (*BalanceResponse, error)`
Returns current SMS credits balance.

#### Sending Messages

##### `SendSMS(senderID, phoneNumber, message string) (*SingleSMSResponse, error)`
Sends a single SMS message with the specified Sender ID using default configuration.

**Parameters:**
- `senderID`: Your Sender ID (max 13 characters, must be approved by provider)
- `phoneNumber`: Recipient phone number
- `message`: SMS message content

##### `SendSMSWithOptions(senderID, phoneNumber, message, scheduledAt, showError string) (*SingleSMSResponse, error)`
Sends a single SMS with additional options.

**Parameters:**
- `senderID`: Your Sender ID
- `phoneNumber`: Recipient phone number
- `message`: SMS message content
- `scheduledAt`: Schedule delivery time (format depends on API)
- `showError`: Show detailed error information

##### `SendBulkSMS(senderID string, phoneNumbers []string, message string) (*BulkSMSResponse, error)`
Sends the same message to multiple recipients using the specified Sender ID.

##### `SendBulkSMSWithOptions(senderID string, phoneNumbers []string, message, scheduledAt, showError string) (*BulkSMSResponse, error)`
Sends bulk SMS with scheduling and error options.

##### `SendUnicodeSMS(senderID, phoneNumber, message string) (*SingleSMSResponse, error)`
Sends Unicode SMS messages (supports Arabic, Chinese, etc.) with the specified Sender ID.

##### `SendXMLSMS(senderID, phoneNumber, message, language string) (string, error)`
Sends SMS using XML API with language specification.

**Parameters:**
- `senderID`: Your Sender ID
- `phoneNumber`: Recipient phone number
- `message`: SMS message content
- `language`: Language code for the message

#### Delivery Status

##### `GetDeliveryStatus(messageID string) (*DeliveryStatusResponse, error)`
Checks delivery status for a single message.

##### `GetBulkDeliveryStatus(messageIDs []string) ([]*DeliveryStatusResponse, error)`
Checks delivery status for multiple messages.

## Usage Examples

### Sending SMS with Different Sender IDs

```go
// Send from different sender IDs for different purposes
client := mobishastra.NewSMSClient(cfg)

// Transactional SMS
txnResponse, err := client.SendSMS("TXNSMS", "1234567890", "Your transaction of $100 was successful")

// Promotional SMS
promoResponse, err := client.SendSMS("PROMO", "1234567890", "Special offer just for you!")

// OTP SMS
otpResponse, err := client.SendSMS("OTPSMS", "1234567890", "Your OTP is 123456")
```

### Sending Scheduled SMS

```go
senderID := "YOURSMS"
scheduledTime := "2024-12-25 10:00:00"
response, err := client.SendSMSWithOptions(
    senderID,
    "1234567890",
    "Merry Christmas!",
    scheduledTime,
    "1", // Show detailed errors
)

if err != nil {
    log.Printf("Failed to schedule SMS: %v", err)
} else if response.Success {
    log.Printf("SMS scheduled successfully. Message ID: %s", response.MessageID)
}
```

### Bulk SMS with Detailed Response

```go
senderID := "YOURSMS"
phoneNumbers := []string{
    "1234567890",
    "0987654321",
    "5551234567",
}

response, err := client.SendBulkSMS(senderID, phoneNumbers, "Welcome to our service!")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total sent: %d/%d\n", response.TotalSent, len(phoneNumbers))
for _, resp := range response.Responses {
    if resp.Success {
        fmt.Printf("✓ %s: Sent (ID: %s)\n", resp.MobileNo, resp.MessageID)
    } else {
        fmt.Printf("✗ %s: Failed - %s (Code: %s)\n",
            resp.MobileNo, resp.ResponseText, resp.ResponseCode)
    }
}
```

### Tracking Delivery Status

```go
// After sending SMS, track delivery
messageID := "1234567890"

// Wait a few seconds for processing
time.Sleep(5 * time.Second)

status, err := client.GetDeliveryStatus(messageID)
if err != nil {
    log.Fatal(err)
}

switch {
case status.IsDelivered:
    fmt.Printf("✓ Message %s delivered at %s\n", status.MessageID, status.Time)
case status.Status == "PENDING":
    fmt.Printf("⏳ Message %s is pending delivery\n", status.MessageID)
case status.Status == "FAILED":
    fmt.Printf("✗ Message %s failed to deliver\n", status.MessageID)
default:
    fmt.Printf("ℹ️ Message %s status: %s\n", status.MessageID, status.Status)
}
```

### Bulk Delivery Status Check

```go
messageIDs := []string{"msg123", "msg456", "msg789"}
statuses, err := client.GetBulkDeliveryStatus(messageIDs)
if err != nil {
    log.Printf("Some status checks failed: %v", err)
}

for _, status := range statuses {
    fmt.Printf("Message %s: %s (Delivered: %v)\n",
        status.MessageID, status.Status, status.IsDelivered)
}
```

### Checking Balance

```go
balance, err := client.CheckBalance()
if err != nil {
    log.Fatal("Balance check failed:", err)
}

if balance.Success {
    fmt.Printf("Available credits: %d\n", balance.Credits)
    fmt.Printf("Balance: %.2f\n", balance.Balance)

    // Warn if balance is low
    if balance.Credits < 100 {
        fmt.Println("Warning: Low balance! Please recharge soon.")
    }
} else {
    fmt.Printf("Error checking balance: %s\n", balance.Error)
}
```

### Sending Unicode SMS (Arabic, Chinese, etc.)

```go
senderID := "YOURSMS"

// Arabic message
arabicMsg := "مرحبا بكم في خدمتنا"
response, err := client.SendUnicodeSMS(senderID, "1234567890", arabicMsg)

// Chinese message
chineseMsg := "欢迎使用我们的服务"
response, err = client.SendUnicodeSMS(senderID, "1234567890", chineseMsg)

if err == nil && response.Success {
    fmt.Println("Unicode SMS sent successfully")
}
```

### Sending XML SMS with Language Support

```go
senderID := "YOURSMS"
response, err := client.SendXMLSMS(
    senderID,
    "1234567890",
    "Your message here",
    "english", // Language parameter
)

if err != nil {
    log.Printf("XML SMS failed: %v", err)
} else {
    fmt.Printf("XML SMS response: %s\n", response)
}
```

### Handling Callbacks

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/hekimapro/mobishastra"
)

func handleInboundSMS(w http.ResponseWriter, r *http.Request) {
    var callback mobishastra.InboundSMSCallback

    if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Process incoming message
    fmt.Printf("Received from %s: %s\n", callback.PhoneNumber, callback.Message)
    fmt.Printf("Keyword: %s, Shortcode: %s\n", callback.Keyword, callback.ShortCode)

    // Auto-reply or process as needed
    w.WriteHeader(http.StatusOK)
}

func handleDeliveryReport(w http.ResponseWriter, r *http.Request) {
    var dlr mobishastra.DeliveryStatusCallback

    if err := json.NewDecoder(r.Body).Decode(&dlr); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    fmt.Printf("Delivery update - ID: %s, Status: %s\n", dlr.MessageID, dlr.Status)
    if dlr.ErrorCode != "" {
        fmt.Printf("Error code: %s\n", dlr.ErrorCode)
    }

    w.WriteHeader(http.StatusOK)
}

func main() {
    cfg := config.LoadConfig()

    http.HandleFunc("/sms/callback", handleInboundSMS)
    http.HandleFunc("/dlr/callback", handleDeliveryReport)

    port := cfg.CallbackPort
    if port == "" {
        port = "8080"
    }

    log.Printf("Starting callback server on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

### Custom HTTP Client Configuration

```go
// Create client with custom timeout
client := &mobishastra.SMSClient{
    Config: cfg,
    HTTPClient: &http.Client{
        Timeout: 60 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:    10,
            IdleConnTimeout: 30 * time.Second,
        },
    },
}
```

## Error Handling

The package includes predefined error codes mapping:

```go
var ErrorCodes = map[string]string{
    "000": "Send Successful",
    "001": "Invalid Receiver",
    "003": "Invalid Message",
    "005": "Authorization failed",
    "006": "DND Number",
    "007": "Cannot Extract Country Code",
    "008": "Empty Receiver",
    "009": "Profile Blocked",
    "010": "Invalid Profile ID",
    "011": "Profile ID expired",
    "012": "Sender Id more than 13 Chars",
    "013": "Server Error",
}
```

### Error Handling Example

```go
response, err := client.SendSMS(senderID, phoneNumber, message)
if err != nil {
    // Network or connection error
    log.Printf("Connection error: %v", err)
    return
}

if !response.Success {
    // API error
    if errorMsg, exists := mobishastra.ErrorCodes[response.ResponseCode]; exists {
        log.Printf("SMS failed: %s (Code: %s)", errorMsg, response.ResponseCode)
    } else {
        log.Printf("SMS failed: %s", response.Error)
    }

    // Handle specific error codes
    switch response.ResponseCode {
    case "005":
        log.Println("Authentication failed - check your credentials")
    case "006":
        log.Println("Number is registered on DND")
    case "009":
        log.Println("Profile is blocked - contact support")
    case "012":
        log.Println("Sender ID exceeds 13 characters")
    }
}
```

## Best Practices

### 1. Sender ID Management
- Sender ID must be approved by your SMS provider
- Maximum length is 13 characters
- Use different Sender IDs for different purposes (transactional, promotional, OTP)
- Store Sender IDs in configuration or database, not hardcoded

```go
// Good practice
const (
    SenderIDTransactional = "TXNSMS"
    SenderIDPromotional   = "PROMO"
    SenderIDOTP          = "OTPSMS"
)

response, err := client.SendSMS(SenderIDTransactional, phone, message)
```

### 2. Environment Variables
Always use environment variables for sensitive credentials. Never hardcode them.

### 3. Error Handling
Always check both the error return value and the `Success` field in responses.

### 4. Rate Limiting
Implement rate limiting to avoid overwhelming the API.

```go
// Simple rate limiter
limiter := time.Tick(1 * time.Second) // 1 SMS per second

for _, phone := range phoneNumbers {
    <-limiter
    go client.SendSMS(senderID, phone, message)
}
```

### 5. Connection Management
The client uses a 30-second timeout by default. Adjust based on your needs:

```go
client := &mobishastra.SMSClient{
    config: cfg,
    httpClient: &http.Client{
        Timeout: 60 * time.Second, // Increase for slow connections
    },
}
```

### 6. Phone Number Formatting
The client automatically formats phone numbers, but ensure numbers include country codes when needed.

```go
// The client will handle these formats:
client.formatMobileNumber("+1234567890")  // -> "1234567890"
client.formatMobileNumber("123-456-7890") // -> "1234567890"
client.formatMobileNumber("(123) 456-7890") // -> "1234567890"
```

## Testing

### Unit Test Example

```go
package mobishastra

import (
    "testing"
    "github.com/hekimapro/mobishastra"
)

func TestSendSMS(t *testing.T) {
    cfg := &mobishastra.Config{
        UserID:   "test_user",
        Password: "test_pass",
        CountryCode: "ALL",
        Priority: "High",
    }

    client := NewSMSClient(cfg)

    // Test with invalid phone number
    response, err := client.SendSMS("TEST", "", "test message")
    if err == nil && response.Success {
        t.Error("Expected error or failure for empty phone number")
    }
}

func TestFormatMobileNumber(t *testing.T) {
    client := &SMSClient{}

    tests := []struct{
        input string
        expected string
    }{
        {"+1234567890", "1234567890"},
        {"123-456-7890", "1234567890"},
        {"(123) 456-7890", "1234567890"},
        {"123 456 7890", "1234567890"},
    }

    for _, test := range tests {
        result := client.formatMobileNumber(test.input)
        if result != test.expected {
            t.Errorf("formatMobileNumber(%s) = %s, expected %s",
                test.input, result, test.expected)
        }
    }
}
```

## Migration Guide (from previous version)

If you were using the previous version where Sender ID was in config:

**Previous version:**
```go
cfg := config.LoadConfig() // included SenderID
client := mobishastra.NewSMSClient(cfg)
response, err := client.SendSMS("1234567890", "Hello")
```

**New version:**
```go
cfg := config.LoadConfig() // SenderID removed from config
client := mobishastra.NewSMSClient(cfg)
response, err := client.SendSMS("YOUR_SENDER_ID", "1234567890", "Hello")
```

## Contributing

Contributions are welcome! Please submit pull requests or create issues for bugs and feature requests.

## License

This project is licensed under the MIT License.

## Support

For API-specific issues, contact Mobishastra support. For package issues, open an issue on GitHub.

---

**Note:** This package is community-maintained and not officially affiliated with Mobishastra. Always test thoroughly before using in production.