package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ---------- System prompt builder ----------

const bio = `Product leader with 15+ years building data and AI-powered platforms that turn behavioural signals into trusted, high-performing user experiences. I lead multi-PM organisations and partner deeply with Data Science and Engineering to move from research to production — shipping intelligent products that improve matching, discovery, decisioning, and outcomes at scale.`

func buildSystemPrompt(experiences []experienceDoc, blogs []blogDoc) string {
	var sb strings.Builder

	sb.WriteString("You are Swati Aggarwal's portfolio assistant. You ONLY answer questions about Swati based on the information below. ")
	sb.WriteString("If asked about anything not covered here, politely say you can only discuss Swati's professional background and work. ")
	sb.WriteString("Do not make up information. Do not invent URLs, dates, or achievements not listed below. ")
	sb.WriteString("Do not answer questions about how long Swati worked at any company, her tenure, employment duration, or time spent at any role. If asked, politely decline and say this information is not available. ")
	sb.WriteString("Keep answers concise and professional.\n\n")

	sb.WriteString("## Bio\n")
	sb.WriteString(bio)
	sb.WriteString("\n\n")

	sb.WriteString("## Professional Experience\n\n")
	for _, exp := range experiences {
		header := fmt.Sprintf("### %s | %s", exp.Company, exp.Role)
		if exp.Location != "" {
			header += " | " + exp.Location
		}
		sb.WriteString(header + "\n")
		for _, bullet := range exp.Bullets {
			sb.WriteString("- " + bullet + "\n")
		}
		sb.WriteString("\n")
	}

	if len(blogs) > 0 {
		sb.WriteString("## Blog Posts\n\n")
		for _, blog := range blogs {
			sb.WriteString(fmt.Sprintf("### %s (%s)\n", blog.Title, blog.Date))
			sb.WriteString(blog.Summary + "\n\n")
		}
	}

	sb.WriteString("## Contact\nVisitors can reach Swati through the Contact page on this website.\n")

	return sb.String()
}

// ---------- Rate limiter ----------

type visitorInfo struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitorInfo
	calls    int // counter for periodic cleanup
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{visitors: make(map[string]*visitorInfo)}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Periodic cleanup every 100 calls
	rl.calls++
	if rl.calls%100 == 0 {
		now := time.Now()
		for k, v := range rl.visitors {
			if now.After(v.resetAt) {
				delete(rl.visitors, k)
			}
		}
	}

	v, ok := rl.visitors[ip]
	now := time.Now()

	if !ok || now.After(v.resetAt) {
		rl.visitors[ip] = &visitorInfo{count: 1, resetAt: now.Add(time.Minute)}
		return true
	}

	v.count++
	return v.count <= 10
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.SplitN(ip, ",", 2)[0]
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ---------- Chat handler ----------

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Messages []chatMessage `json:"messages"`
}

type openaiRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type openaiResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func chatHandler(openaiKey, systemPrompt string, limiter *rateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Rate limit
		if !limiter.allow(getClientIP(r)) {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded, please try again in a minute")
			return
		}

		// Decode request
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if len(req.Messages) == 0 {
			writeError(w, http.StatusBadRequest, "messages cannot be empty")
			return
		}

		// Validate and truncate to last 10 messages
		if len(req.Messages) > 10 {
			req.Messages = req.Messages[len(req.Messages)-10:]
		}

		for _, msg := range req.Messages {
			if msg.Role != "user" && msg.Role != "assistant" {
				writeError(w, http.StatusBadRequest, "message role must be 'user' or 'assistant'")
				return
			}
			if strings.TrimSpace(msg.Content) == "" {
				writeError(w, http.StatusBadRequest, "message content cannot be empty")
				return
			}
		}

		// Build OpenAI request
		messages := make([]chatMessage, 0, len(req.Messages)+1)
		messages = append(messages, chatMessage{Role: "system", Content: systemPrompt})
		messages = append(messages, req.Messages...)

		oaiReq := openaiRequest{
			Model:    "gpt-4o-mini",
			Messages: messages,
		}

		body, err := json.Marshal(oaiReq)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build request")
			return
		}

		// Call OpenAI with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create request")
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+openaiKey)

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			log.Printf("OpenAI request error: %v", err)
			writeError(w, http.StatusBadGateway, "failed to reach AI service")
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("OpenAI read error: %v", err)
			writeError(w, http.StatusBadGateway, "failed to read AI response")
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("OpenAI returned %d: %s", resp.StatusCode, string(respBody))
			writeError(w, http.StatusBadGateway, "AI service returned an error")
			return
		}

		var oaiResp openaiResponse
		if err := json.Unmarshal(respBody, &oaiResp); err != nil {
			log.Printf("OpenAI decode error: %v", err)
			writeError(w, http.StatusBadGateway, "failed to parse AI response")
			return
		}

		if len(oaiResp.Choices) == 0 {
			writeError(w, http.StatusBadGateway, "AI returned no response")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"reply": oaiResp.Choices[0].Message.Content,
		})
	}
}
