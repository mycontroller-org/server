package javascript_helper

const (
	KeyMcUtils = "mcUtils"

	keyCrypto     = "crypto"
	keyConvert    = "convert"
	keyHttpClient = "httpClient"
)

// includes helper functions
func GetHelperUtils() map[string]interface{} {
	return map[string]interface{}{
		keyCrypto:     &Crypto{},
		keyConvert:    &Convert{},
		keyHttpClient: &HttpClient{},
	}
}
