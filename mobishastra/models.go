package mobishastra

import (
	"time"
)

// SMS Request Models
type SendBulkSMSRequest struct {
	User         string   `json:"user"`
	Password     string   `json:"pwd"`
	SenderID     string   `json:"senderid"`
	PhoneNumbers []string `json:"mobileno"`
	MessageText  string   `json:"msgtext"`
	Priority     string   `json:"priority,omitempty"`
	CountryCode  string   `json:"countrycode,omitempty"`
	ScheduledAt  string   `json:"scheduleddate,omitempty"`
	ShowError    string   `json:"showerror,omitempty"`
}

type SendSingleSMSRequest struct {
	User        string `json:"user"`
	Password    string `json:"pwd"`
	SenderID    string `json:"senderid"`
	PhoneNumber string `json:"mobileno"`
	MessageText string `json:"msgtext"`
	Priority    string `json:"priority,omitempty"`
	CountryCode string `json:"countrycode,omitempty"`
	ScheduledAt string `json:"scheduleddate,omitempty"`
}

// SMS Response Models
type SingleSMSResponse struct {
	Success      bool   `json:"success"`
	MessageID    string `json:"message_id,omitempty"`
	ResponseCode string `json:"response_code,omitempty"`
	ResponseText string `json:"response_text,omitempty"`
	Error        string `json:"error,omitempty"`
}

type BulkSMSResponse struct {
	Success   bool                    `json:"success"`
	Responses []SingleSMSResponseData `json:"responses"`
	TotalSent int                     `json:"total_sent"`
}

type SingleSMSResponseData struct {
	MobileNo     string `json:"mobileno"`
	MessageID    string `json:"message_id"`
	ResponseCode string `json:"response_code"`
	ResponseText string `json:"response_text"`
	Success      bool   `json:"success"`
}

// DLR Status Models
type DeliveryStatusRequest struct {
	User      string `json:"user"`
	Password  string `json:"pwd"`
	MessageID string `json:"messageid"`
}

type DeliveryStatusResponse struct {
	MessageID   string    `json:"message_id"`
	Status      string    `json:"dlr_status"`
	Time        time.Time `json:"dlr_time"`
	IsDelivered bool      `json:"is_delivered"`
}

// Callback Models (Inbound SMS)
type InboundSMSCallback struct {
	ShortCode   string    `json:"shortcode" form:"shortcode"`
	PhoneNumber string    `json:"mobileno" form:"mobileno"`
	Keyword     string    `json:"keyword" form:"keyword"`
	Message     string    `json:"message" form:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// DLR Callback Model
type DeliveryStatusCallback struct {
	MessageID   string    `json:"message_id"`
	Status      string    `json:"dlr_status"`
	Time        time.Time `json:"dlr_time"`
	PhoneNumber string    `json:"mobileno,omitempty"`
	ErrorCode   string    `json:"error_code,omitempty"`
}

// Balance Response
type BalanceResponse struct {
	Success bool    `json:"success"`
	Balance float64 `json:"balance,omitempty"`
	Credits int     `json:"credits,omitempty"`
	Error   string  `json:"error,omitempty"`
}

// Error Codes Mapping
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
