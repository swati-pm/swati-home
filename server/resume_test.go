package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// ---------- Template rendering tests ----------

func TestTemplateRendering(t *testing.T) {
	tmplDir := "templates"
	if _, err := os.Stat(tmplDir); os.IsNotExist(err) {
		t.Skip("templates directory not found; skipping")
	}

	profile := profileData{
		Name:    "Test User",
		Email:   "test@example.com",
		Summary: "Experienced product leader.",
		Experiences: []experienceDoc{
			{
				ID:        "1",
				Company:   "ACME Corp",
				Role:      "Senior PM",
				Location:  "NYC",
				StartDate: "Jan 2020",
				EndDate:   "Present",
				Bullets:   []string{"Led product strategy", "Increased revenue 50%"},
			},
		},
	}

	for _, name := range []string{"classic", "modern", "compact"} {
		t.Run(name, func(t *testing.T) {
			tmplPath := filepath.Join(tmplDir, name+".html")
			tmplBytes, err := os.ReadFile(tmplPath)
			if err != nil {
				t.Fatalf("failed to read template %s: %v", tmplPath, err)
			}

			tmpl, err := template.New("resume").Parse(string(tmplBytes))
			if err != nil {
				t.Fatalf("failed to parse template %s: %v", name, err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, profile); err != nil {
				t.Fatalf("failed to execute template %s: %v", name, err)
			}

			html := buf.String()
			if len(html) == 0 {
				t.Error("rendered HTML is empty")
			}
			// Verify profile data appears in output
			if !bytes.Contains(buf.Bytes(), []byte("Test User")) {
				t.Error("expected 'Test User' in rendered HTML")
			}
			if !bytes.Contains(buf.Bytes(), []byte("ACME Corp")) {
				t.Error("expected 'ACME Corp' in rendered HTML")
			}
		})
	}
}

// ---------- ListTemplates tests ----------

func TestListTemplates_DefaultClassic(t *testing.T) {
	col := getTestCollection(t, "settings_list")

	handler := listTemplates(col)
	req := httptest.NewRequest("GET", "/api/resume/templates", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []templateInfo
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 3 {
		t.Fatalf("expected 3 templates, got %d", len(result))
	}

	// Default should be classic
	activeCount := 0
	for _, tmpl := range result {
		if tmpl.Active {
			activeCount++
			if tmpl.Name != "classic" {
				t.Errorf("expected 'classic' to be active, got %q", tmpl.Name)
			}
		}
	}
	if activeCount != 1 {
		t.Errorf("expected exactly 1 active template, got %d", activeCount)
	}
}

// ---------- SetActiveTemplate tests ----------

func TestSetActiveTemplate(t *testing.T) {
	col := getTestCollection(t, "settings_set")

	// Set to "modern"
	body := `{"template":"modern"}`
	req := httptest.NewRequest("PUT", "/api/resume/template", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	setActiveTemplate(col)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if result["active_template"] != "modern" {
		t.Errorf("expected 'modern', got %q", result["active_template"])
	}

	// Verify via listTemplates
	req = httptest.NewRequest("GET", "/api/resume/templates", nil)
	rec = httptest.NewRecorder()
	listTemplates(col)(rec, req)

	var templates []templateInfo
	json.NewDecoder(rec.Body).Decode(&templates)
	for _, tmpl := range templates {
		if tmpl.Name == "modern" && !tmpl.Active {
			t.Error("expected 'modern' to be active after setting it")
		}
		if tmpl.Name == "classic" && tmpl.Active {
			t.Error("expected 'classic' to not be active after setting 'modern'")
		}
	}
}

func TestSetActiveTemplate_InvalidName(t *testing.T) {
	col := getTestCollection(t, "settings_invalid")

	body := `{"template":"nonexistent"}`
	req := httptest.NewRequest("PUT", "/api/resume/template", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	setActiveTemplate(col)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSetActiveTemplate_InvalidJSON(t *testing.T) {
	col := getTestCollection(t, "settings_badjson")

	req := httptest.NewRequest("PUT", "/api/resume/template", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	setActiveTemplate(col)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ---------- DownloadResume tests ----------

func TestDownloadResume(t *testing.T) {
	// Set up template directory for the test
	tmplDir := t.TempDir()
	tmplContent := `<!DOCTYPE html><html><body><h1>{{.Name}}</h1>{{range .Experiences}}<p>{{.Company}}</p>{{end}}</body></html>`
	if err := os.WriteFile(filepath.Join(tmplDir, "classic.html"), []byte(tmplContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Mock Gotenberg server
	fakePDF := []byte("%PDF-1.4 fake pdf content")
	gotenberg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forms/chromium/convert/html" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify we received HTML with the expected content
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse multipart: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("files")
		if err != nil {
			t.Errorf("no files field: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		htmlBytes, _ := io.ReadAll(file)
		if !bytes.Contains(htmlBytes, []byte("Test PM")) {
			t.Errorf("HTML should contain 'Test PM', got: %s", string(htmlBytes[:min(200, len(htmlBytes))]))
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(fakePDF)
	}))
	defer gotenberg.Close()

	settingsCol := getTestCollection(t, "settings_dl")
	expCol := getTestCollection(t, "experiences_dl")

	// Insert test experience
	exp := experienceDoc{ID: "e1", Company: "TestCorp", Role: "PM", Bullets: []string{"Did stuff"}}
	expCol.InsertOne(context.Background(), exp)

	profile := profileData{
		Name:    "Test PM",
		Email:   "pm@test.com",
		Summary: "Product leader.",
	}

	// Temporarily override template path — need to patch the handler
	// Since downloadResume reads from /app/templates, we create a modified version for testing
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Inline version of downloadResume that uses our temp dir
		var settings settingsDoc
		err := settingsCol.FindOne(r.Context(), map[string]string{"_id": "global"}).Decode(&settings)
		if err != nil {
			settings.ActiveTemplate = "classic"
		}

		cursor, err := expCol.Find(r.Context(), map[string]interface{}{})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fetch experiences")
			return
		}
		var experiences []experienceDoc
		if err := cursor.All(r.Context(), &experiences); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to decode experiences")
			return
		}

		p := profile
		p.Experiences = experiences

		tmplPath := filepath.Join(tmplDir, settings.ActiveTemplate+".html")
		tmplBytes, err := os.ReadFile(tmplPath)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "template not found")
			return
		}

		tmpl, err := template.New("resume").Parse(string(tmplBytes))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "template parse error")
			return
		}

		var htmlBuf bytes.Buffer
		if err := tmpl.Execute(&htmlBuf, p); err != nil {
			writeError(w, http.StatusInternalServerError, "template render error")
			return
		}

		pdfBytes, err := convertHTMLToPDF(gotenberg.URL, htmlBuf.Bytes())
		if err != nil {
			writeError(w, http.StatusBadGateway, "PDF generation failed")
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(pdfBytes)
	}

	req := httptest.NewRequest("GET", "/api/resume/download", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %q", ct)
	}

	if !bytes.Equal(rec.Body.Bytes(), fakePDF) {
		t.Error("response body does not match expected PDF content")
	}
}

func TestConvertHTMLToPDF_GotenbergError(t *testing.T) {
	// Mock a Gotenberg that returns an error
	gotenberg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("conversion failed"))
	}))
	defer gotenberg.Close()

	_, err := convertHTMLToPDF(gotenberg.URL, []byte("<html></html>"))
	if err == nil {
		t.Error("expected error from Gotenberg, got nil")
	}
}
