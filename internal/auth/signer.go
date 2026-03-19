// Package auth implements the APsystems OpenAPI signature authentication.
//
// Every API request must carry five custom headers:
//
//	X-CA-AppId             – the 32-character application identifier
//	X-CA-Timestamp         – Unix timestamp in milliseconds
//	X-CA-Nonce             – a unique 32-character UUID (hex, no dashes)
//	X-CA-Signature-Method  – "HmacSHA256" (preferred) or "HmacSHA1"
//	X-CA-Signature         – Base64(HMAC(stringToSign, appSecret))
//
// The string-to-sign is built as:
//
//	timestamp/nonce/appId/requestPath/HTTPMethod/signatureMethod
//
// where requestPath is the LAST segment of the URL path (e.g. for
// /user/api/v2/systems/details/ABC123 the segment is "ABC123").
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"
)

const (
	HeaderAppID           = "X-CA-AppId"
	HeaderTimestamp       = "X-CA-Timestamp"
	HeaderNonce           = "X-CA-Nonce"
	HeaderSignatureMethod = "X-CA-Signature-Method"
	HeaderSignature       = "X-CA-Signature"

	SignatureMethodSHA256 = "HmacSHA256"
)

// generateNonce produces a 32-character lowercase hex string (equivalent
// to a UUID without dashes).
func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// computeSignature builds the string-to-sign and produces the
// Base64-encoded HMAC-SHA256 signature.
//
//	stringToSign = timestamp/nonce/appId/requestPath/HTTPMethod/signatureMethod
func computeSignature(timestamp, nonce, appID, requestPath, method, sigMethod, appSecret string) (string, error) {
	stringToSign := fmt.Sprintf("%s/%s/%s/%s/%s/%s",
		timestamp, nonce, appID, requestPath, method, sigMethod,
	)

	mac := hmac.New(sha256.New, []byte(appSecret))
	if _, err := mac.Write([]byte(stringToSign)); err != nil {
		return "", fmt.Errorf("hmac write: %w", err)
	}

	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// SignRequest injects all required authentication headers into req.
// It is safe to call from concurrent goroutines.
func SignRequest(req *http.Request, appID, appSecret string) error {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	nonce, err := generateNonce()
	if err != nil {
		return err
	}

	// The API documentation states that RequestPath is "the last segment
	// of the path". For a URL like /user/api/v2/systems/details/XYZ
	// the last segment is "XYZ".
	requestPath := path.Base(req.URL.Path)

	signature, err := computeSignature(
		timestamp, nonce, appID, requestPath,
		req.Method, SignatureMethodSHA256, appSecret,
	)
	if err != nil {
		return fmt.Errorf("compute signature: %w", err)
	}

	req.Header.Set(HeaderAppID, appID)
	req.Header.Set(HeaderTimestamp, timestamp)
	req.Header.Set(HeaderNonce, nonce)
	req.Header.Set(HeaderSignatureMethod, SignatureMethodSHA256)
	req.Header.Set(HeaderSignature, signature)

	return nil
}
