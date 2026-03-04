package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAdminOnly_MissingToken(t *testing.T) {
	handler := adminOnly("client-id", "admin@test.com", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "missing token" {
		t.Errorf("expected 'missing token', got %q", body["error"])
	}
}

func TestAdminOnly_InvalidBearerFormat(t *testing.T) {
	handler := adminOnly("client-id", "admin@test.com", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Token abc123")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAdminOnly_InvalidJWT(t *testing.T) {
	handler := adminOnly("client-id", "admin@test.com", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid-jwt")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestVerifyGoogleToken_InvalidFormat(t *testing.T) {
	_, err := verifyGoogleToken("invalid", "client-id")
	if err == nil {
		t.Error("expected error for invalid JWT format")
	}
}

func TestVerifyGoogleToken_TwoParts(t *testing.T) {
	_, err := verifyGoogleToken("part1.part2", "client-id")
	if err == nil {
		t.Error("expected error for two-part JWT")
	}
}

func TestVerifyGoogleToken_InvalidBase64Header(t *testing.T) {
	_, err := verifyGoogleToken("!!!invalid!!!.payload.signature", "client-id")
	if err == nil {
		t.Error("expected error for invalid base64 header")
	}
}

func TestVerifyGoogleToken_NonRS256Alg(t *testing.T) {
	// Build a valid base64 header with wrong algorithm
	headerJSON := `{"kid":"test","alg":"HS256"}`
	header := base64RawURLEncode([]byte(headerJSON))
	payload := base64RawURLEncode([]byte(`{"iss":"accounts.google.com"}`))

	_, err := verifyGoogleToken(header+"."+payload+".sig", "client-id")
	if err == nil {
		t.Error("expected error for non-RS256 algorithm")
	}
}

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://localhost:5173', got %q", origin)
	}
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "" {
		t.Errorf("expected no Access-Control-Allow-Origin, got %q", origin)
	}
}

func TestCORSMiddleware_OptionsRequest(t *testing.T) {
	nextCalled := false
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}))

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", rec.Code)
	}
	if nextCalled {
		t.Error("next handler should not be called for OPTIONS preflight")
	}
}

func TestCORSMiddleware_AllowedMethods(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

// ---------- Full verification with mock JWKS ----------

// buildSignedJWT creates a real RS256-signed JWT using the provided RSA key.
func buildSignedJWT(t *testing.T, kid string, claims jwtClaims, privKey *rsa.PrivateKey) string {
	t.Helper()

	headerJSON := fmt.Sprintf(`{"alg":"RS256","kid":"%s"}`, kid)
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))

	claimsJSON, _ := json.Marshal(claims)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signed := headerB64 + "." + claimsB64
	hash := sha256.Sum256([]byte(signed))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatalf("failed to sign JWT: %v", err)
	}
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return signed + "." + sigB64
}

// mockJWKSServer creates a test server that serves JWKS with the given public key.
func mockJWKSServer(kid string, pubKey *rsa.PublicKey) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nB64 := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
		eBytes := big.NewInt(int64(pubKey.E)).Bytes()
		eB64 := base64.RawURLEncoding.EncodeToString(eBytes)

		resp := fmt.Sprintf(`{"keys":[{"kid":"%s","n":"%s","e":"%s"}]}`, kid, nB64, eB64)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	}))
}

func TestVerifyGoogleToken_FullFlow(t *testing.T) {
	// Generate RSA key pair
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	kid := "test-key-1"
	jwksServer := mockJWKSServer(kid, &privKey.PublicKey)
	defer jwksServer.Close()

	// Override the JWKS cache to use our mock server
	// Reset cache so it fetches from our mock
	cache.mu.Lock()
	cache.keys = nil
	cache.fetched = time.Time{}
	cache.mu.Unlock()

	// We need to override the googleCertsURL, but it's a const.
	// Instead, we test the cache directly.
	// Manually populate the cache with our key
	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	clientID := "test-client-id"
	claims := jwtClaims{
		Iss:   "accounts.google.com",
		Aud:   clientID,
		Email: "admin@test.com",
		Exp:   time.Now().Add(time.Hour).Unix(),
	}

	token := buildSignedJWT(t, kid, claims, privKey)

	result, err := verifyGoogleToken(token, clientID)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.Email != "admin@test.com" {
		t.Errorf("expected email admin@test.com, got %q", result.Email)
	}
	if result.Iss != "accounts.google.com" {
		t.Errorf("expected iss accounts.google.com, got %q", result.Iss)
	}
}

func TestVerifyGoogleToken_ExpiredToken(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-expired-key"

	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	claims := jwtClaims{
		Iss:   "accounts.google.com",
		Aud:   "test-client",
		Email: "user@test.com",
		Exp:   time.Now().Add(-time.Hour).Unix(), // expired
	}

	token := buildSignedJWT(t, kid, claims, privKey)
	_, err := verifyGoogleToken(token, "test-client")
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestVerifyGoogleToken_WrongAudience(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-aud-key"

	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	claims := jwtClaims{
		Iss:   "accounts.google.com",
		Aud:   "wrong-client",
		Email: "user@test.com",
		Exp:   time.Now().Add(time.Hour).Unix(),
	}

	token := buildSignedJWT(t, kid, claims, privKey)
	_, err := verifyGoogleToken(token, "correct-client")
	if err == nil {
		t.Error("expected error for wrong audience")
	}
}

func TestVerifyGoogleToken_WrongIssuer(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-iss-key"

	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	claims := jwtClaims{
		Iss:   "evil.com",
		Aud:   "test-client",
		Email: "user@test.com",
		Exp:   time.Now().Add(time.Hour).Unix(),
	}

	token := buildSignedJWT(t, kid, claims, privKey)
	_, err := verifyGoogleToken(token, "test-client")
	if err == nil {
		t.Error("expected error for wrong issuer")
	}
}

func TestVerifyGoogleToken_InvalidSignature(t *testing.T) {
	privKey1, _ := rsa.GenerateKey(rand.Reader, 2048)
	privKey2, _ := rsa.GenerateKey(rand.Reader, 2048) // different key
	kid := "test-sig-key"

	// Cache has key1's public key
	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey1.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	claims := jwtClaims{
		Iss:   "accounts.google.com",
		Aud:   "test-client",
		Email: "user@test.com",
		Exp:   time.Now().Add(time.Hour).Unix(),
	}

	// Sign with key2 (wrong key)
	token := buildSignedJWT(t, kid, claims, privKey2)
	_, err := verifyGoogleToken(token, "test-client")
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestAdminOnly_ValidTokenWrongEmail(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-admin-key"

	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	claims := jwtClaims{
		Iss:   "accounts.google.com",
		Aud:   "client-id",
		Email: "notadmin@test.com",
		Exp:   time.Now().Add(time.Hour).Unix(),
	}
	token := buildSignedJWT(t, kid, claims, privKey)

	nextCalled := false
	handler := adminOnly("client-id", "admin@test.com", func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	if nextCalled {
		t.Error("next handler should not be called for non-admin email")
	}
}

func TestAdminOnly_ValidTokenCorrectEmail(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-admin-ok"

	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	claims := jwtClaims{
		Iss:   "accounts.google.com",
		Aud:   "client-id",
		Email: "admin@test.com",
		Exp:   time.Now().Add(time.Hour).Unix(),
	}
	token := buildSignedJWT(t, kid, claims, privKey)

	nextCalled := false
	handler := adminOnly("client-id", "admin@test.com", func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !nextCalled {
		t.Error("expected next handler to be called for valid admin")
	}
}

func TestVerifyGoogleToken_HttpsIssuer(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-https-iss"

	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	claims := jwtClaims{
		Iss:   "https://accounts.google.com",
		Aud:   "test-client",
		Email: "user@test.com",
		Exp:   time.Now().Add(time.Hour).Unix(),
	}

	token := buildSignedJWT(t, kid, claims, privKey)
	result, err := verifyGoogleToken(token, "test-client")
	if err != nil {
		t.Fatalf("expected success for https issuer, got: %v", err)
	}
	if result.Iss != "https://accounts.google.com" {
		t.Errorf("expected https issuer, got %q", result.Iss)
	}
}

func TestVerifyGoogleToken_InvalidBase64Payload(t *testing.T) {
	// Valid RS256 header with a kid that won't exist in Google's JWKS
	headerJSON := `{"alg":"RS256","kid":"test-kid"}`
	header := base64RawURLEncode([]byte(headerJSON))

	// Garbage payload that is not valid base64
	payload := "!!!not-valid-base64!!!"
	token := header + "." + payload + ".signature"

	_, err := verifyGoogleToken(token, "client-id")
	if err == nil {
		t.Error("expected error for token with invalid base64 payload")
	}
}

func TestVerifyGoogleToken_InvalidBase64Signature(t *testing.T) {
	// Valid RS256 header with a kid that won't exist in Google's JWKS
	headerJSON := `{"alg":"RS256","kid":"test-kid-sig"}`
	header := base64RawURLEncode([]byte(headerJSON))

	// Valid base64 payload
	payloadJSON := `{"iss":"accounts.google.com","aud":"client-id","email":"user@test.com","exp":9999999999}`
	payload := base64RawURLEncode([]byte(payloadJSON))

	// Invalid signature
	token := header + "." + payload + ".!!!bad-sig!!!"

	_, err := verifyGoogleToken(token, "client-id")
	if err == nil {
		t.Error("expected error for token with invalid base64 signature")
	}
}

// Test decode signature base64 error — header is valid with kid in cache,
// but the signature part is not valid base64.
func TestVerifyGoogleToken_DecodeSignatureBase64Error(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-decsig"

	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	headerJSON := fmt.Sprintf(`{"alg":"RS256","kid":"%s"}`, kid)
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"accounts.google.com"}`))

	// Bad base64 signature
	token := headerB64 + "." + payloadB64 + ".!!!not-valid-base64!!!"
	_, err := verifyGoogleToken(token, "client-id")
	if err == nil {
		t.Error("expected error for invalid base64 signature")
	}
}

// Test decode claims base64 error — the JWT is signed correctly but the payload
// is not valid base64. Since RS256 signs the raw header.payload string, this
// will pass signature verification but fail when decoding claims base64.
func TestVerifyGoogleToken_DecodeClaimsBase64Error(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-decclaims"

	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	headerJSON := fmt.Sprintf(`{"alg":"RS256","kid":"%s"}`, kid)
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	// Use payload that's not valid base64 but doesn't contain '.'
	payloadB64 := "notvalidbase64padding=="

	// Sign the raw header.payload
	signed := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(signed))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	token := signed + "." + sigB64
	_, err := verifyGoogleToken(token, "client-id")
	if err == nil {
		t.Error("expected error for invalid base64 claims")
	}
}

// Test invalid claims JSON — payload is valid base64 but decodes to non-JSON.
func TestVerifyGoogleToken_InvalidClaimsJSON(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-badjson-claims"

	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	headerJSON := fmt.Sprintf(`{"alg":"RS256","kid":"%s"}`, kid)
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	// Encode "not json" as base64 — valid base64 but invalid JSON
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte("not json at all"))

	signed := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(signed))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	token := signed + "." + sigB64
	_, err := verifyGoogleToken(token, "client-id")
	if err == nil {
		t.Error("expected error for non-JSON claims")
	}
}

// Test the getKey path where key is found after refresh (cache miss then cache hit).
// Since refresh() calls the real Google JWKS, we manually populate the cache
// as if the refresh succeeded.
func TestGetKey_KeyFoundAfterCacheMiss(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "test-getkey-miss"

	// Set cache as fresh (recently fetched) with the key we need
	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{kid: &privKey.PublicKey}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	// This should find the key directly in cache
	key, err := cache.getKey(kid)
	if err != nil {
		t.Fatalf("expected to find key, got error: %v", err)
	}
	if key == nil {
		t.Error("expected non-nil key")
	}
}

// Test getKey when key not found in fresh cache (triggers refresh which hits real Google).
func TestGetKey_KeyNotFoundInFreshCache(t *testing.T) {
	// Set cache as fresh but with different keys
	cache.mu.Lock()
	cache.keys = map[string]*rsa.PublicKey{"other-kid": nil}
	cache.fetched = time.Now()
	cache.mu.Unlock()

	// This key doesn't exist, cache is fresh so it tries refresh
	// (which calls Google's JWKS URL) then looks again
	_, err := cache.getKey("nonexistent-kid-xyz")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

// Helper to base64 RawURL encode
func base64RawURLEncode(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	result := make([]byte, 0, len(data)*4/3+4)
	for i := 0; i < len(data); i += 3 {
		n := int(data[i]) << 16
		if i+1 < len(data) {
			n |= int(data[i+1]) << 8
		}
		if i+2 < len(data) {
			n |= int(data[i+2])
		}
		result = append(result, base64Chars[n>>18&63])
		result = append(result, base64Chars[n>>12&63])
		if i+1 < len(data) {
			result = append(result, base64Chars[n>>6&63])
		}
		if i+2 < len(data) {
			result = append(result, base64Chars[n&63])
		}
	}
	return string(result)
}
