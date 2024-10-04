package gcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/TheLickIn13Keys/omi-webapp/internal/auth"
	"github.com/TheLickIn13Keys/omi-webapp/internal/models"
	"github.com/TheLickIn13Keys/omi-webapp/internal/transcription"
)

func SaveGCPCredentials(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var creds models.GCPCredentials
		_ = json.NewDecoder(r.Body).Decode(&creds)

		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		creds.UserID = userID

		_, err = collection.UpdateOne(
			context.TODO(),
			bson.M{"user_id": userID},
			bson.M{"$set": creds},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			http.Error(w, "Error saving GCP credentials", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "GCP credentials saved successfully"})
	}
}

func GetConversationAudio(gcpCollection, conversationsCollection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		conversationID, _ := primitive.ObjectIDFromHex(params["id"])

		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch the conversation
		var conversation models.Conversation
		err = conversationsCollection.FindOne(context.TODO(), bson.M{"_id": conversationID, "user_id": userID}).Decode(&conversation)
		if err != nil {
			http.Error(w, "Conversation not found", http.StatusNotFound)
			return
		}

		// If the conversation doesn't have an audio file, return an empty response
		if conversation.AudioFile == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"audio_file": nil,
			})
			return
		}

		// Fetch GCP credentials for the user
		var creds models.GCPCredentials
		err = gcpCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&creds)
		if err != nil {
			log.Printf("Error fetching GCP credentials: %v", err)
			http.Error(w, "GCP credentials not found", http.StatusNotFound)
			return
		}

		// Decode base64 credentials
		jsonCreds, err := base64.StdEncoding.DecodeString(creds.Credentials)
		if err != nil {
			log.Printf("Error decoding GCP credentials: %v", err)
			http.Error(w, "Invalid GCP credentials", http.StatusInternalServerError)
			return
		}

		// Parse the JSON credentials
		var parsedCreds struct {
			ClientEmail string `json:"client_email"`
			PrivateKey  string `json:"private_key"`
		}
		if err := json.Unmarshal(jsonCreds, &parsedCreds); err != nil {
			log.Printf("Error parsing GCP credentials: %v", err)
			http.Error(w, "Invalid GCP credentials format", http.StatusInternalServerError)
			return
		}

		// Generate a signed URL for the audio file
		url, err := storage.SignedURL(creds.BucketName, conversation.AudioFile.Name, &storage.SignedURLOptions{
			GoogleAccessID: parsedCreds.ClientEmail,
			PrivateKey:     []byte(parsedCreds.PrivateKey),
			Method:         "GET",
			Expires:        time.Now().Add(15 * time.Minute),
		})
		if err != nil {
			log.Printf("Error generating signed URL: %v", err)
			http.Error(w, "Failed to generate signed URL", http.StatusInternalServerError)
			return
		}

		// Update the URL in the response
		conversation.AudioFile.URL = url

		json.NewEncoder(w).Encode(map[string]interface{}{
			"audio_file": conversation.AudioFile,
		})
	}
}
func QueryBucket(gcpCollection, conversationsCollection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch GCP credentials for the user
		var creds models.GCPCredentials
		err = gcpCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&creds)
		if err != nil {
			http.Error(w, "GCP credentials not found", http.StatusNotFound)
			return
		}

		// Decode base64 credentials
		jsonCreds, err := base64.StdEncoding.DecodeString(creds.Credentials)
		if err != nil {
			http.Error(w, "Invalid GCP credentials", http.StatusInternalServerError)
			return
		}

		// Create GCP storage client
		ctx := context.Background()
		client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonCreds))
		if err != nil {
			http.Error(w, "Failed to create GCP storage client", http.StatusInternalServerError)
			return
		}
		defer client.Close()

		// List objects in the bucket
		bucket := client.Bucket(creds.BucketName)
		it := bucket.Objects(ctx, nil)

		newConversations := []models.Conversation{}

		for {
			attrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				http.Error(w, "Error listing bucket objects", http.StatusInternalServerError)
				return
			}

			// Check if a conversation already exists for this file
			var existingConversation models.Conversation
			err = conversationsCollection.FindOne(context.TODO(), bson.M{"user_id": userID, "audio_file.name": attrs.Name}).Decode(&existingConversation)
			if err == mongo.ErrNoDocuments {
				// Create a new conversation for this file immediately
				newConversation := models.Conversation{
					UserID: userID,
					Name:   strings.TrimSuffix(filepath.Base(attrs.Name), filepath.Ext(attrs.Name)),
					AudioFile: &models.AudioFile{
						Name: attrs.Name,
						URL:  "", // We'll generate this URL when needed
					},
					Transcript: []models.TranscriptionSentence{{Sentence: "Processing transcription..."}},
					CreatedAt:  attrs.Created,
					UpdatedAt:  attrs.Updated,
				}

				result, err := conversationsCollection.InsertOne(context.TODO(), newConversation)
				if err != nil {
					http.Error(w, "Error creating new conversation", http.StatusInternalServerError)
					return
				}

				newConversation.ID = result.InsertedID.(primitive.ObjectID)
				newConversations = append(newConversations, newConversation)

				// Start transcription in a goroutine
				go func(convID primitive.ObjectID) {
					// Generate a signed URL for the audio file
					signedURL, err := generateSignedURL(bucket, attrs.Name, creds)
					if err != nil {
						log.Printf("Error generating signed URL: %v", err)
						return
					}

					// Transcribe the audio
					transcript, err := transcription.TranscribeAudio(signedURL, creds.GladiaKey)
					if err != nil {
						log.Printf("Error transcribing audio: %v", err)
						transcript = []models.TranscriptionSentence{{Sentence: "Error transcribing audio"}}
					}

					// Update the conversation with the transcription
					_, err = conversationsCollection.UpdateOne(
						context.TODO(),
						bson.M{"_id": convID},
						bson.M{"$set": bson.M{"transcript": transcript, "updated_at": time.Now()}},
					)
					if err != nil {
						log.Printf("Error updating conversation with transcription: %v", err)
					}
				}(newConversation.ID)
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"new_conversations": newConversations,
		})
	}
}

func generateSignedURL(bucket *storage.BucketHandle, objectName string, creds models.GCPCredentials) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		Expires:        time.Now().Add(15 * time.Minute),
		GoogleAccessID: creds.ClientEmail,
		PrivateKey:     []byte(creds.PrivateKey),
	}
	return bucket.SignedURL(objectName, opts)
}

func ServeAudioFile(gcpCollection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		conversationID := params["id"]
		fileName := params["file"]

		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch GCP credentials for the user
		var creds models.GCPCredentials
		err = gcpCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&creds)
		if err != nil {
			http.Error(w, "GCP credentials not found", http.StatusNotFound)
			return
		}

		// Decode base64 credentials
		jsonCreds, err := base64.StdEncoding.DecodeString(creds.Credentials)
		if err != nil {
			http.Error(w, "Invalid GCP credentials", http.StatusInternalServerError)
			return
		}

		// Create GCP storage client
		ctx := context.Background()
		client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonCreds))
		if err != nil {
			http.Error(w, "Failed to create GCP storage client", http.StatusInternalServerError)
			return
		}
		defer client.Close()

		// Get the object from the bucket
		bucket := client.Bucket(creds.BucketName)
		obj := bucket.Object(fmt.Sprintf("%s/%s", conversationID, fileName))
		reader, err := obj.NewReader(ctx)
		if err != nil {
			http.Error(w, "Failed to read audio file", http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		// Set appropriate headers
		w.Header().Set("Content-Type", "audio/mpeg") // Adjust content type as needed
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

		// Stream the file to the response
		_, err = io.Copy(w, reader)
		if err != nil {
			http.Error(w, "Failed to stream audio file", http.StatusInternalServerError)
			return
		}
	}
}
