package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------- Prompt builder tests ----------

func TestBuildSuggestPrompt_Basic(t *testing.T) {
	req := suggestRequest{
		Company:   "ACME Corp",
		Role:      "Product Manager",
		Location:  "London",
		StartDate: "Jan 2020",
		EndDate:   "Present",
		Bullets:   []string{"Led product strategy", "Grew revenue by 20%"},
	}

	messages := buildSuggestPrompt(req)
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("expected system role, got %q", messages[0].Role)
	}
	if messages[1].Role != "user" {
		t.Errorf("expected user role, got %q", messages[1].Role)
	}

	// System prompt should contain anti-fabrication rules
	if !strings.Contains(messages[0].Content, "DO NOT invent") {
		t.Error("system prompt missing anti-fabrication rules")
	}

	// User message should contain experience data
	userContent := messages[1].Content
	if !strings.Contains(userContent, "ACME Corp") {
		t.Error("user message missing company")
	}
	if !strings.Contains(userContent, "Product Manager") {
		t.Error("user message missing role")
	}
	if !strings.Contains(userContent, "London") {
		t.Error("user message missing location")
	}
	if !strings.Contains(userContent, "Led product strategy") {
		t.Error("user message missing bullet")
	}
}

func TestBuildSuggestPrompt_WithRoleDescription(t *testing.T) {
	req := suggestRequest{
		Company:         "ACME Corp",
		Role:            "Product Manager",
		Bullets:         []string{"Led team"},
		RoleDescription: "Senior PM role at a fintech company",
	}

	messages := buildSuggestPrompt(req)
	if !strings.Contains(messages[0].Content, "Senior PM role at a fintech company") {
		t.Error("system prompt should include role description")
	}
	if !strings.Contains(messages[0].Content, "tailoring this experience") {
		t.Error("system prompt should include tailoring instruction")
	}
}

func TestBuildSuggestPrompt_NoRoleDescription(t *testing.T) {
	req := suggestRequest{
		Company: "ACME Corp",
		Role:    "PM",
		Bullets: []string{"Led team"},
	}

	messages := buildSuggestPrompt(req)
	if strings.Contains(messages[0].Content, "tailoring this experience") {
		t.Error("system prompt should NOT include tailoring when no role description")
	}
}

// ---------- stripCodeFences tests ----------

func TestStripCodeFences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain JSON", `["a","b"]`, `["a","b"]`},
		{"json fences", "```json\n[\"a\",\"b\"]\n```", `["a","b"]`},
		{"plain fences", "```\n[\"a\",\"b\"]\n```", `["a","b"]`},
		{"with whitespace", "  ```json\n[\"a\"]\n```  ", `["a"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripCodeFences(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// ---------- Validation tests ----------

func TestSuggestHandler_EmptyBody(t *testing.T) {
	handler := suggestHandler("test-key", "http://localhost")

	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte("")))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSuggestHandler_MissingCompany(t *testing.T) {
	handler := suggestHandler("test-key", "http://localhost")

	body := `{"role":"PM","bullets":["Did stuff"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSuggestHandler_MissingRole(t *testing.T) {
	handler := suggestHandler("test-key", "http://localhost")

	body := `{"company":"ACME","bullets":["Did stuff"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSuggestHandler_EmptyBullets(t *testing.T) {
	handler := suggestHandler("test-key", "http://localhost")

	body := `{"company":"ACME","role":"PM","bullets":[]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSuggestHandler_NoOpenAIKey(t *testing.T) {
	handler := suggestHandler("", "http://localhost")

	body := `{"company":"ACME","role":"PM","bullets":["Did stuff"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

// ---------- Integration tests with mock OpenAI ----------

func TestSuggestHandler_Success(t *testing.T) {
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("expected 'Bearer test-key', got %q", auth)
		}

		response := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{
				{Message: chatMessage{Role: "assistant", Content: `["Spearheaded product strategy","Drove 20% revenue growth"]`}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockOpenAI.Close()

	handler := suggestHandler("test-key", mockOpenAI.URL)

	body := `{"company":"ACME","role":"PM","bullets":["Led product strategy","Grew revenue by 20%"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result suggestResponse
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result.Bullets) != 2 {
		t.Fatalf("expected 2 bullets, got %d", len(result.Bullets))
	}
	if result.Bullets[0] != "Spearheaded product strategy" {
		t.Errorf("unexpected bullet: %q", result.Bullets[0])
	}
}

func TestSuggestHandler_SuccessWithCodeFences(t *testing.T) {
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{
				{Message: chatMessage{Role: "assistant", Content: "```json\n[\"Improved bullet\"]\n```"}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockOpenAI.Close()

	handler := suggestHandler("test-key", mockOpenAI.URL)

	body := `{"company":"ACME","role":"PM","bullets":["Original bullet"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result suggestResponse
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result.Bullets) != 1 || result.Bullets[0] != "Improved bullet" {
		t.Errorf("unexpected result: %v", result.Bullets)
	}
}

func TestSuggestHandler_WithRoleDescription(t *testing.T) {
	var capturedPrompt string
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var oaiReq openaiRequest
		json.NewDecoder(r.Body).Decode(&oaiReq)
		capturedPrompt = oaiReq.Messages[0].Content

		response := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{
				{Message: chatMessage{Role: "assistant", Content: `["Tailored bullet"]`}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockOpenAI.Close()

	handler := suggestHandler("test-key", mockOpenAI.URL)

	body := `{"company":"ACME","role":"PM","bullets":["Led team"],"role_description":"Looking for a senior PM with data skills"}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(capturedPrompt, "Looking for a senior PM with data skills") {
		t.Error("prompt should contain role description")
	}
}

func TestSuggestHandler_OpenAIError(t *testing.T) {
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer mockOpenAI.Close()

	handler := suggestHandler("test-key", mockOpenAI.URL)

	body := `{"company":"ACME","role":"PM","bullets":["Led team"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestSuggestHandler_InvalidAIResponse(t *testing.T) {
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{
				{Message: chatMessage{Role: "assistant", Content: "This is not JSON at all"}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockOpenAI.Close()

	handler := suggestHandler("test-key", mockOpenAI.URL)

	body := `{"company":"ACME","role":"PM","bullets":["Led team"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestSuggestHandler_EmptyChoices(t *testing.T) {
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := openaiResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockOpenAI.Close()

	handler := suggestHandler("test-key", mockOpenAI.URL)

	body := `{"company":"ACME","role":"PM","bullets":["Led team"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestSuggestHandler_InvalidJSONFromOpenAI(t *testing.T) {
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{not valid json at all"))
	}))
	defer mockOpenAI.Close()

	handler := suggestHandler("test-key", mockOpenAI.URL)

	body := `{"company":"ACME","role":"PM","bullets":["Led team"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestSuggestHandler_UnreachableServer(t *testing.T) {
	handler := suggestHandler("test-key", "http://127.0.0.1:1")

	body := `{"company":"ACME","role":"PM","bullets":["Led team"]}`
	req := httptest.NewRequest("POST", "/api/experiences/suggest", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestBuildSuggestPrompt_NoLocationNoDates(t *testing.T) {
	req := suggestRequest{
		Company: "ACME",
		Role:    "PM",
		Bullets: []string{"Led team"},
	}

	messages := buildSuggestPrompt(req)
	userContent := messages[1].Content
	if strings.Contains(userContent, "Location") {
		t.Error("should not include Location when empty")
	}
	if strings.Contains(userContent, "Dates") {
		t.Error("should not include Dates when both empty")
	}
}
