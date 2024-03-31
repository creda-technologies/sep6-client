package sep6_client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/strkey"
)

func RemoveSlashFromUrl(url string) string {
	asRunes := []rune(url)
	stringLen := len(asRunes)
	if asRunes[stringLen-1] == '/' {
		asRunes := asRunes[0 : stringLen-1]
		return string(asRunes)
	}
	return url
}

// sendRequest sends an HTTP request and decodes the JSON response.
func sendRequest(req *http.Request, v interface{}) error {
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return &FetchError{Field: "Network", Message: fmt.Sprintf("Sending request failed: %v", err)}
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		return &FetchError{
			Field:   "Rate Limit",
			Message: "service has rate limited you, try again later",
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &FetchError{Field: "HTTP Response", Message: fmt.Sprintf("Received non-200 response: %s", string(body))}
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return &FetchError{Field: "JSON Decode", Message: fmt.Sprintf("Decoding response failed: %v", err)}
	}
	return nil
}

// buildURL constructs a URL with the given base URL, path, and query parameters.
func buildURL(baseURL string, path string, queryParams url.Values) string {
	if queryParams != nil {
		return fmt.Sprintf("%s/%s?%s", baseURL, path, queryParams.Encode())
	}
	return fmt.Sprintf("%s/%s", baseURL, path)
}

// validateNotEmpty checks if the provided strings are not empty. Returns an error if any of them are empty.
func validateNotEmpty(fields map[string]string) error {
	for fieldName, fieldValue := range fields {
		if fieldValue == "" {
			return &ValidationError{Field: fieldName, Message: "cannot be empty"}
		}
	}
	return nil
}

// validateAmount checks if the amount is positive.
func validateAmount(amount float32) error {
	if amount <= 0 {
		return &ValidationError{Field: "amount", Message: "must be positive"}
	}
	return nil
}

// sanitizeURL removes the trailing slash from a URL if it exists.
func sanitizeURL(urlStr string) string {
	if urlStr[len(urlStr)-1] == '/' {
		return urlStr[:len(urlStr)-1]
	}
	return urlStr
}

func fetchSigningKey(serverURL string) (*string, error) {
	tomlURL := fmt.Sprintf("%s/.well-known/stellar.toml", serverURL)
	resp, err := http.Get(tomlURL)
	if err != nil {
		return nil, &FetchError{Field: "URL", Message: "Failed to fetch stellar.toml - " + err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &FetchError{Field: "StatusCode", Message: fmt.Sprintf("Failed to fetch stellar.toml - StatusCode: %d", resp.StatusCode)}
	}

	var config StellarToml

	contentType := resp.Header.Get("Content-Type")
	if contentType == "application/octet-stream" {
		fileData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, &FetchError{Field: "TOML Decode", Message: "Failed to decode stellar.toml - " + err.Error()}
		}
		toml.Decode(string(fileData), &config)
	} else {
		_, err = toml.NewDecoder(resp.Body).Decode(&config)
		if err != nil {
			return nil, &FetchError{Field: "TOML Decode", Message: "Failed to decode stellar.toml - " + err.Error()}
		}
	}
	return &config.SigningKey, nil
}

func StellarPublicKeyToEd25519(stellarPublicKey string) ([]byte, error) {
	// Decode the StrKey encoded Stellar public key
	decoded, err := strkey.Decode(strkey.VersionByteAccountID, stellarPublicKey)
	if err != nil {
		return nil, &ValidationError{Field: "StellarPublicKey", Message: fmt.Sprintf("failed to decode Stellar public key: %v", err)}
	}
	return decoded, nil
}

func verifySignatureFromString(signatureHeader, requestBody, stellarPublicKey, walletHost string, maxTimeDiff time.Duration) error {
	if signatureHeader == "" {
		return &ValidationError{Field: "SignatureHeader", Message: "no signature header provided"}
	}

	// Extract timestamp and base64 signature from the header
	var timestampStr, base64Sig string
	for _, part := range strings.Split(signatureHeader, ",") {
		if strings.HasPrefix(part, "t=") {
			timestampStr = strings.TrimPrefix(part, "t=")
		}
		if strings.HasPrefix(part, " s=") {
			base64Sig = strings.TrimPrefix(part, " s=")
		}
	}
	if timestampStr == "" || base64Sig == "" {
		return &ValidationError{Field: "SignatureHeader", Message: "signature header is malformed"}
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return &ValidationError{Field: "Timestamp", Message: "invalid timestamp in signature header"}
	}
	if time.Duration(time.Since(time.Unix(timestamp, 0)).Minutes()) > maxTimeDiff {
		return &ValidationError{Field: "Timestamp", Message: "request is not fresh"}
	}

	// Prepare payload for verification
	payload := strings.Join([]string{timestampStr, ".", walletHost, ".", requestBody}, "")

	// Decode base64 signature
	sigBytes, err := base64.StdEncoding.DecodeString(base64Sig)
	if err != nil {
		return &ValidationError{Field: "Signature", Message: "failed to decode base64 signature"}
	}

	// Verify signature
	kp, err := keypair.Parse(stellarPublicKey)
	if err != nil {
		return &ValidationError{Field: "StellarPublicKey", Message: "invalid stellar public key"}
	}

	verifyErr := kp.Verify([]byte(payload), sigBytes)
	if verifyErr != nil {
		return &ValidationError{Field: "SignatureVerification", Message: fmt.Sprintf("signature verification failed - %s", verifyErr.Error())}
	}

	return nil
}
