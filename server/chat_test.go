package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	experiences := []experienceDoc{
		{
			ID:       "1",
			Company:  "bunny.net",
			Role:     "Product Department Lead",
			Location: "London, UK",
			Bullets:  []string{"Led product strategy", "Shipped AI features"},
		},
	}
	blogs := []blogDoc{
		{
			ID:      "b1",
			Title:   "AI Governance",
			Summary: "Building responsible AI",
			Date:    "2025-01-15",
		},
	}

	prompt := buildSystemPrompt(experiences, blogs)

	// Check bio is included
	if !strings.Contains(prompt, "Product leader with 15+ years") {
		t.Error("expected bio text in system prompt")
	}

	// Check experience data
	if !strings.Contains(prompt, "bunny.net") {
		t.Error("expected company name in prompt")
	}
	if !strings.Contains(prompt, "Product Department Lead") {
		t.Error("expected role in prompt")
	}
	if !strings.Contains(prompt, "London, UK") {
		t.Error("expected location in prompt")
	}
	if !strings.Contains(prompt, "Led product strategy") {
		t.Error("expected bullet point in prompt")
	}

	// Check blog data
	if !strings.Contains(prompt, "AI Governance") {
		t.Error("expected blog title in prompt")
	}
	if !strings.Contains(prompt, "Building responsible AI") {
		t.Error("expected blog summary in prompt")
	}

	// Check tenure refusal instruction is present
	if !strings.Contains(prompt, "Do not answer questions about how long Swati worked") {
		t.Error("expected tenure refusal instruction in prompt")
	}

	// Check contact section
	if !strings.Contains(prompt, "Contact") {
		t.Error("expected contact section in prompt")
	}
}

func TestBuildSystemPrompt_NoBlogs(t *testing.T) {
	prompt := buildSystemPrompt(nil, nil)
	if !strings.Contains(prompt, "## Bio") {
		t.Error("expected bio section even with empty data")
	}
	if strings.Contains(prompt, "## Blog Posts") {
		t.Error("should not include blog section when no blogs")
	}
}

func TestBuildSystemPrompt_ExperienceWithoutLocation(t *testing.T) {
	experiences := []experienceDoc{
		{ID: "1", Company: "ACME", Role: "PM", Location: ""},
	}
	prompt := buildSystemPrompt(experiences, nil)
	if strings.Contains(prompt, "| |") {
		t.Error("should not have empty location pipe")
	}
}

// ---------- Rate limiter tests ----------

func TestRateLimiter_AllowsFirst10(t *testing.T) {
	rl := newRateLimiter()
	for i := 0; i < 10; i++ {
		if !rl.allow("192.168.1.1") {
			t.Errorf("expected request %d to be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksAfter10(t *testing.T) {
	rl := newRateLimiter()
	for i := 0; i < 10; i++ {
		rl.allow("192.168.1.1")
	}
	if rl.allow("192.168.1.1") {
		t.Error("expected 11th request to be blocked")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := newRateLimiter()
	for i := 0; i < 10; i++ {
		rl.allow("192.168.1.1")
	}
	// Different IP should be allowed
	if !rl.allow("192.168.1.2") {
		t.Error("expected different IP to be allowed")
	}
}

// ---------- getClientIP tests ----------

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	if ip := getClientIP(req); ip != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %q", ip)
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.2, 10.0.0.3")
	if ip := getClientIP(req); ip != "10.0.0.2" {
		t.Errorf("expected 10.0.0.2, got %q", ip)
	}
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.4:12345"
	if ip := getClientIP(req); ip != "10.0.0.4" {
		t.Errorf("expected 10.0.0.4, got %q", ip)
	}
}

// ---------- Chat handler validation tests ----------

func TestChatHandler_EmptyBody(t *testing.T) {
	limiter := newRateLimiter()
	handler := chatHandler("fake-key", "https://api.openai.com", "system prompt", limiter)

	req := httptest.NewRequest("POST", "/api/chat", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "messages cannot be empty" {
		t.Errorf("expected 'messages cannot be empty', got %q", body["error"])
	}
}

func TestChatHandler_InvalidRole(t *testing.T) {
	limiter := newRateLimiter()
	handler := chatHandler("fake-key", "https://api.openai.com", "system prompt", limiter)

	payload := chatRequest{
		Messages: []chatMessage{{Role: "admin", Content: "hello"}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestChatHandler_EmptyContent(t *testing.T) {
	limiter := newRateLimiter()
	handler := chatHandler("fake-key", "https://api.openai.com", "system prompt", limiter)

	payload := chatRequest{
		Messages: []chatMessage{{Role: "user", Content: "  "}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestChatHandler_InvalidJSON(t *testing.T) {
	limiter := newRateLimiter()
	handler := chatHandler("fake-key", "https://api.openai.com", "system prompt", limiter)

	req := httptest.NewRequest("POST", "/api/chat", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestChatHandler_RateLimited(t *testing.T) {
	limiter := newRateLimiter()
	handler := chatHandler("fake-key", "https://api.openai.com", "system prompt", limiter)

	// Exhaust rate limit
	for i := 0; i < 11; i++ {
		limiter.allow("192.0.2.1")
	}

	payload := chatRequest{
		Messages: []chatMessage{{Role: "user", Content: "hello"}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.0.2.1:1234"
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
}
