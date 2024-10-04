package transcription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/TheLickIn13Keys/omi-webapp/internal/models"
)

const gladiaV2BaseURL = "https://api.gladia.io/v2/"

type TranscriptionRequest struct {
	AudioURL            string `json:"audio_url"`
	DiarizationEnhanced bool   `json:"diarization_enhanced"`
	Sentences           bool   `json:"sentences"`
}

type TranscriptionResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	ResultURL string `json:"result_url"`
}

type TranscriptionResult struct {
	Status string `json:"status"`
	Result struct {
		Transcription struct {
			FullTranscript string                         `json:"full_transcript"`
			Sentences      []models.TranscriptionSentence `json:"sentences"`
		} `json:"transcription"`
	} `json:"result"`
}

func TranscribeAudio(audioURL, gladiaKey string) ([]models.TranscriptionSentence, error) {
	requestData := TranscriptionRequest{
		AudioURL:            audioURL,
		DiarizationEnhanced: true,
		Sentences:           true,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request data: %v", err)
	}

	req, err := http.NewRequest("POST", gladiaV2BaseURL+"transcription/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	print("sending request")

	req.Header.Set("x-gladia-key", gladiaKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var transcriptionResp TranscriptionResponse
	err = json.Unmarshal(body, &transcriptionResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	if transcriptionResp.ResultURL == "" {
		return nil, fmt.Errorf("no result URL in response")
	}

	return pollForResult(transcriptionResp.ResultURL, gladiaKey)
}

func pollForResult(resultURL, gladiaKey string) ([]models.TranscriptionSentence, error) {
	client := &http.Client{}

	for {
		req, err := http.NewRequest("GET", resultURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating poll request: %v", err)
		}

		req.Header.Set("x-gladia-key", gladiaKey)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending poll request: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading poll response body: %v", err)
		}

		var pollResult TranscriptionResult
		err = json.Unmarshal(body, &pollResult)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling poll response: %v", err)
		}

		if pollResult.Status == "done" {
			return pollResult.Result.Transcription.Sentences, nil
		}

		print("polling...")

		time.Sleep(5 * time.Second)
	}
}
