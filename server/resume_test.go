package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
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

// mockGotenberg creates a test server that mimics Gotenberg PDF conversion.
func mockGotenberg(t *testing.T, fakePDF []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forms/chromium/convert/html" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("files")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(fakePDF)
	}))
}

func TestDownloadResume(t *testing.T) {
	tmplDir := t.TempDir()
	tmplContent := `<!DOCTYPE html><html><body><h1>{{.Name}}</h1>{{range .Experiences}}<p>{{.Company}}</p>{{end}}</body></html>`
	os.WriteFile(filepath.Join(tmplDir, "classic.html"), []byte(tmplContent), 0644)

	fakePDF := []byte("%PDF-1.4 fake pdf content")
	gotenberg := mockGotenberg(t, fakePDF)
	defer gotenberg.Close()

	settingsCol := getTestCollection(t, "settings_dl")
	expCol := getTestCollection(t, "experiences_dl")
	expCol.InsertOne(context.Background(), experienceDoc{ID: "e1", Company: "TestCorp", Role: "PM", Bullets: []string{"Did stuff"}})

	profile := profileData{Name: "Test PM", Email: "pm@test.com", Summary: "Product leader."}

	handler := downloadResume(settingsCol, expCol, gotenberg.URL, tmplDir, profile)

	req := httptest.NewRequest("GET", "/api/resume/download", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %q", ct)
	}
	if !bytes.Equal(rec.Body.Bytes(), fakePDF) {
		t.Error("response body does not match expected PDF content")
	}
	// Verify Content-Disposition
	cd := rec.Header().Get("Content-Disposition")
	if cd == "" {
		t.Error("expected Content-Disposition header")
	}
	// Verify Content-Length
	cl := rec.Header().Get("Content-Length")
	if cl != fmt.Sprintf("%d", len(fakePDF)) {
		t.Errorf("expected Content-Length %d, got %s", len(fakePDF), cl)
	}
}

func TestDownloadResume_MissingTemplate(t *testing.T) {
	tmplDir := t.TempDir() // empty dir — no templates

	settingsCol := getTestCollection(t, "settings_dl_notempl")
	expCol := getTestCollection(t, "experiences_dl_notempl")

	profile := profileData{Name: "Test PM", Email: "pm@test.com"}
	handler := downloadResume(settingsCol, expCol, "http://unused", tmplDir, profile)

	req := httptest.NewRequest("GET", "/api/resume/download", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestDownloadResume_GotenbergFailure(t *testing.T) {
	tmplDir := t.TempDir()
	os.WriteFile(filepath.Join(tmplDir, "classic.html"), []byte(`<html>{{.Name}}</html>`), 0644)

	gotenberg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("conversion failed"))
	}))
	defer gotenberg.Close()

	settingsCol := getTestCollection(t, "settings_dl_goterr")
	expCol := getTestCollection(t, "experiences_dl_goterr")

	profile := profileData{Name: "Test PM", Email: "pm@test.com"}
	handler := downloadResume(settingsCol, expCol, gotenberg.URL, tmplDir, profile)

	req := httptest.NewRequest("GET", "/api/resume/download", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestDownloadResume_BadTemplate(t *testing.T) {
	tmplDir := t.TempDir()
	// Write an invalid Go template
	os.WriteFile(filepath.Join(tmplDir, "classic.html"), []byte(`{{.Invalid`), 0644)

	settingsCol := getTestCollection(t, "settings_dl_badtmpl")
	expCol := getTestCollection(t, "experiences_dl_badtmpl")

	profile := profileData{Name: "Test PM", Email: "pm@test.com"}
	handler := downloadResume(settingsCol, expCol, "http://unused", tmplDir, profile)

	req := httptest.NewRequest("GET", "/api/resume/download", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for bad template, got %d", rec.Code)
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

func TestConvertHTMLToPDF_Success(t *testing.T) {
	fakePDF := []byte("%PDF-1.4 test content")
	gotenberg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forms/chromium/convert/html" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(fakePDF)
	}))
	defer gotenberg.Close()

	result, err := convertHTMLToPDF(gotenberg.URL, []byte("<html><body>Test</body></html>"))
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !bytes.Equal(result, fakePDF) {
		t.Error("returned PDF does not match expected content")
	}
}

func TestConvertHTMLToPDF_InvalidURL(t *testing.T) {
	_, err := convertHTMLToPDF("http://127.0.0.1:1", []byte("<html></html>"))
	if err == nil {
		t.Error("expected error for unreachable gotenberg, got nil")
	}
}

func TestDownloadResume_TemplateExecuteError(t *testing.T) {
	tmplDir := t.TempDir()
	// Template that references a method on a string — will fail on Execute
	os.WriteFile(filepath.Join(tmplDir, "classic.html"), []byte(`{{.Name.Bad}}`), 0644)

	settingsCol := getTestCollection(t, "settings_dl_execerr")
	expCol := getTestCollection(t, "experiences_dl_execerr")

	profile := profileData{Name: "Test PM", Email: "pm@test.com"}
	handler := downloadResume(settingsCol, expCol, "http://unused", tmplDir, profile)

	req := httptest.NewRequest("GET", "/api/resume/download", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for template execute error, got %d", rec.Code)
	}
}
