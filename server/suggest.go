package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type suggestRequest struct {
	Company         string   `json:"company"`
	Role            string   `json:"role"`
	Location        string   `json:"location"`
	StartDate       string   `json:"start_date"`
	EndDate         string   `json:"end_date"`
	Bullets         []string `json:"bullets"`
	RoleDescription string   `json:"role_description,omitempty"`
}

type suggestResponse struct {
	Bullets []string `json:"bullets"`
}

func buildSuggestPrompt(req suggestRequest) []chatMessage {
	var sb strings.Builder

	sb.WriteString(`You are a professional resume bullet point editor. Your ONLY job is to improve the wording, impact, and clarity of the provided experience bullet points.

ABSOLUTE RULES — violating any of these makes your output invalid:
1. ONLY use information explicitly present in the provided experience data (company name, role title, location, dates, and existing bullet content).
2. DO NOT invent, fabricate, or assume ANY:
   - Metrics, numbers, or percentages not already in the bullets
   - Technologies, tools, or platforms not already mentioned
   - Team sizes, revenue figures, or KPIs not already stated
   - Projects, products, or initiatives not already referenced
   - Dates, durations, or timelines not already provided
3. You may restructure sentences, improve verb choices (use strong action verbs), improve clarity, remove redundancy, and make phrasing more concise and impactful.
4. Preserve the total number of bullets — return exactly the same count as provided.
5. Return ONLY a JSON array of strings, one per bullet. No other text, no markdown, no explanation.`)

	if req.RoleDescription != "" {
		sb.WriteString("\n\nThe user is tailoring this experience for a specific role. Here is the target role description:\n---\n")
		sb.WriteString(req.RoleDescription)
		sb.WriteString("\n---\nAdjust the WORDING and EMPHASIS of the bullets to highlight aspects most relevant to this role. You MUST NOT add any skills, technologies, metrics, or details not already present in the original bullets.")
	}

	systemMsg := sb.String()

	var userContent strings.Builder
	userContent.WriteString(fmt.Sprintf("Company: %s\n", req.Company))
	userContent.WriteString(fmt.Sprintf("Role: %s\n", req.Role))
	if req.Location != "" {
		userContent.WriteString(fmt.Sprintf("Location: %s\n", req.Location))
	}
	if req.StartDate != "" || req.EndDate != "" {
		userContent.WriteString(fmt.Sprintf("Dates: %s – %s\n", req.StartDate, req.EndDate))
	}
	userContent.WriteString("\nCurrent bullets:\n")
	for i, b := range req.Bullets {
		userContent.WriteString(fmt.Sprintf("%d. %s\n", i+1, b))
	}

	return []chatMessage{
		{Role: "system", Content: systemMsg},
		{Role: "user", Content: userContent.String()},
	}
}

// stripCodeFences removes markdown code fences that models sometimes wrap JSON in.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func suggestHandler(openaiKey, openaiBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if openaiKey == "" {
			writeError(w, http.StatusServiceUnavailable, "AI service not configured")
			return
		}

		var req suggestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(req.Company) == "" || strings.TrimSpace(req.Role) == "" {
			writeError(w, http.StatusBadRequest, "company and role are required")
			return
		}
		if len(req.Bullets) == 0 {
			writeError(w, http.StatusBadRequest, "bullets cannot be empty")
			return
		}

		messages := buildSuggestPrompt(req)

		oaiReq := openaiRequest{
			Model:    "gpt-4o-mini",
			Messages: messages,
		}

		body, err := json.Marshal(oaiReq)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build request")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiBaseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create request")
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+openaiKey)

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			log.Printf("suggest: OpenAI request error: %v", err)
			writeError(w, http.StatusBadGateway, "failed to reach AI service")
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("suggest: OpenAI read error: %v", err)
			writeError(w, http.StatusBadGateway, "failed to read AI response")
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("suggest: OpenAI returned %d: %s", resp.StatusCode, string(respBody))
			writeError(w, http.StatusBadGateway, "AI service returned an error")
			return
		}

		var oaiResp openaiResponse
		if err := json.Unmarshal(respBody, &oaiResp); err != nil {
			log.Printf("suggest: OpenAI decode error: %v", err)
			writeError(w, http.StatusBadGateway, "failed to parse AI response")
			return
		}

		if len(oaiResp.Choices) == 0 {
			writeError(w, http.StatusBadGateway, "AI returned no response")
			return
		}

		content := stripCodeFences(oaiResp.Choices[0].Message.Content)

		var bullets []string
		if err := json.Unmarshal([]byte(content), &bullets); err != nil {
			log.Printf("suggest: failed to parse bullets JSON: %v, content: %s", err, content)
			writeError(w, http.StatusBadGateway, "failed to parse AI suggestions")
			return
		}

		writeJSON(w, http.StatusOK, suggestResponse{Bullets: bullets})
	}
}
