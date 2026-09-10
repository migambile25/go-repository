package primepay

// credentials holds the gateway app ID, app secret, and base URL every
// request in this package signs against. Set once via Configure before
// calling Collect/DisburseToMobile/DisburseToBank/CheckStatus/Summarize/
// MobileNameLookup/BankNameLookup.
var credentials struct {
	appID     string
	appSecret string
	baseURL   string
}

// Configure sets the PayMeAfrica gateway credentials this package signs and
// routes requests with. Call once at startup, before any other function in
// this package. There is no other configuration entry point: this package
// no longer reads environment variables or a .env file for credentials.
func Configure(appID, appSecret, baseURL string) {
	credentials.appID = appID
	credentials.appSecret = appSecret
	credentials.baseURL = baseURL
}
