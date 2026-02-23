package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
