package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ---------- Document types ----------

type experienceDoc struct {
	ID        string   `bson:"_id" json:"id"`
	Company   string   `bson:"company" json:"company"`
	Role      string   `bson:"role" json:"role"`
	Location  string   `bson:"location" json:"location"`
	StartDate string   `bson:"start_date" json:"start_date"`
	EndDate   string   `bson:"end_date" json:"end_date"`
	Bullets   []string `bson:"bullets" json:"bullets"`
}

type blogDoc struct {
	ID      string `bson:"_id" json:"id"`
	Title   string `bson:"title" json:"title"`
	Summary string `bson:"summary" json:"summary"`
	URL     string `bson:"url" json:"url"`
	Date    string `bson:"date" json:"date"`
}

type contactDoc struct {
	ID        string `bson:"_id" json:"id"`
	Name      string `bson:"name" json:"name"`
	Email     string `bson:"email" json:"email"`
	Phone     string `bson:"phone" json:"phone"`
	Subject   string `bson:"subject" json:"subject"`
	Message   string `bson:"message" json:"message"`
	CreatedAt string `bson:"created_at" json:"created_at"`
}

// ---------- Helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---------- Experience handlers ----------

func listExperiences(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cursor, err := col.Find(r.Context(), bson.M{})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var results []experienceDoc
		if err := cursor.All(r.Context(), &results); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if results == nil {
			results = []experienceDoc{}
		}
		writeJSON(w, http.StatusOK, results)
	}
}

func createExperience(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var doc experienceDoc
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		doc.ID = bson.NewObjectID().Hex()
		if _, err := col.InsertOne(r.Context(), doc); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, doc)
	}
}

func updateExperience(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var doc experienceDoc
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		doc.ID = id
		update := bson.M{"$set": bson.M{
			"company":    doc.Company,
			"role":       doc.Role,
			"location":   doc.Location,
			"start_date": doc.StartDate,
			"end_date":   doc.EndDate,
			"bullets":    doc.Bullets,
		}}
		if _, err := col.UpdateOne(r.Context(), bson.M{"_id": id}, update); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, doc)
	}
}

func deleteExperience(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := col.DeleteOne(r.Context(), bson.M{"_id": id}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ---------- Blog handlers ----------

func listBlogs(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cursor, err := col.Find(r.Context(), bson.M{})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var results []blogDoc
		if err := cursor.All(r.Context(), &results); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if results == nil {
			results = []blogDoc{}
		}
		writeJSON(w, http.StatusOK, results)
	}
}

func createBlog(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var doc blogDoc
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		doc.ID = bson.NewObjectID().Hex()
		if _, err := col.InsertOne(r.Context(), doc); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, doc)
	}
}

func updateBlog(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var doc blogDoc
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		doc.ID = id
		update := bson.M{"$set": bson.M{
			"title":   doc.Title,
			"summary": doc.Summary,
			"url":     doc.URL,
			"date":    doc.Date,
		}}
		if _, err := col.UpdateOne(r.Context(), bson.M{"_id": id}, update); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, doc)
	}
}

func deleteBlog(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := col.DeleteOne(r.Context(), bson.M{"_id": id}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ---------- Contact handlers ----------

func createContact(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var doc contactDoc
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if doc.Name == "" || doc.Email == "" || doc.Subject == "" || doc.Message == "" {
			writeError(w, http.StatusBadRequest, "name, email, subject, and message are required")
			return
		}
		doc.ID = bson.NewObjectID().Hex()
		doc.CreatedAt = time.Now().UTC().Format(time.RFC3339)
		if _, err := col.InsertOne(r.Context(), doc); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, doc)
	}
}

func listContacts(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := options.Find().SetSort(bson.M{"created_at": -1})
		cursor, err := col.Find(r.Context(), bson.M{}, opts)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var results []contactDoc
		if err := cursor.All(r.Context(), &results); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if results == nil {
			results = []contactDoc{}
		}
		writeJSON(w, http.StatusOK, results)
	}
}

func deleteContact(col *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := col.DeleteOne(r.Context(), bson.M{"_id": id}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ---------- CORS middleware ----------

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := map[string]bool{
			"http://localhost":      true,
			"http://localhost:5173": true,
		}
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---------- main ----------

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongo:27017"
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if googleClientID == "" || adminEmail == "" {
		log.Fatal("GOOGLE_CLIENT_ID and ADMIN_EMAIL env vars are required")
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Println("OPENAI_API_KEY not set; chat endpoint will be disabled")
	}

	openaiBaseURL := os.Getenv("OPENAI_BASE_URL")
	if openaiBaseURL == "" {
		openaiBaseURL = "https://api.openai.com/v1"
	}

	siteName := os.Getenv("SITE_NAME")
	if siteName == "" {
		siteName = "Swati Aggarwal"
	}

	gotenbergURL := os.Getenv("GOTENBERG_URL")
	if gotenbergURL == "" {
		gotenbergURL = "http://gotenberg:3000"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping: %v", err)
	}
	log.Println("Connected to MongoDB")

	db := client.Database("portfolio")
	expCol := db.Collection("experiences")
	blogCol := db.Collection("blogs")
	contactCol := db.Collection("contacts")
	settingsCol := db.Collection("settings")

	seedData(ctx, expCol, blogCol)

	// Load data for chat system prompt
	var systemPrompt string
	if openaiKey != "" {
		var experiences []experienceDoc
		cursor, err := expCol.Find(ctx, bson.M{})
		if err != nil {
			log.Fatalf("load experiences for chat: %v", err)
		}
		if err := cursor.All(ctx, &experiences); err != nil {
			log.Fatalf("decode experiences for chat: %v", err)
		}

		var blogs []blogDoc
		cursor, err = blogCol.Find(ctx, bson.M{})
		if err != nil {
			log.Fatalf("load blogs for chat: %v", err)
		}
		if err := cursor.All(ctx, &blogs); err != nil {
			log.Fatalf("decode blogs for chat: %v", err)
		}

		systemPrompt = buildSystemPrompt(experiences, blogs)
		log.Printf("Chat system prompt built (%d chars)", len(systemPrompt))
	}

	auth := func(h http.HandlerFunc) http.HandlerFunc {
		return adminOnly(googleClientID, adminEmail, h)
	}

	mux := http.NewServeMux()

	// Experience routes (GET is public, mutations require admin)
	mux.HandleFunc("GET /api/experiences", listExperiences(expCol))
	mux.HandleFunc("POST /api/experiences", auth(createExperience(expCol)))
	mux.HandleFunc("PUT /api/experiences/{id}", auth(updateExperience(expCol)))
	mux.HandleFunc("DELETE /api/experiences/{id}", auth(deleteExperience(expCol)))

	// Blog routes (GET is public, mutations require admin)
	mux.HandleFunc("GET /api/blogs", listBlogs(blogCol))
	mux.HandleFunc("POST /api/blogs", auth(createBlog(blogCol)))
	mux.HandleFunc("PUT /api/blogs/{id}", auth(updateBlog(blogCol)))
	mux.HandleFunc("DELETE /api/blogs/{id}", auth(deleteBlog(blogCol)))

	// Contact routes (POST is public, GET and DELETE require admin)
	mux.HandleFunc("POST /api/contacts", createContact(contactCol))
	mux.HandleFunc("GET /api/contacts", auth(listContacts(contactCol)))
	mux.HandleFunc("DELETE /api/contacts/{id}", auth(deleteContact(contactCol)))

	// Resume routes (all admin-only)
	profile := profileData{
		Name:    siteName,
		Email:   adminEmail,
		Summary: "Product leader with 15+ years building data and AI-powered platforms that turn behavioural signals into trusted, high-performing user experiences.",
	}
	mux.HandleFunc("GET /api/resume/templates", auth(listTemplates(settingsCol)))
	mux.HandleFunc("PUT /api/resume/template", auth(setActiveTemplate(settingsCol)))
	mux.HandleFunc("GET /api/resume/download", auth(downloadResume(settingsCol, expCol, gotenbergURL, profile)))

	// Chat route (public, rate-limited)
	if openaiKey != "" {
		limiter := newRateLimiter()
		mux.HandleFunc("POST /api/chat", chatHandler(openaiKey, openaiBaseURL, systemPrompt, limiter))
		log.Println("Chat endpoint enabled")
	}

	addr := ":8080"
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(h2c.NewHandler(mux, &http2.Server{}))); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
