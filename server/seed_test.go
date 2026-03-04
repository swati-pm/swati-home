package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSeedExperiences_SkipsWhenDataExists(t *testing.T) {
	col := getTestCollection(t, "seed_exp_skip")
	ctx := context.Background()

	col.InsertOne(ctx, experienceDoc{
		ID:      "existing-exp",
		Company: "Existing Corp",
		Role:    "Engineer",
	})

	countBefore, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("count before: %v", err)
	}

	seedExperiences(ctx, col, "/nonexistent")

	countAfter, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("count after: %v", err)
	}

	if countAfter != countBefore {
		t.Errorf("expected count to remain %d, got %d", countBefore, countAfter)
	}
}

func TestSeedExperiences_NoFile(t *testing.T) {
	col := getTestCollection(t, "seed_exp_nofile")
	ctx := context.Background()

	seedExperiences(ctx, col, "/nonexistent")

	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 documents, got %d", count)
	}
}

func TestSeedExperiences_WithData(t *testing.T) {
	col := getTestCollection(t, "seed_exp_data")
	ctx := context.Background()

	dataDir := t.TempDir()
	experiences := []seedExperience{
		{ID: "e1", Company: "TestCo", Role: "PM", Location: "NYC", StartDate: "Jan 2020", EndDate: "Present", Bullets: []string{"Led things"}},
		{ID: "e2", Company: "ACME", Role: "Engineer", Location: "SF", StartDate: "Jan 2018", EndDate: "Dec 2019", Bullets: []string{"Built things"}},
	}
	data, _ := json.Marshal(experiences)
	os.WriteFile(filepath.Join(dataDir, "experience.json"), data, 0644)

	seedExperiences(ctx, col, dataDir)

	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 documents, got %d", count)
	}
}

func TestSeedExperiences_InvalidJSON(t *testing.T) {
	col := getTestCollection(t, "seed_exp_badjson")
	ctx := context.Background()

	dataDir := t.TempDir()
	os.WriteFile(filepath.Join(dataDir, "experience.json"), []byte("not json"), 0644)

	seedExperiences(ctx, col, dataDir)

	count, _ := col.CountDocuments(ctx, bson.M{})
	if count != 0 {
		t.Errorf("expected 0 documents after bad JSON, got %d", count)
	}
}

func TestSeedBlogs_SkipsWhenDataExists(t *testing.T) {
	col := getTestCollection(t, "seed_blog_skip")
	ctx := context.Background()

	col.InsertOne(ctx, blogDoc{
		ID:    "existing-blog",
		Title: "Existing Post",
	})

	countBefore, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("count before: %v", err)
	}

	seedBlogs(ctx, col, "/nonexistent")

	countAfter, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("count after: %v", err)
	}

	if countAfter != countBefore {
		t.Errorf("expected count to remain %d, got %d", countBefore, countAfter)
	}
}

func TestSeedBlogs_NoFile(t *testing.T) {
	col := getTestCollection(t, "seed_blog_nofile")
	ctx := context.Background()

	seedBlogs(ctx, col, "/nonexistent")

	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 documents, got %d", count)
	}
}

func TestSeedBlogs_WithData(t *testing.T) {
	col := getTestCollection(t, "seed_blog_data")
	ctx := context.Background()

	dataDir := t.TempDir()
	blogs := []seedBlog{
		{ID: "b1", Title: "Post 1", Summary: "Summary 1", URL: "https://test.com/1", Date: "2025-01-15"},
	}
	data, _ := json.Marshal(blogs)
	os.WriteFile(filepath.Join(dataDir, "blogs.json"), data, 0644)

	seedBlogs(ctx, col, dataDir)

	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 document, got %d", count)
	}
}

func TestSeedBlogs_InvalidJSON(t *testing.T) {
	col := getTestCollection(t, "seed_blog_badjson")
	ctx := context.Background()

	dataDir := t.TempDir()
	os.WriteFile(filepath.Join(dataDir, "blogs.json"), []byte("not json"), 0644)

	seedBlogs(ctx, col, dataDir)

	count, _ := col.CountDocuments(ctx, bson.M{})
	if count != 0 {
		t.Errorf("expected 0 documents after bad JSON, got %d", count)
	}
}

func TestSeedData(t *testing.T) {
	expCol := getTestCollection(t, "seed_data_exp")
	blogCol := getTestCollection(t, "seed_data_blog")
	ctx := context.Background()

	dataDir := t.TempDir()
	experiences := []seedExperience{{ID: "e1", Company: "Co", Role: "PM"}}
	blogs := []seedBlog{{ID: "b1", Title: "Blog", Summary: "S", URL: "u", Date: "d"}}
	expData, _ := json.Marshal(experiences)
	blogData, _ := json.Marshal(blogs)
	os.WriteFile(filepath.Join(dataDir, "experience.json"), expData, 0644)
	os.WriteFile(filepath.Join(dataDir, "blogs.json"), blogData, 0644)

	seedData(ctx, expCol, blogCol, dataDir)

	expCount, _ := expCol.CountDocuments(ctx, bson.M{})
	blogCount, _ := blogCol.CountDocuments(ctx, bson.M{})

	if expCount != 1 {
		t.Errorf("expected 1 experience, got %d", expCount)
	}
	if blogCount != 1 {
		t.Errorf("expected 1 blog, got %d", blogCount)
	}
}
