package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var db *dynamodb.DynamoDB
var tableName string

func init() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	}))
	db = dynamodb.New(sess)
	tableName = os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "Items"
	}
}

func putItemHandler(w http.ResponseWriter, r *http.Request) {
	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"id":   {S: aws.String(item.ID)},
			"name": {S: aws.String(item.Name)},
		},
	}

	_, err = db.PutItem(input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to put item: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item created"})
}

func getItemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' query parameter", http.StatusBadRequest)
		return
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
	}

	result, err := db.GetItem(input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get item: %s", err), http.StatusInternalServerError)
		return
	}

	if result.Item == nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	item := Item{
		ID:   *result.Item["id"].S,
		Name: *result.Item["name"].S,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	_, err := db.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		http.Error(w, `{"status": "unhealthy", "error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	http.HandleFunc("/item", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			putItemHandler(w, r)
		} else if r.Method == http.MethodGet {
			getItemHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/healthcheck", healthCheckHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("DynamoDB API Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

