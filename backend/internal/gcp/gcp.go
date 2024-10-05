package gcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
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

		var conversation models.Conversation
		err = conversationsCollection.FindOne(context.TODO(), bson.M{"_id": conversationID, "user_id": userID}).Decode(&conversation)
		if err != nil {
			http.Error(w, "Conversation not found", http.StatusNotFound)
			return
		}

		if conversation.AudioFile == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"audio_file": nil,
			})
			return
		}

		var creds models.GCPCredentials
		err = gcpCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&creds)
		if err != nil {
			log.Printf("Error fetching GCP credentials: %v", err)
			http.Error(w, "GCP credentials not found", http.StatusNotFound)
			return
		}

		jsonCreds, err := base64.StdEncoding.DecodeString(creds.Credentials)
		if err != nil {
			log.Printf("Error decoding GCP credentials: %v", err)
			http.Error(w, "Invalid GCP credentials", http.StatusInternalServerError)
			return
		}

		var parsedCreds struct {
			ClientEmail string `json:"client_email"`
			PrivateKey  string `json:"private_key"`
		}
		if err := json.Unmarshal(jsonCreds, &parsedCreds); err != nil {
			log.Printf("Error parsing GCP credentials: %v", err)
			http.Error(w, "Invalid GCP credentials format", http.StatusInternalServerError)
			return
		}

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

		var creds models.GCPCredentials
		err = gcpCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&creds)
		if err != nil {
			http.Error(w, "GCP credentials not found", http.StatusNotFound)
			return
		}

		jsonCreds, err := base64.StdEncoding.DecodeString(creds.Credentials)
		if err != nil {
			http.Error(w, "Invalid GCP credentials", http.StatusInternalServerError)
			return
		}

		var parsedCreds struct {
			ClientEmail string `json:"client_email"`
			PrivateKey  string `json:"private_key"`
		}
		if err := json.Unmarshal(jsonCreds, &parsedCreds); err != nil {
			http.Error(w, "Invalid GCP credentials format", http.StatusInternalServerError)
			return
		}

		ctx := context.Background()
		client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonCreds))
		if err != nil {
			http.Error(w, "Failed to create GCP storage client", http.StatusInternalServerError)
			return
		}
		defer client.Close()

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

			var existingConversation models.Conversation
			err = conversationsCollection.FindOne(context.TODO(), bson.M{"user_id": userID, "audio_file.name": attrs.Name}).Decode(&existingConversation)
			if err == mongo.ErrNoDocuments {

				newConversation := models.Conversation{
					UserID: userID,
					Name:   strings.TrimSuffix(filepath.Base(attrs.Name), filepath.Ext(attrs.Name)),
					AudioFile: &models.AudioFile{
						Name: attrs.Name,
						URL:  "",
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

				go initiateTranscription(conversationsCollection, gcpCollection, newConversation.ID, jsonCreds, creds)
			} else if err == nil {

				if len(existingConversation.Transcript) == 0 || (len(existingConversation.Transcript) == 1 && existingConversation.Transcript[0].Sentence == "Processing transcription...") {
					go initiateTranscription(conversationsCollection, gcpCollection, existingConversation.ID, jsonCreds, creds)
				}
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"new_conversations": newConversations,
		})
	}
}

func initiateTranscription(conversationsCollection, gcpCollection *mongo.Collection, conversationID primitive.ObjectID, jsonCreds []byte, creds models.GCPCredentials) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		var conversation models.Conversation
		err := conversationsCollection.FindOne(context.TODO(), bson.M{"_id": conversationID}).Decode(&conversation)
		if err != nil {
			log.Printf("Error fetching conversation: %v", err)
			return
		}

		signedURL, err := generateSignedURL(jsonCreds, creds.BucketName, conversation.AudioFile.Name)
		if err != nil {
			log.Printf("Error generating signed URL: %v", err)
			return
		}

		transcript, summary, actionItems, err := transcription.TranscribeAudio(signedURL, creds.GladiaKey)
		if err != nil {
			log.Printf("Error transcribing audio (attempt %d): %v", i+1, err)
			if i == maxRetries-1 {
				transcript = []models.TranscriptionSentence{{Sentence: "Error transcribing audio after multiple attempts"}}
				summary = "Error generating summary"
				actionItems = []string{"Error generating action items"}
			} else {
				time.Sleep(time.Duration(i+1) * 5 * time.Second)
				continue
			}
		}

		_, err = conversationsCollection.UpdateOne(
			context.TODO(),
			bson.M{"_id": conversationID},
			bson.M{"$set": bson.M{
				"transcript":   transcript,
				"summary":      summary,
				"action_items": actionItems,
				"updated_at":   time.Now(),
			}},
		)
		if err != nil {
			log.Printf("Error updating conversation with transcription: %v", err)
		}

		if err == nil {
			break
		}
	}
}

func UploadAudio(gcpCredentialsCollection *mongo.Collection, userID primitive.ObjectID, localFilePath, filename string) (string, error) {

	var gcpCreds models.GCPCredentials
	err := gcpCredentialsCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&gcpCreds)
	if err != nil {
		return "", fmt.Errorf("error fetching GCP credentials: %v", err)
	}

	jsonCreds, err := base64.StdEncoding.DecodeString(gcpCreds.Credentials)
	if err != nil {
		return "", fmt.Errorf("error decoding GCP credentials: %v", err)
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonCreds))
	if err != nil {
		return "", fmt.Errorf("failed to create GCP storage client: %v", err)
	}
	defer client.Close()

	f, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	bucket := client.Bucket(gcpCreds.BucketName)
	obj := bucket.Object(filename)
	wc := obj.NewWriter(ctx)
	wc.ContentType = contentType
	if _, err = io.Copy(wc, f); err != nil {
		return "", fmt.Errorf("failed to copy file to GCP: %v", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCP writer: %v", err)
	}

	url, err := generateSignedURL(jsonCreds, gcpCreds.BucketName, filename)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %v", err)
	}

	return url, nil
}

func generateSignedURL(jsonCreds []byte, bucketName, objectName string) (string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonCreds))
	if err != nil {
		return "", fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	}

	url, err := client.Bucket(bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %v", err)
	}

	return url, nil
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

		var creds models.GCPCredentials
		err = gcpCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&creds)
		if err != nil {
			http.Error(w, "GCP credentials not found", http.StatusNotFound)
			return
		}

		jsonCreds, err := base64.StdEncoding.DecodeString(creds.Credentials)
		if err != nil {
			http.Error(w, "Invalid GCP credentials", http.StatusInternalServerError)
			return
		}

		ctx := context.Background()
		client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonCreds))
		if err != nil {
			http.Error(w, "Failed to create GCP storage client", http.StatusInternalServerError)
			return
		}
		defer client.Close()

		bucket := client.Bucket(creds.BucketName)
		obj := bucket.Object(fmt.Sprintf("%s/%s", conversationID, fileName))
		reader, err := obj.NewReader(ctx)
		if err != nil {
			http.Error(w, "Failed to read audio file", http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

		_, err = io.Copy(w, reader)
		if err != nil {
			http.Error(w, "Failed to stream audio file", http.StatusInternalServerError)
			return
		}
	}
}
