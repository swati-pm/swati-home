package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type seedExperience struct {
	ID        string   `json:"id"`
	Company   string   `json:"company"`
	Role      string   `json:"role"`
	Location  string   `json:"location"`
	StartDate string   `json:"startDate"`
	EndDate   string   `json:"endDate"`
	Bullets   []string `json:"bullets"`
}

type seedBlog struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	URL     string `json:"url"`
	Date    string `json:"date"`
}

func seedData(ctx context.Context, expCol, blogCol *mongo.Collection) {
	seedExperiences(ctx, expCol)
	seedBlogs(ctx, blogCol)
}

func seedExperiences(ctx context.Context, col *mongo.Collection) {
	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("seed: count experiences: %v", err)
		return
	}
	if count > 0 {
		log.Printf("seed: experiences already has %d docs, skipping", count)
		return
	}

	data, err := os.ReadFile("/app/data/experience.json")
	if err != nil {
		log.Printf("seed: read experience.json: %v", err)
		return
	}

	var items []seedExperience
	if err := json.Unmarshal(data, &items); err != nil {
		log.Printf("seed: unmarshal experiences: %v", err)
		return
	}

	var docs []interface{}
	for _, item := range items {
		docs = append(docs, experienceDoc{
			ID:        item.ID,
			Company:   item.Company,
			Role:      item.Role,
			Location:  item.Location,
			StartDate: item.StartDate,
			EndDate:   item.EndDate,
			Bullets:   item.Bullets,
		})
	}

	if _, err := col.InsertMany(ctx, docs); err != nil {
		log.Printf("seed: insert experiences: %v", err)
		return
	}
	log.Printf("seed: inserted %d experiences", len(docs))
}

func seedBlogs(ctx context.Context, col *mongo.Collection) {
	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("seed: count blogs: %v", err)
		return
	}
	if count > 0 {
		log.Printf("seed: blogs already has %d docs, skipping", count)
		return
	}

	data, err := os.ReadFile("/app/data/blogs.json")
	if err != nil {
		log.Printf("seed: read blogs.json: %v", err)
		return
	}

	var items []seedBlog
	if err := json.Unmarshal(data, &items); err != nil {
		log.Printf("seed: unmarshal blogs: %v", err)
		return
	}

	var docs []interface{}
	for _, item := range items {
		docs = append(docs, blogDoc{
			ID:      item.ID,
			Title:   item.Title,
			Summary: item.Summary,
			URL:     item.URL,
			Date:    item.Date,
		})
	}

	if _, err := col.InsertMany(ctx, docs); err != nil {
		log.Printf("seed: insert blogs: %v", err)
		return
	}
	log.Printf("seed: inserted %d blogs", len(docs))
}
