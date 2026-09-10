package primepay

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func generateHmacSignature(rawRequestBody []byte, unixTimestamp int64, secretKey string) string {
	timestampBytes := []byte(strconv.FormatInt(unixTimestamp, 10))
	stringToSign := append(rawRequestBody, timestampBytes...)

	hmacHasher := hmac.New(sha256.New, []byte(secretKey))
	hmacHasher.Write(stringToSign)

	return base64.StdEncoding.EncodeToString(hmacHasher.Sum(nil))
}

func buildAuthenticationHeaders(rawRequestBody []byte) http.Header {
	unixTimestamp := time.Now().Unix()
	signature := generateHmacSignature(rawRequestBody, unixTimestamp, credentials.appSecret)

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("X-App-ID", credentials.appID)
	headers.Set("X-Timestamp", strconv.FormatInt(unixTimestamp, 10))
	headers.Set("X-Signature", signature)

	return headers
}

func post(path string, requestPayload any) ([]byte, error) {

	rawRequestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, err
	}

	url := credentials.baseURL + path
	httpRequest, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(rawRequestBody))
	if err != nil {
		return nil, err
	}

	httpRequest.Header = buildAuthenticationHeaders(rawRequestBody)

	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("POST %s response (%d): %s", url, httpResponse.StatusCode, string(responseBody))

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return nil, fmt.Errorf("POST %s failed with status %d", url, httpResponse.StatusCode)
	}

	return responseBody, nil
}
