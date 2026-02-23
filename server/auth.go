package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

type jwksCache struct {
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
}

var cache = &jwksCache{}

func (c *jwksCache) getKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if time.Since(c.fetched) < time.Hour && c.keys != nil {
		if key, ok := c.keys[kid]; ok {
			c.mu.RUnlock()
			return key, nil
		}
	}
	c.mu.RUnlock()

	if err := c.refresh(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	key, ok := c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key %s not found in Google JWKS", kid)
	}
	return key, nil
}

func (c *jwksCache) refresh() error {
	resp, err := http.Get(googleCertsURL)
	if err != nil {
		return fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decode JWKS: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		keys[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: e,
		}
	}

	c.mu.Lock()
	c.keys = keys
	c.fetched = time.Now()
	c.mu.Unlock()
	return nil
}

type jwtHeader struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

type jwtClaims struct {
	Iss   string `json:"iss"`
	Aud   string `json:"aud"`
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
}

// verifyGoogleToken verifies a Google ID token's signature and claims.
func verifyGoogleToken(tokenStr, clientID string) (*jwtClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode header.
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	// Get the public key from Google's JWKS.
	pubKey, err := cache.getKey(header.Kid)
	if err != nil {
		return nil, err
	}

	// Verify RS256 signature.
	signed := []byte(parts[0] + "." + parts[1])
	hash := sha256.Sum256(signed)
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sig); err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	// Decode claims.
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode claims: %w", err)
	}
	var claims jwtClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}

	// Verify issuer.
	if claims.Iss != "accounts.google.com" && claims.Iss != "https://accounts.google.com" {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Iss)
	}

	// Verify audience matches our Google Client ID.
	if claims.Aud != clientID {
		return nil, fmt.Errorf("invalid audience")
	}

	// Verify expiration.
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// adminOnly wraps a handler to require a valid Google ID token from the admin email.
func adminOnly(clientID, adminEmail string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")

		claims, err := verifyGoogleToken(token, clientID)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token: "+err.Error())
			return
		}

		if claims.Email != adminEmail {
			writeError(w, http.StatusForbidden, "not authorized")
			return
		}

		next(w, r)
	}
}
