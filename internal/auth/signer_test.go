package auth

import (
	"net/http"
	"testing"
)

func TestSignRequest_SetsAllHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.apsystemsema.com:9282/user/api/v2/systems/details/ABC123", nil)

	err := SignRequest(req, "testappid12345678901234567890ab", "testsecret12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, h := range []string{HeaderAppID, HeaderTimestamp, HeaderNonce, HeaderSignatureMethod, HeaderSignature} {
		if req.Header.Get(h) == "" {
			t.Errorf("header %s is empty", h)
		}
	}

	if req.Header.Get(HeaderSignatureMethod) != SignatureMethodSHA256 {
		t.Errorf("expected signature method %s, got %s", SignatureMethodSHA256, req.Header.Get(HeaderSignatureMethod))
	}

	nonce := req.Header.Get(HeaderNonce)
	if len(nonce) != 32 {
		t.Errorf("nonce should be 32 hex chars, got %d: %s", len(nonce), nonce)
	}
}

func TestComputeSignature_Deterministic(t *testing.T) {
	sig1, _ := computeSignature("1696665600000", "aabbccdd11223344aabbccdd11223344", "myappid", "details", "GET", SignatureMethodSHA256, "secret123456")
	sig2, _ := computeSignature("1696665600000", "aabbccdd11223344aabbccdd11223344", "myappid", "details", "GET", SignatureMethodSHA256, "secret123456")

	if sig1 != sig2 {
		t.Error("same inputs should produce same signature")
	}

	sig3, _ := computeSignature("1696665600001", "aabbccdd11223344aabbccdd11223344", "myappid", "details", "GET", SignatureMethodSHA256, "secret123456")
	if sig1 == sig3 {
		t.Error("different timestamp should produce different signature")
	}
}
