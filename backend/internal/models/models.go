package models

import (
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var JWTSecret = []byte(os.Getenv("JWT_SECRET"))

type TranscriptionWord struct {
	Word       string  `json:"word" bson:"word"`
	Start      float64 `json:"start" bson:"start"`
	End        float64 `json:"end" bson:"end"`
	Confidence float64 `json:"confidence" bson:"confidence"`
}

type TranscriptionSentence struct {
	Sentence   string              `json:"sentence" bson:"sentence"`
	Start      float64             `json:"start" bson:"start"`
	End        float64             `json:"end" bson:"end"`
	Words      []TranscriptionWord `json:"words" bson:"words"`
	Confidence float64             `json:"confidence" bson:"confidence"`
	Speaker    string              `json:"speaker" bson:"speaker"`
	Channel    int                 `json:"channel" bson:"channel"`
}

type Conversation struct {
	ID          primitive.ObjectID      `json:"id,omitempty" bson:"_id,omitempty"`
	UserID      primitive.ObjectID      `json:"user_id" bson:"user_id"`
	Name        string                  `json:"name" bson:"name"`
	AudioFile   *AudioFile              `json:"audio_file" bson:"audio_file"`
	Transcript  []TranscriptionSentence `json:"transcript" bson:"transcript"`
	ChatHistory []ChatMessage           `json:"chat_history" bson:"chat_history"`
	Summary     string                  `json:"summary" bson:"summary"`
	ActionItems []string                `json:"action_items" bson:"action_items"`
	CreatedAt   time.Time               `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at" bson:"updated_at"`
}

type ChatMessage struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Content   string             `json:"content" bson:"content"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
}

type User struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Email    string             `json:"email" bson:"email"`
	Password string             `json:"password,omitempty" bson:"password"`
}
type GCPCredentials struct {
	UserID      primitive.ObjectID `bson:"user_id"`
	Credentials string             `bson:"credentials"`
	BucketName  string             `bson:"bucket_name"`
	GladiaKey   string             `bson:"gladia_key"`
}

type AudioFile struct {
	Name string `json:"name" bson:"name"`
	URL  string `json:"url" bson:"url"`
}
