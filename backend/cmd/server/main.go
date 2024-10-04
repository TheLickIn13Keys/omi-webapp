package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/TheLickIn13Keys/omi-webapp/internal/auth"
	"github.com/TheLickIn13Keys/omi-webapp/internal/conversations"
	"github.com/TheLickIn13Keys/omi-webapp/internal/gcp"
)

var (
	client                   *mongo.Client
	conversationsCollection  *mongo.Collection
	usersCollection          *mongo.Collection
	gcpCredentialsCollection *mongo.Collection
)

func main() {
	// Set up MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	conversationsCollection = client.Database("omi_friend").Collection("conversations")
	usersCollection = client.Database("omi_friend").Collection("users")
	gcpCredentialsCollection = client.Database("omi_friend").Collection("gcp_credentials")

	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/register", auth.RegisterUser(usersCollection)).Methods("POST")
	router.HandleFunc("/login", auth.LoginUser(usersCollection)).Methods("POST")
	router.HandleFunc("/logout", auth.LogoutUser).Methods("POST")
	router.HandleFunc("/conversations", auth.AuthMiddleware(conversations.GetConversations(conversationsCollection))).Methods("GET")
	router.HandleFunc("/conversations/{id}", auth.AuthMiddleware(conversations.GetConversation(conversationsCollection))).Methods("GET")
	router.HandleFunc("/conversations", auth.AuthMiddleware(conversations.CreateConversation(conversationsCollection))).Methods("POST")
	router.HandleFunc("/conversations/{id}/messages", auth.AuthMiddleware(conversations.AddMessage(conversationsCollection))).Methods("POST")
	router.HandleFunc("/conversations/{id}/transcript", auth.AuthMiddleware(conversations.UpdateTranscript(conversationsCollection))).Methods("PUT")
	router.HandleFunc("/gcp-credentials", auth.AuthMiddleware(gcp.SaveGCPCredentials(gcpCredentialsCollection))).Methods("POST")

	// Update this line to pass both collections
	router.HandleFunc("/conversations/{id}/audio", auth.AuthMiddleware(gcp.GetConversationAudio(gcpCredentialsCollection, conversationsCollection))).Methods("GET")

	router.HandleFunc("/audio/{id}/{file}", auth.AuthMiddleware(gcp.ServeAudioFile(gcpCredentialsCollection))).Methods("GET")
	router.HandleFunc("/query-bucket", auth.AuthMiddleware(gcp.QueryBucket(gcpCredentialsCollection, conversationsCollection))).Methods("GET")

	// Create a CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Wrap the router with the CORS handler
	handler := c.Handler(router)

	// Start the server
	log.Println("Server is starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
