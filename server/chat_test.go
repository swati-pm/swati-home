package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

// ---------- getClientIP edge case ----------

func TestGetClientIP_RemoteAddrNoPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.5"
	if ip := getClientIP(req); ip != "10.0.0.5" {
		t.Errorf("expected 10.0.0.5, got %q", ip)
	}
}

// ---------- Chat handler tests with mock OpenAI ----------

// newMockOpenAI creates a test HTTP server that mimics the OpenAI
// /chat/completions endpoint. The handler function receives the decoded
// openaiRequest so callers can inspect what was sent.
func newMockOpenAI(handler func(w http.ResponseWriter, oaiReq openaiRequest)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var oaiReq openaiRequest
		if err := json.Unmarshal(body, &oaiReq); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		handler(w, oaiReq)
	}))
}

func TestChatHandler_Success(t *testing.T) {
	mock := newMockOpenAI(func(w http.ResponseWriter, oaiReq openaiRequest) {
		resp := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{
				{Message: chatMessage{Role: "assistant", Content: "Hello from AI"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer mock.Close()

	limiter := newRateLimiter()
	handler := chatHandler("test-key", mock.URL, "test system prompt", limiter)

	payload := chatRequest{
		Messages: []chatMessage{{Role: "user", Content: "Hi"}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["reply"] != "Hello from AI" {
		t.Errorf("expected reply 'Hello from AI', got %q", result["reply"])
	}
}

func TestChatHandler_OpenAINon200(t *testing.T) {
	mock := newMockOpenAI(func(w http.ResponseWriter, oaiReq openaiRequest) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":{"message":"server error"}}`)
	})
	defer mock.Close()

	limiter := newRateLimiter()
	handler := chatHandler("test-key", mock.URL, "test system prompt", limiter)

	payload := chatRequest{
		Messages: []chatMessage{{Role: "user", Content: "Hi"}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}

	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if !strings.Contains(result["error"], "AI service returned an error") {
		t.Errorf("unexpected error message: %q", result["error"])
	}
}

func TestChatHandler_OpenAIEmptyChoices(t *testing.T) {
	mock := newMockOpenAI(func(w http.ResponseWriter, oaiReq openaiRequest) {
		resp := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer mock.Close()

	limiter := newRateLimiter()
	handler := chatHandler("test-key", mock.URL, "test system prompt", limiter)

	payload := chatRequest{
		Messages: []chatMessage{{Role: "user", Content: "Hi"}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}

	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if !strings.Contains(result["error"], "AI returned no response") {
		t.Errorf("unexpected error message: %q", result["error"])
	}
}

func TestChatHandler_OpenAIInvalidJSON(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{not valid json`)
	}))
	defer mock.Close()

	limiter := newRateLimiter()
	handler := chatHandler("test-key", mock.URL, "test system prompt", limiter)

	payload := chatRequest{
		Messages: []chatMessage{{Role: "user", Content: "Hi"}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}

	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if !strings.Contains(result["error"], "failed to parse AI response") {
		t.Errorf("unexpected error message: %q", result["error"])
	}
}

func TestChatHandler_TruncatesTo10Messages(t *testing.T) {
	var capturedMessages []chatMessage

	mock := newMockOpenAI(func(w http.ResponseWriter, oaiReq openaiRequest) {
		capturedMessages = oaiReq.Messages
		resp := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{
				{Message: chatMessage{Role: "assistant", Content: "ok"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer mock.Close()

	limiter := newRateLimiter()
	handler := chatHandler("test-key", mock.URL, "test system prompt", limiter)

	// Build 12 messages alternating user/assistant
	messages := make([]chatMessage, 12)
	for i := 0; i < 12; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		messages[i] = chatMessage{Role: role, Content: fmt.Sprintf("msg-%d", i)}
	}

	payload := chatRequest{Messages: messages}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// OpenAI should receive 1 system message + 10 user/assistant messages = 11 total
	expectedTotal := 11
	if len(capturedMessages) != expectedTotal {
		t.Fatalf("expected %d messages sent to OpenAI, got %d", expectedTotal, len(capturedMessages))
	}

	// First message must be the system prompt
	if capturedMessages[0].Role != "system" {
		t.Errorf("expected first message role 'system', got %q", capturedMessages[0].Role)
	}
	if capturedMessages[0].Content != "test system prompt" {
		t.Errorf("expected system prompt content, got %q", capturedMessages[0].Content)
	}

	// The truncation should keep the LAST 10 messages (indices 2..11 from original),
	// so the first user message forwarded should be "msg-2"
	if capturedMessages[1].Content != "msg-2" {
		t.Errorf("expected first forwarded message to be 'msg-2', got %q", capturedMessages[1].Content)
	}

	// And the last should be "msg-11"
	if capturedMessages[10].Content != "msg-11" {
		t.Errorf("expected last forwarded message to be 'msg-11', got %q", capturedMessages[10].Content)
	}
}

func TestChatHandler_ExactlyTenMessagesNotTruncated(t *testing.T) {
	var capturedMessages []chatMessage

	mock := newMockOpenAI(func(w http.ResponseWriter, oaiReq openaiRequest) {
		capturedMessages = oaiReq.Messages
		resp := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{
				{Message: chatMessage{Role: "assistant", Content: "ok"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer mock.Close()

	limiter := newRateLimiter()
	handler := chatHandler("test-key", mock.URL, "test system prompt", limiter)

	// Build exactly 10 messages
	messages := make([]chatMessage, 10)
	for i := 0; i < 10; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		messages[i] = chatMessage{Role: role, Content: fmt.Sprintf("msg-%d", i)}
	}

	payload := chatRequest{Messages: messages}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// 1 system + 10 user/assistant = 11
	expectedTotal := 11
	if len(capturedMessages) != expectedTotal {
		t.Fatalf("expected %d messages sent to OpenAI, got %d", expectedTotal, len(capturedMessages))
	}

	// First user message should be "msg-0" (no truncation occurred)
	if capturedMessages[1].Content != "msg-0" {
		t.Errorf("expected first forwarded message to be 'msg-0', got %q", capturedMessages[1].Content)
	}
}

func TestChatHandler_SystemPromptIncluded(t *testing.T) {
	var capturedMessages []chatMessage

	mock := newMockOpenAI(func(w http.ResponseWriter, oaiReq openaiRequest) {
		capturedMessages = oaiReq.Messages
		resp := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{
				{Message: chatMessage{Role: "assistant", Content: "ok"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer mock.Close()

	limiter := newRateLimiter()
	systemPrompt := "You are a helpful assistant for testing."
	handler := chatHandler("test-key", mock.URL, systemPrompt, limiter)

	payload := chatRequest{
		Messages: []chatMessage{{Role: "user", Content: "hello"}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if len(capturedMessages) < 1 {
		t.Fatal("expected at least 1 message sent to OpenAI")
	}

	if capturedMessages[0].Role != "system" {
		t.Errorf("expected first message role 'system', got %q", capturedMessages[0].Role)
	}
	if capturedMessages[0].Content != systemPrompt {
		t.Errorf("expected system prompt %q, got %q", systemPrompt, capturedMessages[0].Content)
	}
}

func TestChatHandler_AuthorizationHeader(t *testing.T) {
	var capturedAuthHeader string

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuthHeader = r.Header.Get("Authorization")

		resp := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{
				{Message: chatMessage{Role: "assistant", Content: "ok"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mock.Close()

	limiter := newRateLimiter()
	handler := chatHandler("sk-test-secret-key", mock.URL, "prompt", limiter)

	payload := chatRequest{
		Messages: []chatMessage{{Role: "user", Content: "hello"}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	expected := "Bearer sk-test-secret-key"
	if capturedAuthHeader != expected {
		t.Errorf("expected Authorization header %q, got %q", expected, capturedAuthHeader)
	}
}

// ---------- Rate limiter cleanup path ----------

func TestRateLimiter_CleanupPath(t *testing.T) {
	rl := newRateLimiter()

	// Fill up an entry that will expire
	rl.mu.Lock()
	rl.visitors["expired-ip"] = &visitorInfo{count: 5, resetAt: time.Now().Add(-time.Minute)}
	rl.calls = 99 // Next call will trigger cleanup (100th)
	rl.mu.Unlock()

	// This call should trigger cleanup and remove "expired-ip"
	result := rl.allow("new-ip")
	if !result {
		t.Error("expected new-ip to be allowed")
	}

	rl.mu.Lock()
	_, exists := rl.visitors["expired-ip"]
	rl.mu.Unlock()
	if exists {
		t.Error("expected expired-ip to be cleaned up")
	}
}

func TestRateLimiter_CleanupKeepsActive(t *testing.T) {
	rl := newRateLimiter()

	// Fill up an entry that has NOT expired
	rl.mu.Lock()
	rl.visitors["active-ip"] = &visitorInfo{count: 5, resetAt: time.Now().Add(time.Minute)}
	rl.calls = 99
	rl.mu.Unlock()

	rl.allow("trigger-ip")

	rl.mu.Lock()
	_, exists := rl.visitors["active-ip"]
	rl.mu.Unlock()
	if !exists {
		t.Error("expected active-ip to NOT be cleaned up")
	}
}

func TestChatHandler_UnreachableServer(t *testing.T) {
	limiter := newRateLimiter()
	// Use a port that's definitely not listening
	handler := chatHandler("test-key", "http://127.0.0.1:1", "system prompt", limiter)

	payload := chatRequest{
		Messages: []chatMessage{{Role: "user", Content: "Hi"}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502 for unreachable server, got %d", rec.Code)
	}
}

func TestRateLimiter_ResetAfterExpiry(t *testing.T) {
	rl := newRateLimiter()

	// Simulate expired visitor
	rl.mu.Lock()
	rl.visitors["192.168.1.1"] = &visitorInfo{count: 15, resetAt: time.Now().Add(-time.Second)}
	rl.mu.Unlock()

	// Should be allowed again since the window has expired
	if !rl.allow("192.168.1.1") {
		t.Error("expected request to be allowed after window expiry")
	}
}
