package mobishastra

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SMSClient struct {
	config     *Config
	httpClient *http.Client
}

func NewSMSClient(cfg *Config) *SMSClient {
	return &SMSClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// logRequest logs request payload in JSON format
func (client *SMSClient) logRequest(method, url string, body io.Reader) {
	logData := map[string]any{
		"method":    method,
		"url":       url,
		"type":      "request",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if body != nil {
		bodyBytes, _ := io.ReadAll(body)
		bodyStr := string(bodyBytes)

		// Try to parse as JSON if it looks like JSON
		var jsonBody any
		if strings.Contains(bodyStr, "{") || strings.Contains(bodyStr, "<") {
			// For XML or JSON, try to parse
			if strings.Contains(bodyStr, "{") {
				if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
					logData["body"] = jsonBody
				} else {
					logData["body"] = bodyStr
				}
			} else {
				logData["body"] = bodyStr
			}
		} else {
			logData["body"] = bodyStr
		}

		// Restore body for actual request
		if r, ok := body.(io.ReadCloser); ok {
			r.Close()
		}
	}

	jsonLog, _ := json.Marshal(logData)
	fmt.Println(string(jsonLog))
}

// logResponse logs response payload in JSON format
func (client *SMSClient) logResponse(resp *http.Response, body []byte) {
	bodyStr := string(body)

	// Try to parse response body as JSON
	var jsonBody any
	parseErr := json.Unmarshal(body, &jsonBody)

	logData := map[string]any{
		"status_code": resp.StatusCode,
		"type":        "response",
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if parseErr == nil && jsonBody != nil {
		// Successfully parsed as JSON
		logData["body"] = jsonBody
	} else if strings.Contains(bodyStr, "Send Successful") || strings.Contains(bodyStr, ",") {
		// Parse the response format from the SMS API
		parsedResponse := client.parseResponseToJSON(bodyStr)
		logData["body"] = parsedResponse
	} else {
		// Keep as string if not parseable
		logData["body"] = bodyStr
	}

	jsonLog, _ := json.Marshal(logData)
	fmt.Println(string(jsonLog))
}

// parseResponseToJSON converts SMS API response to JSON format
func (client *SMSClient) parseResponseToJSON(response string) any {
	response = strings.TrimSpace(response)

	if strings.Contains(response, "Send Successful") {
		return map[string]any{
			"status":  "success",
			"message": response,
			"code":    "000",
		}
	} else if strings.Contains(response, ",") {
		parts := strings.Split(response, ",")
		if len(parts) == 2 {
			return map[string]any{
				"message_id": strings.TrimSpace(parts[0]),
				"status":     strings.TrimSpace(parts[1]),
				"is_success": true,
			}
		} else if len(parts) == 3 {
			return map[string]any{
				"message_id":    strings.TrimSpace(parts[0]),
				"status":        strings.TrimSpace(parts[1]),
				"delivery_time": strings.TrimSpace(parts[2]),
				"is_delivered":  strings.Contains(strings.ToUpper(parts[1]), "DELIVERED"),
			}
		}
		return map[string]any{
			"raw_response": response,
			"parts":        parts,
		}
	} else if strings.Contains(strings.ToLower(response), "invalid") {
		return map[string]any{
			"status":  "error",
			"message": response,
			"code":    client.extractErrorCode(response),
		}
	}

	return map[string]any{
		"raw_response": response,
	}
}

// parseQueryParamsToJSON converts URL query parameters to JSON
func (client *SMSClient) parseQueryParamsToJSON(urlStr string) map[string]any {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}

	params := make(map[string]any)
	for key, values := range parsedURL.Query() {
		if len(values) == 1 {
			params[key] = values[0]
		} else {
			params[key] = values
		}
	}

	// Remove sensitive data
	delete(params, "pwd")
	if user, ok := params["user"]; ok {
		params["user"] = user
	}

	return params
}

// Authenticate checks if credentials are valid by checking balance
func (client *SMSClient) Authenticate() (*BalanceResponse, error) {
	return client.CheckBalance()
}

// CheckBalance returns current SMS credits balance
func (client *SMSClient) CheckBalance() (*BalanceResponse, error) {

	apiURL := fmt.Sprintf("%s?user=%s&pwd=%s", client.config.CheckBalanceURL, client.config.UserID, client.config.Password)

	// Log request with parsed params
	requestLog := map[string]any{
		"method":    "GET",
		"url":       apiURL,
		"type":      "request",
		"timestamp": time.Now().Format(time.RFC3339),
		"params":    client.parseQueryParamsToJSON(apiURL),
	}
	jsonLog, _ := json.Marshal(requestLog)
	fmt.Println(string(jsonLog))

	response, err := client.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("balance check failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	balanceString := strings.TrimSpace(string(body))

	// Try to parse balance as number
	balanceValue, parseErr := strconv.ParseFloat(balanceString, 64)

	// Log response
	responseLog := map[string]any{
		"status_code": response.StatusCode,
		"type":        "response",
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if parseErr == nil {
		responseLog["body"] = map[string]any{
			"balance": balanceValue,
			"credits": int(balanceValue),
		}
	} else {
		responseLog["body"] = map[string]any{
			"raw_response": balanceString,
			"is_error":     strings.Contains(balanceString, "Invalid"),
		}
	}

	jsonLog, _ = json.Marshal(responseLog)
	fmt.Println(string(jsonLog))

	// Try to parse as float
	balance, err := strconv.ParseFloat(balanceString, 64)
	if err != nil {
		// Check for error message
		if strings.Contains(balanceString, "Invalid") {
			return &BalanceResponse{
				Success: false,
				Error:   balanceString,
			}, nil
		}
		return nil, fmt.Errorf("unexpected balance response: %s", balanceString)
	}

	return &BalanceResponse{
		Success: true,
		Balance: balance,
		Credits: int(balance),
	}, nil
}

// SendSMS sends a single SMS
func (client *SMSClient) SendSMS(senderID, phoneNumber, message string) (*SingleSMSResponse, error) {
	return client.SendSMSWithOptions(senderID, phoneNumber, message, "", "")
}

// SendSMSWithOptions sends a single SMS with additional options
func (client *SMSClient) SendSMSWithOptions(senderID, phoneNumber, message, scheduledAt, showError string) (*SingleSMSResponse, error) {

	apiURL := fmt.Sprintf("%s?user=%s&pwd=%s&senderid=%s&mobileno=%s&msgtext=%s&CountryCode=%s",
		client.config.SendSMSURL,
		client.config.UserID,
		client.config.Password,
		url.QueryEscape(senderID),
		client.formatMobileNumber(phoneNumber),
		url.QueryEscape(message),
		client.config.CountryCode,
	)

	if client.config.Priority != "" {
		apiURL += fmt.Sprintf("&priority=%s", client.config.Priority)
	}

	if scheduledAt != "" {
		apiURL += fmt.Sprintf("&scheduledDate=%s", url.QueryEscape(scheduledAt))
	}

	if showError != "" {
		apiURL += fmt.Sprintf("&ShowError=%s", showError)
	}

	// Log request with parsed params
	requestLog := map[string]any{
		"method":    "GET",
		"url":       apiURL,
		"type":      "request",
		"timestamp": time.Now().Format(time.RFC3339),
		"params":    client.parseQueryParamsToJSON(apiURL),
	}
	jsonLog, _ := json.Marshal(requestLog)
	fmt.Println(string(jsonLog))

	resp, err := client.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("SMS request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	responseText := strings.TrimSpace(string(body))

	// Log response
	client.logResponse(resp, body)

	// Parse response
	smsResp := &SingleSMSResponse{
		ResponseText: responseText,
	}

	// Check for success
	if strings.Contains(responseText, "Send Successful") {
		smsResp.Success = true
		smsResp.ResponseCode = "000"
	} else if strings.Contains(responseText, ",") {
		// DLR format with message ID
		parts := strings.Split(responseText, ",")
		if len(parts) >= 1 {
			smsResp.MessageID = strings.TrimSpace(parts[0])
			if len(parts) >= 2 {
				smsResp.ResponseText = strings.TrimSpace(parts[1])
			}
		}
		smsResp.Success = true
	} else {
		smsResp.Success = false
		smsResp.Error = responseText
		smsResp.ResponseCode = client.extractErrorCode(responseText)
	}

	return smsResp, nil
}

// SendBulkSMS sends SMS to multiple numbers
func (client *SMSClient) SendBulkSMS(senderID string, phoneNumbers []string, message string) (*BulkSMSResponse, error) {
	return client.SendBulkSMSWithOptions(senderID, phoneNumbers, message, "", "")
}

// SendBulkSMSWithOptions sends bulk SMS with options
func (client *SMSClient) SendBulkSMSWithOptions(senderID string, phoneNumbers []string, message, scheduledAt, showError string) (*BulkSMSResponse, error) {

	if len(phoneNumbers) == 0 {
		return nil, fmt.Errorf("no mobile numbers provided")
	}

	// Format mobile numbers comma-separated
	formattedNumbers := make([]string, len(phoneNumbers))
	for i, num := range phoneNumbers {
		formattedNumbers[i] = client.formatMobileNumber(num)
	}
	mobileNoStr := strings.Join(formattedNumbers, ",")

	apiURL := fmt.Sprintf("%s?user=%s&pwd=%s&senderid=%s&mobileno=%s&msgtext=%s&CountryCode=%s",
		client.config.SendSMSCommaURL,
		client.config.UserID,
		client.config.Password,
		url.QueryEscape(senderID),
		mobileNoStr,
		url.QueryEscape(message),
		client.config.CountryCode,
	)

	if client.config.Priority != "" {
		apiURL += fmt.Sprintf("&priority=%s", client.config.Priority)
	}

	if scheduledAt != "" {
		apiURL += fmt.Sprintf("&scheduledDate=%s", url.QueryEscape(scheduledAt))
	}

	if showError != "" {
		apiURL += fmt.Sprintf("&ShowError=%s", showError)
	}

	// Log request with parsed params
	requestLog := map[string]any{
		"method":    "GET",
		"url":       apiURL,
		"type":      "request",
		"timestamp": time.Now().Format(time.RFC3339),
		"params":    client.parseQueryParamsToJSON(apiURL),
	}
	jsonLog, _ := json.Marshal(requestLog)
	fmt.Println(string(jsonLog))

	resp, err := client.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("bulk SMS request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response
	client.logResponse(resp, body)

	responseText := strings.TrimSpace(string(body))

	bulkResp := &BulkSMSResponse{
		Success:   true,
		Responses: make([]SingleSMSResponseData, 0),
	}

	// Parse comma-separated response (if multiple)
	// Format can be: "Send Successful" or multiple comma-separated responses
	if strings.Contains(responseText, "Send Successful") {
		for _, mobileNo := range phoneNumbers {
			bulkResp.Responses = append(bulkResp.Responses, SingleSMSResponseData{
				MobileNo:     mobileNo,
				ResponseText: responseText,
				ResponseCode: "000",
				Success:      true,
			})
		}
		bulkResp.TotalSent = len(phoneNumbers)
	} else if strings.Contains(responseText, ",") {
		// Handle comma-separated responses for each number
		parts := strings.Split(responseText, ",")
		for i, part := range parts {
			if i < len(phoneNumbers) {
				success := !strings.Contains(part, "Invalid") && !strings.Contains(part, "Failed")
				code := client.extractErrorCode(part)
				bulkResp.Responses = append(bulkResp.Responses, SingleSMSResponseData{
					MobileNo:     phoneNumbers[i],
					ResponseText: strings.TrimSpace(part),
					ResponseCode: code,
					Success:      success,
				})
				if !success {
					bulkResp.Success = false
				}
			}
		}
		bulkResp.TotalSent = len(bulkResp.Responses)
	} else {
		// Single response for all numbers
		success := !strings.Contains(responseText, "Invalid") &&
			!strings.Contains(responseText, "Failed") &&
			!strings.Contains(responseText, "Blocked")

		for _, mobileNo := range phoneNumbers {
			bulkResp.Responses = append(bulkResp.Responses, SingleSMSResponseData{
				MobileNo:     mobileNo,
				ResponseText: responseText,
				ResponseCode: client.extractErrorCode(responseText),
				Success:      success,
			})
		}
		bulkResp.Success = success
		bulkResp.TotalSent = len(phoneNumbers)
	}

	return bulkResp, nil
}

// SendUnicodeSMS sends Unicode SMS (for Arabic, etc.)
func (client *SMSClient) SendUnicodeSMS(senderID, phoneNumber, message string) (*SingleSMSResponse, error) {
	return client.SendSMSWithOptions(senderID, phoneNumber, message, "", "")
}

// GetDeliveryStatus checks delivery status of a message
func (client *SMSClient) GetDeliveryStatus(messageID string) (*DeliveryStatusResponse, error) {
	apiURL := fmt.Sprintf("%s?user=%s&pwd=%s&messageid=%s",
		client.config.DeliveryStatusURL,
		client.config.UserID,
		client.config.Password,
		messageID,
	)

	// Log request with parsed params
	requestLog := map[string]any{
		"method":    "GET",
		"url":       apiURL,
		"type":      "request",
		"timestamp": time.Now().Format(time.RFC3339),
		"params":    client.parseQueryParamsToJSON(apiURL),
	}
	jsonLog, _ := json.Marshal(requestLog)
	fmt.Println(string(jsonLog))

	resp, err := client.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("DLR status request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DLR response: %w", err)
	}

	// Log response
	client.logResponse(resp, body)

	responseStr := strings.TrimSpace(string(body))

	// Expected format: Message_id,DLR_Status,DLR_Time
	parts := strings.Split(responseStr, ",")

	if len(parts) < 3 {
		return &DeliveryStatusResponse{
			MessageID:   messageID,
			Status:      "UNKNOWN",
			IsDelivered: false,
		}, nil
	}

	dlrTime, _ := time.Parse("2006-01-02 15:04:05", parts[2])
	if dlrTime.IsZero() {
		dlrTime = time.Now()
	}

	status := strings.ToLower(strings.TrimSpace(parts[1]))

	if status == "delivrd" {
		status = "delivered"
	} else if status == "dlr_not_recv" {
		status = "pending"
	} else if strings.Contains(status, "undeliv") {
		status = "undelivered"
	} else if strings.Contains(status, "expired") {
		status = "expired"
	} else {
		status = "failed"
	}

	isDelivered := status == "delivered"

	return &DeliveryStatusResponse{
		MessageID:   strings.TrimSpace(parts[0]),
		Status:      status,
		Time:        dlrTime,
		IsDelivered: isDelivered,
	}, nil
}

// GetBulkDeliveryStatus checks delivery status for multiple messages
func (client *SMSClient) GetBulkDeliveryStatus(messageIDs []string) ([]*DeliveryStatusResponse, error) {
	results := make([]*DeliveryStatusResponse, 0, len(messageIDs))

	for _, msgID := range messageIDs {
		status, err := client.GetDeliveryStatus(msgID)
		if err != nil {
			// Continue with other IDs even if one fails
			results = append(results, &DeliveryStatusResponse{
				MessageID:   msgID,
				Status:      "ERROR",
				IsDelivered: false,
			})
			continue
		}
		results = append(results, status)
	}

	return results, nil
}

// SendXMLSMS sends SMS using XML API
func (client *SMSClient) SendXMLSMS(senderID, phoneNumber, message, language string) (string, error) {
	xmlBody := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<request>
    <user>%s</user>
    <pwd>%s</pwd>
    <messages>
        <message>
            <number>%s</number>
            <msg>%s</msg>
            <sender>%s</sender>
            <language>%s</language>
        </message>
    </messages>
</request>`,
		client.config.UserID,
		client.config.Password,
		client.formatMobileNumber(phoneNumber),
		message,
		senderID,
		language,
	)

	// Log request with XML body
	requestLog := map[string]any{
		"method":    "POST",
		"url":       "https://mshastra.com/sendsms_api_xml.aspx",
		"type":      "request",
		"timestamp": time.Now().Format(time.RFC3339),
		"headers": map[string]string{
			"Content-Type": "application/xml",
		},
		"body": xmlBody, // XML as string since it's not JSON
	}
	jsonLog, _ := json.Marshal(requestLog)
	fmt.Println(string(jsonLog))

	req, err := http.NewRequest("POST", "https://mshastra.com/sendsms_api_xml.aspx", strings.NewReader(xmlBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/xml")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Log response
	responseLog := map[string]any{
		"status_code": resp.StatusCode,
		"type":        "response",
		"timestamp":   time.Now().Format(time.RFC3339),
		"body":        string(body),
	}
	jsonLog, _ = json.Marshal(responseLog)
	fmt.Println(string(jsonLog))

	return string(body), nil
}

func (client *SMSClient) formatMobileNumber(phoneNumber string) string {
	// Remove any spaces or special characters
	phoneNumber = strings.TrimSpace(phoneNumber)
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")

	return phoneNumber
}

func (client *SMSClient) extractErrorCode(response string) string {
	for code, msg := range ErrorCodes {
		if strings.Contains(response, msg) || strings.Contains(response, code) {
			return code
		}
	}
	return "999"
}
