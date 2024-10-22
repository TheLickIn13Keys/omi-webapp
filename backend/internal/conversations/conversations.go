package conversations

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/TheLickIn13Keys/omi-webapp/internal/auth"
	"github.com/TheLickIn13Keys/omi-webapp/internal/models"
	"github.com/TheLickIn13Keys/omi-webapp/internal/transcription"
)

func GetConversations(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var conversations []models.Conversation
		cursor, err := collection.Find(context.TODO(), bson.M{"user_id": userID})
		if err != nil {
			http.Error(w, "Error fetching conversations", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			var conversation models.Conversation
			cursor.Decode(&conversation)
			conversations = append(conversations, conversation)
		}
		json.NewEncoder(w).Encode(conversations)
	}
}

func GetConversation(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		id, _ := primitive.ObjectIDFromHex(params["id"])
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var conversation models.Conversation
		err = collection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userID}).Decode(&conversation)
		if err != nil {
			http.Error(w, "Conversation not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(conversation)
	}
}

func CreateConversation(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var conversation models.Conversation
		_ = json.NewDecoder(r.Body).Decode(&conversation)
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conversation.UserID = userID
		conversation.CreatedAt = time.Now()
		conversation.UpdatedAt = time.Now()

		result, err := collection.InsertOne(context.TODO(), conversation)
		if err != nil {
			http.Error(w, "Error creating conversation", http.StatusInternalServerError)
			return
		}
		conversation.ID = result.InsertedID.(primitive.ObjectID)
		json.NewEncoder(w).Encode(conversation)
	}
}

func AddMessage(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		conversationID, _ := primitive.ObjectIDFromHex(params["id"])
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var message models.ChatMessage
		_ = json.NewDecoder(r.Body).Decode(&message)
		message.UserID = userID
		message.Timestamp = time.Now()

		update := bson.M{
			"$push": bson.M{"chat_history": message},
			"$set":  bson.M{"updated_at": time.Now()},
		}

		var conversation models.Conversation
		err = collection.FindOne(context.TODO(), bson.M{"_id": conversationID, "user_id": userID}).Decode(&conversation)
		if err == nil && len(conversation.ChatHistory) == 0 && conversation.AudioFile != nil {

			var gcpCreds models.GCPCredentials
			gcpCollection := collection.Database().Collection("gcp_credentials")
			err = gcpCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&gcpCreds)
			if err == nil {
				go func() {
					transcript, summary, actionItems, err := transcription.TranscribeAudio(conversation.AudioFile.URL, gcpCreds.GladiaKey)
					if err == nil {
						_, _ = collection.UpdateOne(
							context.TODO(),
							bson.M{"_id": conversationID},
							bson.M{
								"$set": bson.M{
									"transcript":   transcript,
									"summary":      summary,
									"action_items": actionItems,
								},
							},
						)
					}
				}()
			}
		}
		_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": conversationID, "user_id": userID}, update)
		if err != nil {
			http.Error(w, "Error adding message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(message)
	}
}

func GlobalSearch(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Search query is required", http.StatusBadRequest)
			return
		}

		filter := bson.M{
			"user_id": userID,
			"$or": []bson.M{
				{"name": bson.M{"$regex": query, "$options": "i"}},
				{"transcript.sentence": bson.M{"$regex": query, "$options": "i"}},
			},
		}

		cursor, err := collection.Find(context.TODO(), filter)
		if err != nil {
			http.Error(w, "Error performing search", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())

		var results []models.Conversation
		for cursor.Next(context.TODO()) {
			var conversation models.Conversation
			if err := cursor.Decode(&conversation); err != nil {
				http.Error(w, "Error decoding search results", http.StatusInternalServerError)
				return
			}
			results = append(results, conversation)
		}

		json.NewEncoder(w).Encode(results)
	}
}

func UpdateTranscript(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		conversationID, _ := primitive.ObjectIDFromHex(params["id"])
		userID, err := auth.GetUserIDFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var transcriptUpdate struct {
			Transcript []string `json:"transcript"`
		}
		_ = json.NewDecoder(r.Body).Decode(&transcriptUpdate)

		update := bson.M{
			"$set": bson.M{
				"transcript": transcriptUpdate.Transcript,
				"updated_at": time.Now(),
			},
		}
		_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": conversationID, "user_id": userID}, update)
		if err != nil {
			http.Error(w, "Error updating transcript", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Transcript updated successfully"})
	}
}
