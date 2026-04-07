package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages       []Message              `json:"messages"`
	SystemPrompt   string                 `json:"systemPrompt"`
	DocumentInputs map[string]string      `json:"documentInputs"`
	ChatContext    string                 `json:"chatContext"`
}

type GenerateDocRequest struct {
	StepName         string            `json:"stepName"`
	ChatHistory      []Message         `json:"chatHistory"`
	DocumentInputs   map[string]string `json:"documentInputs"`
	ChatContext      string            `json:"chatContext"`
	GenerationPrompt string            `json:"generationPrompt"`
}

type GenerateOptionsRequest struct {
	QuestionText string    `json:"questionText"`
	ChatHistory  []Message `json:"chatHistory"`
}

type GenerateOptionsResponse struct {
	Reasoning        string   `json:"reasoning"`
	Options          []string `json:"options"`
	RecommendedIndex *int     `json:"recommendedIndex"`
	Confidence       string   `json:"confidence"`
	Question         string   `json:"-"` // We'll populate this manually from the request
}

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

func (c *Client) GenerateOptions(req GenerateOptionsRequest) (*GenerateOptionsResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Logging
	logFile, _ := os.OpenFile("tui-api-debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		fmt.Fprintf(logFile, "\n--- GENERATE OPTIONS REQUEST ---\nQuestion: %q\n", req.QuestionText)
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/generate-options", c.BaseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		if logFile != nil {
			fmt.Fprintf(logFile, "REQUEST ERROR: %v\n", err)
			logFile.Close()
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		if logFile != nil {
			fmt.Fprintf(logFile, "API ERROR (%d): %s\n", resp.StatusCode, string(b))
			logFile.Close()
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(b))
	}

	var res GenerateOptionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		if logFile != nil {
			fmt.Fprintf(logFile, "DECODE ERROR: %v\n", err)
			logFile.Close()
		}
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	res.Question = req.QuestionText

	if logFile != nil {
		fmt.Fprintf(logFile, "SUCCESS: %d options found\n", len(res.Options))
		logFile.Close()
	}

	return &res, nil
}

func (c *Client) StreamChat(req ChatRequest, onChunk func(string)) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/chat", c.BaseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to chat: %s", string(b))
	}

	// For debugging, log chunks to a file
	logFile, _ := os.OpenFile("tui-api-debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		fmt.Fprintf(logFile, "\n--- NEW STREAM [%s] ---\n", time.Now().Format(time.Kitchen))
	}

	start := time.Now()
	firstChunk := true
	buf := make([]byte, 4096) // Larger buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if firstChunk {
				if logFile != nil {
					fmt.Fprintf(logFile, "FIRST CHUNK after %v\n", time.Since(start))
				}
				firstChunk = false
			}
			chunk := string(buf[:n])
			onChunk(chunk)
		}
		if err == io.EOF {
			if logFile != nil {
				fmt.Fprintf(logFile, "--- STREAM EOF (Total: %v) ---\n", time.Since(start))
			}
			break
		}
		if err != nil {
			if logFile != nil {
				fmt.Fprintf(logFile, "--- STREAM ERROR after %v: %v ---\n", time.Since(start), err)
			}
			return err
		}
	}

	if logFile != nil {
		logFile.Close()
	}
	return nil
}

func (c *Client) StreamGenerateDoc(req GenerateDocRequest, onChunk func(string)) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/generate-doc", c.BaseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to generate doc: %s", string(b))
	}

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			onChunk(string(buf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}
