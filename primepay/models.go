package primepay

import (
	"encoding/json"
	"time"
)

type CollectionRequestPayload struct {
	Action              string  `json:"action"`
	Amount              float64 `json:"amount"`
	CustomerPhoneNumber string  `json:"msisdn"`
	Reference           string  `json:"reference"`
	CallbackURL         string  `json:"callback_url"`
}

type CollectionResponsePayload struct {
	Status           string `json:"status"`
	TransactionID    string `json:"transaction_id"`
	PaymentStatus    string `json:"payment_status"`
	ProviderResponse struct {
		Result     string `json:"result"`
		ResultCode string `json:"resultcode"`
		Message    string `json:"message"`
	} `json:"provider_response"`
}

type MobileDisbursementRequestPayload struct {
	Action              string  `json:"action"`
	Amount              float64 `json:"amount"`
	CustomerPhoneNumber string  `json:"msisdn"`
	Channel             string  `json:"channel"`
	Reference           string  `json:"reference"`
}

type BankDisbursementRequestPayload struct {
	Action            string  `json:"action"`
	Amount            float64 `json:"amount"`
	SenderPhoneNumber string  `json:"msisdn"`
	Channel           string  `json:"channel"`
	Reference         string  `json:"reference"`
	BankCode          string  `json:"recipient_bank_code"`
	AccountNumber     string  `json:"recipient_account"`
	AccountName       string  `json:"recipient_name"`
	Remarks           string  `json:"remarks"`
}

type DisbursementResponsePayload struct {
	Status           string          `json:"status"`
	TransactionID    string          `json:"transaction_id"`
	PaymentStatus    string          `json:"payment_status"`
	ProviderResponse json.RawMessage `json:"provider_response"`
	Financials       struct {
		SystemProfit    float64 `json:"system_profit"`
		ProviderFee     float64 `json:"provider_fee"`
		TotalDeductible float64 `json:"total_deductible"`
	} `json:"financials"`
}

type statusRequestPayload struct {
	Reference string `json:"reference"`
}

type StatusResponsePayload struct {
	Currency      string    `json:"currency"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	Reference     string    `json:"reference"`
	PaymentStatus string    `json:"payment_status"`
	CreatedTime   time.Time `json:"created_time"`
}

type CallbackResponsePayload struct {
	TranasctionID       string  `json:"transid"`
	Reference           string  `json:"reference"`
	Result              string  `json:"result"`
	Amount              float64 `json:"amount"`
	CustomerPhoneNumber string  `json:"msisdn"`
	PaymentStatus       string  `json:"payment_status"`
}

type MobileNameLookupRequestPayload struct {
	Type                string `json:"type"`
	CustomerPhoneNumber string `json:"msisdn"`
	Channel             string `json:"channel"`
}

type BankNameLookupRequestPayload struct {
	Type          string `json:"type"`
	BankCode      string `json:"bank_code"`
	AccountNumber string `json:"account_number"`
}

type NameLookupResponsePayload struct {
	Status       string          `json:"status"`
	Name         string          `json:"name"`
	OriginalData json.RawMessage `json:"original_data"`
}

type SummaryResponsePayload struct {
	Status string `json:"status"`
	Data   struct {
		TotalVolume    float64 `json:"total_volume"`
		TotalCount     int     `json:"total_count"`
		SuccessCount   int     `json:"success_count"`
		FailedCount    int     `json:"failed_count"`
		PendingCount   int     `json:"pending_count"`
		SuccessRate    float64 `json:"success_rate"`
		AccountBalance float64 `json:"account_balance"`
	} `json:"data"`
}
