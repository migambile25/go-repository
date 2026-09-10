package mobishastra

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// API Credentials
	UserID   string
	Password string

	// API Endpoints
	SendSMSURL        string
	SendSMSCommaURL   string
	CheckBalanceURL   string
	DeliveryStatusURL string

	// Callback Settings
	CallbackSecret string
	CallbackPort   string

	// Default Settings
	CountryCode string
	Priority    string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		UserID:            getEnv("SMS_USER_ID", ""),
		Password:          getEnv("SMS_PASSWORD", ""),
		SendSMSURL:        getEnv("SMS_SEND_URL", "https://mshastra.com/sendurl.aspx"),
		SendSMSCommaURL:   getEnv("SMS_SEND_COMM_URL", "https://mshastra.com/sendurlcomma.aspx"),
		CheckBalanceURL:   getEnv("SMS_BALANCE_URL", "https://mshastra.com/balance.aspx"),
		DeliveryStatusURL: getEnv("SMS_DELIVERY_STATUS_URL", "http://login.telkosh.com/dlrstatus_api.aspx"),
		CallbackSecret:    getEnv("SMS_CALLBACK_SECRET", ""),
		CallbackPort:      getEnv("SMS_CALLBACK_PORT", ""),
		CountryCode:       getEnv("SMS_COUNTRY_CODE", "ALL"),
		Priority:          getEnv("SMS_PRIORITY", "High"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
