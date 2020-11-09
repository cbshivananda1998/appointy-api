package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Article struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title    string             `json:"title" bson:"title"`
	SubTitle string             `json:"subtitle" bson:"subtitle"`
	Content  string             `json:"content" bson:"content"`
	Stamp    time.Time          `json:"stamp" bson:"stamp"`
}

func getAllArticles(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var articles []Article
	collection := client.Database("appointy-api").Collection("articles")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var article Article
		cursor.Decode(&article)
		articles = append(articles, article)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(articles)
}

func createArticle(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-type", "application/json")
	var article Article
	_ = json.NewDecoder(request.Body).Decode(&article)
	article.Stamp = time.Now()
	collection := client.Database("appointy-api").Collection("articles")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, article)
	json.NewEncoder(response).Encode(result)
}
func getArticleWithId(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		response.Header().Set("Content-type", "application/json")
		idt := request.URL.Path[len("/article/"):]
		id, _ := primitive.ObjectIDFromHex(idt)
		var article Article
		collection := client.Database("appointy-api").Collection("articles")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		err := collection.FindOne(ctx, Article{ID: id}).Decode(&article)
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(response).Encode(article)
	} else {
		http.Redirect(response, request, "/", http.StatusFound)
	}
}
func main() {
	fmt.Println("Started")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	http.HandleFunc("/articles", getAllArticles)
	http.HandleFunc("/article", createArticle)
	http.HandleFunc("/article/{id}", getArticleWithId)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
