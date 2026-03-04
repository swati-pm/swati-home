package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ---------- Document types ----------

type settingsDoc struct {
	ID             string `bson:"_id" json:"id"`
	ActiveTemplate string `bson:"active_template" json:"active_template"`
}

type profileData struct {
	Name        string
	Email       string
	Summary     string
	Experiences []experienceDoc
}

type templateInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

var templates = []templateInfo{
	{Name: "classic", Description: "Traditional single-column layout with serif font. Clean and conservative."},
	{Name: "modern", Description: "Contemporary design with accent color sidebar and sans-serif font."},
	{Name: "compact", Description: "Dense layout for experienced PMs. Fits more roles on fewer pages."},
}

// ---------- Resume handlers ----------

func listTemplates(settingsCol *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var settings settingsDoc
		err := settingsCol.FindOne(r.Context(), bson.M{"_id": "global"}).Decode(&settings)
		if err != nil {
			settings.ActiveTemplate = "classic"
		}

		result := make([]templateInfo, len(templates))
		for i, t := range templates {
			result[i] = t
			result[i].Active = t.Name == settings.ActiveTemplate
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func setActiveTemplate(settingsCol *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Template string `json:"template"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Validate template name
		valid := false
		for _, t := range templates {
			if t.Name == body.Template {
				valid = true
				break
			}
		}
		if !valid {
			writeError(w, http.StatusBadRequest, "invalid template name")
			return
		}

		opts := options.UpdateOne().SetUpsert(true)
		_, err := settingsCol.UpdateOne(
			r.Context(),
			bson.M{"_id": "global"},
			bson.M{"$set": bson.M{"active_template": body.Template}},
			opts,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"active_template": body.Template})
	}
}

func downloadResume(settingsCol, expCol *mongo.Collection, gotenbergURL, templateDir string, profile profileData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get active template
		var settings settingsDoc
		err := settingsCol.FindOne(r.Context(), bson.M{"_id": "global"}).Decode(&settings)
		if err != nil {
			settings.ActiveTemplate = "classic"
		}

		// 2. Fetch experiences
		cursor, err := expCol.Find(r.Context(), bson.M{})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fetch experiences")
			return
		}
		var experiences []experienceDoc
		if err := cursor.All(r.Context(), &experiences); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to decode experiences")
			return
		}

		profile.Experiences = experiences

		// 3. Load and render HTML template
		tmplPath := filepath.Join(templateDir, settings.ActiveTemplate+".html")
		tmplBytes, err := os.ReadFile(tmplPath)
		if err != nil {
			log.Printf("resume: failed to read template %s: %v", tmplPath, err)
			writeError(w, http.StatusInternalServerError, "template not found")
			return
		}

		tmpl, err := template.New("resume").Parse(string(tmplBytes))
		if err != nil {
			log.Printf("resume: failed to parse template: %v", err)
			writeError(w, http.StatusInternalServerError, "template parse error")
			return
		}

		var htmlBuf bytes.Buffer
		if err := tmpl.Execute(&htmlBuf, profile); err != nil {
			log.Printf("resume: failed to execute template: %v", err)
			writeError(w, http.StatusInternalServerError, "template render error")
			return
		}

		// 4. Send to Gotenberg
		pdfBytes, err := convertHTMLToPDF(gotenbergURL, htmlBuf.Bytes())
		if err != nil {
			log.Printf("resume: gotenberg error: %v", err)
			writeError(w, http.StatusBadGateway, "PDF generation failed")
			return
		}

		// 5. Return PDF
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-resume.pdf"`, profile.Name))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
		w.Write(pdfBytes)
	}
}

// convertHTMLToPDF sends HTML to Gotenberg and returns the PDF bytes.
func convertHTMLToPDF(gotenbergURL string, html []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(html); err != nil {
		return nil, fmt.Errorf("write html: %w", err)
	}
	writer.Close()

	resp, err := http.Post(
		gotenbergURL+"/forms/chromium/convert/html",
		writer.FormDataContentType(),
		&buf,
	)
	if err != nil {
		return nil, fmt.Errorf("gotenberg request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg returned %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
