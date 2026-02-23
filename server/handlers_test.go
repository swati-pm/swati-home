package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// getTestCollection connects to MongoDB and returns a test collection.
// Skips the test if MONGO_URI is not set or connection fails.
func getTestCollection(t *testing.T, name string) *mongo.Collection {
	t.Helper()

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Skipf("skipping: cannot connect to MongoDB: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("skipping: cannot ping MongoDB: %v", err)
	}

	col := client.Database("portfolio_test").Collection(name)

	// Clean collection before test
	col.DeleteMany(context.Background(), bson.M{})

	t.Cleanup(func() {
		col.DeleteMany(context.Background(), bson.M{})
		client.Disconnect(context.Background())
	})

	return col
}

// ---------- Experience handler integration tests ----------

func TestListExperiences_Empty(t *testing.T) {
	col := getTestCollection(t, "experiences")
	handler := listExperiences(col)

	req := httptest.NewRequest("GET", "/api/experiences", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []experienceDoc
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d items", len(result))
	}
}

func TestCreateExperience(t *testing.T) {
	col := getTestCollection(t, "experiences")
	handler := createExperience(col)

	exp := experienceDoc{
		Company:  "TestCo",
		Role:     "Engineer",
		Location: "NYC",
		Bullets:  []string{"Built things"},
	}
	body, _ := json.Marshal(exp)

	req := httptest.NewRequest("POST", "/api/experiences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created experienceDoc
	json.NewDecoder(rec.Body).Decode(&created)
	if created.Company != "TestCo" {
		t.Errorf("expected company 'TestCo', got %q", created.Company)
	}
	if created.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateAndListExperiences(t *testing.T) {
	col := getTestCollection(t, "experiences")

	// Create
	exp := experienceDoc{Company: "ACME", Role: "PM", Location: "London", Bullets: []string{"Led team"}}
	body, _ := json.Marshal(exp)
	req := httptest.NewRequest("POST", "/api/experiences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	createExperience(col)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rec.Code)
	}

	// List
	req = httptest.NewRequest("GET", "/api/experiences", nil)
	rec = httptest.NewRecorder()
	listExperiences(col)(rec, req)

	var result []experienceDoc
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 experience, got %d", len(result))
	}
	if result[0].Company != "ACME" {
		t.Errorf("expected company 'ACME', got %q", result[0].Company)
	}
}

func TestUpdateExperience(t *testing.T) {
	col := getTestCollection(t, "experiences")

	// Insert directly
	doc := experienceDoc{ID: "test-update-id", Company: "Old", Role: "Old Role"}
	col.InsertOne(context.Background(), doc)

	// Update
	updated := experienceDoc{Company: "New", Role: "New Role", Location: "Berlin", Bullets: []string{}}
	body, _ := json.Marshal(updated)
	req := httptest.NewRequest("PUT", "/api/experiences/test-update-id", bytes.NewReader(body))
	req.SetPathValue("id", "test-update-id")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	updateExperience(col)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result experienceDoc
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Company != "New" {
		t.Errorf("expected 'New', got %q", result.Company)
	}
}

func TestDeleteExperience(t *testing.T) {
	col := getTestCollection(t, "experiences")

	// Insert
	doc := experienceDoc{ID: "del-id", Company: "ToDelete"}
	col.InsertOne(context.Background(), doc)

	// Delete
	req := httptest.NewRequest("DELETE", "/api/experiences/del-id", nil)
	req.SetPathValue("id", "del-id")
	rec := httptest.NewRecorder()
	deleteExperience(col)(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	// Verify deleted
	count, _ := col.CountDocuments(context.Background(), bson.M{"_id": "del-id"})
	if count != 0 {
		t.Error("expected document to be deleted")
	}
}

func TestCreateExperience_InvalidJSON(t *testing.T) {
	col := getTestCollection(t, "experiences")
	handler := createExperience(col)

	req := httptest.NewRequest("POST", "/api/experiences", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ---------- Blog handler integration tests ----------

func TestListBlogs_Empty(t *testing.T) {
	col := getTestCollection(t, "blogs")
	handler := listBlogs(col)

	req := httptest.NewRequest("GET", "/api/blogs", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []blogDoc
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d items", len(result))
	}
}

func TestCreateAndListBlogs(t *testing.T) {
	col := getTestCollection(t, "blogs")

	// Create
	blog := blogDoc{Title: "Test Post", Summary: "A summary", URL: "https://test.com", Date: "2025-01-15"}
	body, _ := json.Marshal(blog)
	req := httptest.NewRequest("POST", "/api/blogs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	createBlog(col)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// List
	req = httptest.NewRequest("GET", "/api/blogs", nil)
	rec = httptest.NewRecorder()
	listBlogs(col)(rec, req)

	var result []blogDoc
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 blog, got %d", len(result))
	}
	if result[0].Title != "Test Post" {
		t.Errorf("expected 'Test Post', got %q", result[0].Title)
	}
}

func TestUpdateBlog(t *testing.T) {
	col := getTestCollection(t, "blogs")

	doc := blogDoc{ID: "blog-upd", Title: "Old Title", Summary: "Old"}
	col.InsertOne(context.Background(), doc)

	updated := blogDoc{Title: "New Title", Summary: "New Summary", URL: "https://new.com", Date: "2025-06-01"}
	body, _ := json.Marshal(updated)
	req := httptest.NewRequest("PUT", "/api/blogs/blog-upd", bytes.NewReader(body))
	req.SetPathValue("id", "blog-upd")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	updateBlog(col)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDeleteBlog(t *testing.T) {
	col := getTestCollection(t, "blogs")

	doc := blogDoc{ID: "blog-del", Title: "Delete Me"}
	col.InsertOne(context.Background(), doc)

	req := httptest.NewRequest("DELETE", "/api/blogs/blog-del", nil)
	req.SetPathValue("id", "blog-del")
	rec := httptest.NewRecorder()
	deleteBlog(col)(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

// ---------- Contact handler integration tests ----------

func TestCreateContact_Valid(t *testing.T) {
	col := getTestCollection(t, "contacts")
	handler := createContact(col)

	contact := contactDoc{
		Name:    "Jane",
		Email:   "jane@test.com",
		Subject: "Hello",
		Message: "Test message",
	}
	body, _ := json.Marshal(contact)

	req := httptest.NewRequest("POST", "/api/contacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created contactDoc
	json.NewDecoder(rec.Body).Decode(&created)
	if created.Name != "Jane" {
		t.Errorf("expected name 'Jane', got %q", created.Name)
	}
	if created.CreatedAt == "" {
		t.Error("expected created_at to be set")
	}
}

func TestCreateContact_MissingFields(t *testing.T) {
	col := getTestCollection(t, "contacts")
	handler := createContact(col)

	contact := contactDoc{Name: "Jane", Email: "jane@test.com"}
	body, _ := json.Marshal(contact)

	req := httptest.NewRequest("POST", "/api/contacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if result["error"] != "name, email, subject, and message are required" {
		t.Errorf("unexpected error: %q", result["error"])
	}
}

func TestListContacts(t *testing.T) {
	col := getTestCollection(t, "contacts")

	// Insert
	doc := contactDoc{ID: "c1", Name: "Jane", Email: "j@t.com", Subject: "Hi", Message: "Hello", CreatedAt: "2025-01-15T10:00:00Z"}
	col.InsertOne(context.Background(), doc)

	req := httptest.NewRequest("GET", "/api/contacts", nil)
	rec := httptest.NewRecorder()
	listContacts(col)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []contactDoc
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(result))
	}
}

func TestDeleteContact(t *testing.T) {
	col := getTestCollection(t, "contacts")

	doc := contactDoc{ID: "cdel", Name: "Delete"}
	col.InsertOne(context.Background(), doc)

	req := httptest.NewRequest("DELETE", "/api/contacts/cdel", nil)
	req.SetPathValue("id", "cdel")
	rec := httptest.NewRecorder()
	deleteContact(col)(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
