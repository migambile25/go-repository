package primepay

import "encoding/json"

func Collect(payload *CollectionRequestPayload) (*CollectionResponsePayload, error) {

	path := "/api/v1/transact"
	payload.Action = "collection"
	payload.CustomerPhoneNumber = normalizePhoneNumber(payload.CustomerPhoneNumber)

	responseByte, err := post(path, payload)
	if err != nil {
		return nil, err
	}

	var responseBody CollectionResponsePayload
	if err := json.Unmarshal(responseByte, &responseBody); err != nil {
		return nil, err
	}

	return &responseBody, nil
}

func DisburseToMobile(payload *MobileDisbursementRequestPayload) (*DisbursementResponsePayload, error) {

	path := "/api/v1/transact"
	payload.Action = "disbursement"
	payload.CustomerPhoneNumber = normalizePhoneNumber(payload.CustomerPhoneNumber)

	responseByte, err := post(path, payload)
	if err != nil {
		return nil, err
	}

	var responseBody DisbursementResponsePayload
	if err := json.Unmarshal(responseByte, &responseBody); err != nil {
		return nil, err
	}

	return &responseBody, nil
}

func CheckStatus(reference string) (*StatusResponsePayload, error) {

	path := "/api/v1/query"
	payload := &statusRequestPayload{Reference: reference}

	responseByte, err := post(path, payload)
	if err != nil {
		return nil, err
	}

	var responseBody StatusResponsePayload
	if err := json.Unmarshal(responseByte, &responseBody); err != nil {
		return nil, err
	}

	return &responseBody, nil
}

func DisburseToBank(payload *BankDisbursementRequestPayload) (*DisbursementResponsePayload, error) {

	payload.Channel = "BANK"
	path := "/api/v1/transact"
	payload.Action = "disbursement"
	payload.SenderPhoneNumber = normalizePhoneNumber(payload.SenderPhoneNumber)

	responseByte, err := post(path, payload)
	if err != nil {
		return nil, err
	}

	var responseBody DisbursementResponsePayload
	if err := json.Unmarshal(responseByte, &responseBody); err != nil {
		return nil, err
	}

	return &responseBody, nil
}

func MobileNameLookup(payload *MobileNameLookupRequestPayload) (*NameLookupResponsePayload, error) {

	payload.Type = "wallet"
	path := "/api/v1/lookup"

	responseByte, err := post(path, payload)
	if err != nil {
		return nil, err
	}

	var responseBody NameLookupResponsePayload
	if err := json.Unmarshal(responseByte, &responseBody); err != nil {
		return nil, err
	}

	return &responseBody, nil
}

func BankNameLookup(payload *BankNameLookupRequestPayload) (*NameLookupResponsePayload, error) {

	payload.Type = "bank"
	path := "/api/v1/lookup"

	responseByte, err := post(path, payload)
	if err != nil {
		return nil, err
	}

	var responseBody NameLookupResponsePayload
	if err := json.Unmarshal(responseByte, &responseBody); err != nil {
		return nil, err
	}

	return &responseBody, nil
}

func Summarize() (*SummaryResponsePayload, error) {

	path := "/api/v1/summarize"

	responseByte, err := post(path, nil)
	if err != nil {
		return nil, err
	}

	var responseBody SummaryResponsePayload
	if err := json.Unmarshal(responseByte, &responseBody); err != nil {
		return nil, err
	}

	return &responseBody, nil
}
