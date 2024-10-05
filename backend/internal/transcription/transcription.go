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
	Summarization       bool   `json:"summarization"`
	AudioToLLM          bool   `json:"audio_to_llm"`
	AudioToLLMConfig    struct {
		Prompts []string `json:"prompts"`
	} `json:"audio_to_llm_config"`
}

type TranscriptionResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	ResultURL string `json:"result_url"`
}

type TranscriptionResult struct {
	Status string `json:"status"`
	Result struct {
		Metadata struct {
			AudioDuration            float64 `json:"audio_duration"`
			NumberOfDistinctChannels int     `json:"number_of_distinct_channels"`
			BillingTime              float64 `json:"billing_time"`
			TranscriptionTime        float64 `json:"transcription_time"`
		} `json:"metadata"`
		Transcription struct {
			Utterances     []models.TranscriptionSentence `json:"utterances"`
			FullTranscript string                         `json:"full_transcript"`
			Languages      []string                       `json:"languages"`
			Sentences      []models.TranscriptionSentence `json:"sentences"`
		} `json:"transcription"`
		AudioToLLM struct {
			Success  bool          `json:"success"`
			IsEmpty  bool          `json:"is_empty"`
			Results  []LLMResponse `json:"results"`
			ExecTime float64       `json:"exec_time"`
			Error    interface{}   `json:"error"`
		} `json:"audio_to_llm"`
	} `json:"result"`
}

type LLMResponse struct {
	Success  bool        `json:"success"`
	IsEmpty  bool        `json:"is_empty"`
	Results  LLMResults  `json:"results"`
	ExecTime float64     `json:"exec_time"`
	Error    interface{} `json:"error"`
}

type LLMResults struct {
	Prompt   string `json:"prompt"`
	Response string `json:"response"`
}

func TranscribeAudio(audioURL, gladiaKey string) ([]models.TranscriptionSentence, string, []string, error) {
	requestData := TranscriptionRequest{
		AudioURL:            audioURL,
		DiarizationEnhanced: true,
		Sentences:           true,
		Summarization:       true,
		AudioToLLM:          true,
		AudioToLLMConfig: struct {
			Prompts []string `json:"prompts"`
		}{
			Prompts: []string{
				"Extract the key action items the transcription as bullet points",
				"Generate a title from this transcription",
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error marshaling request data: %v", err)
	}

	req, err := http.NewRequest("POST", gladiaV2BaseURL+"transcription/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("x-gladia-key", gladiaKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error reading response body: %v", err)
	}

	var transcriptionResp TranscriptionResponse
	err = json.Unmarshal(body, &transcriptionResp)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	if transcriptionResp.ResultURL == "" {
		return nil, "", nil, fmt.Errorf("no result URL in response")
	}

	return pollForResult(transcriptionResp.ResultURL, gladiaKey)
}

func pollForResult(resultURL, gladiaKey string) ([]models.TranscriptionSentence, string, []string, error) {
	client := &http.Client{}

	for {
		req, err := http.NewRequest("GET", resultURL, nil)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating poll request: %v", err)
		}

		req.Header.Set("x-gladia-key", gladiaKey)

		resp, err := client.Do(req)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error sending poll request: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error reading poll response body: %v", err)
		}

		var pollResult TranscriptionResult
		err = json.Unmarshal(body, &pollResult)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error unmarshaling poll response: %v", err)
		}

		if pollResult.Status == "done" {
			sentences := pollResult.Result.Transcription.Sentences
			if len(sentences) == 0 {
				sentences = pollResult.Result.Transcription.Utterances
			}

			summary := pollResult.Result.Transcription.FullTranscript
			actionItems := make([]string, 0)
			for _, llmResponse := range pollResult.Result.AudioToLLM.Results {
				if llmResponse.Results.Prompt == "Extract the key action items the transcription as bullet points" {
					actionItems = append(actionItems, llmResponse.Results.Response)
				}
			}

			return sentences, summary, actionItems, nil
		}

		time.Sleep(5 * time.Second)
	}
}
