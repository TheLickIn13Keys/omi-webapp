package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/TheLickIn13Keys/omi-webapp/internal/auth"
	"github.com/TheLickIn13Keys/omi-webapp/internal/conversations"
	"github.com/TheLickIn13Keys/omi-webapp/internal/gcp"
	"github.com/TheLickIn13Keys/omi-webapp/internal/models"
)

var (
	client                   *mongo.Client
	conversationsCollection  *mongo.Collection
	usersCollection          *mongo.Collection
	gcpCredentialsCollection *mongo.Collection
)

func main() {
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

	router.HandleFunc("/register", auth.RegisterUser(usersCollection)).Methods("POST")
	router.HandleFunc("/login", auth.LoginUser(usersCollection)).Methods("POST")
	router.HandleFunc("/logout", auth.LogoutUser).Methods("POST")
	router.HandleFunc("/conversations", auth.AuthMiddleware(conversations.GetConversations(conversationsCollection))).Methods("GET")
	router.HandleFunc("/conversations/{id}", auth.AuthMiddleware(conversations.GetConversation(conversationsCollection))).Methods("GET")
	router.HandleFunc("/conversations", auth.AuthMiddleware(conversations.CreateConversation(conversationsCollection))).Methods("POST")
	router.HandleFunc("/conversations/{id}/messages", auth.AuthMiddleware(conversations.AddMessage(conversationsCollection))).Methods("POST")
	router.HandleFunc("/conversations/{id}/transcript", auth.AuthMiddleware(conversations.UpdateTranscript(conversationsCollection))).Methods("PUT")
	router.HandleFunc("/conversations/{id}/audio", auth.AuthMiddleware(gcp.GetConversationAudio(gcpCredentialsCollection, conversationsCollection))).Methods("GET")
	router.HandleFunc("/gcp-credentials", auth.AuthMiddleware(gcp.SaveGCPCredentials(gcpCredentialsCollection))).Methods("POST")
	router.HandleFunc("/search", auth.AuthMiddleware(conversations.GlobalSearch(conversationsCollection))).Methods("GET")
	router.HandleFunc("/audio/{id}/{file}", auth.AuthMiddleware(gcp.ServeAudioFile(gcpCredentialsCollection))).Methods("GET")
	router.HandleFunc("/query-bucket", auth.AuthMiddleware(gcp.QueryBucket(gcpCredentialsCollection, conversationsCollection))).Methods("GET")
	router.HandleFunc("/upload-audio", auth.AuthMiddleware(handleAudioUpload(gcpCredentialsCollection, conversationsCollection))).Methods("POST")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Println("Server is starting on port 8080...")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", handler))
}

func handleAudioUpload(gcpCredentialsCollection, conversationsCollection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			log.Printf("Error parsing multipart form: %v", err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("Error retrieving file from form: %v", err)
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
		tempFilePath := filepath.Join("uploads", filename)

		if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
			log.Printf("Error creating uploads directory: %v", err)
			http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
			return
		}

		out, err := os.Create(tempFilePath)
		if err != nil {
			log.Printf("Error creating temporary file: %v", err)
			http.Error(w, "Unable to create the file for writing", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			log.Printf("Error writing to temporary file: %v", err)
			http.Error(w, "Error writing the file", http.StatusInternalServerError)
			return
		}

		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			log.Printf("Error getting user ID: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		gcpURL, err := gcp.UploadAudio(gcpCredentialsCollection, userID, tempFilePath, filename)
		if err != nil {
			log.Printf("Error uploading to GCP: %v", err)
			http.Error(w, fmt.Sprintf("Error uploading to GCP: %v", err), http.StatusInternalServerError)
			return
		}

		if err := os.Remove(tempFilePath); err != nil {
			log.Printf("Error deleting temporary file: %v", err)

		}

		conversation := models.Conversation{
			UserID: userID,
			Name:   header.Filename,
			AudioFile: &models.AudioFile{
				Name: filename,
				URL:  gcpURL,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result, err := conversationsCollection.InsertOne(context.TODO(), conversation)
		if err != nil {
			log.Printf("Error inserting conversation into database: %v", err)
			http.Error(w, "Error creating conversation", http.StatusInternalServerError)
			return
		}

		conversation.ID = result.InsertedID.(primitive.ObjectID)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(conversation); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
}
