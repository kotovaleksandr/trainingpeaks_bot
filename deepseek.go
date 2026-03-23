package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type aiAdvisor interface {
	GetNutritionAdvice(done []Workout, tomorrow []Workout) (string, error)
}

type deepseekClient struct {
	apiKey  string
	baseURL string
	model   string
	prompt  string
}

type dsMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type dsRequest struct {
	Model    string      `json:"model"`
	Messages []dsMessage `json:"messages"`
}

type dsResponse struct {
	Choices []struct {
		Message dsMessage `json:"message"`
	} `json:"choices"`
}

func (d deepseekClient) GetNutritionAdvice(done []Workout, tomorrow []Workout) (string, error) {
	if d.apiKey == "" {
		return "", fmt.Errorf("DeepSeek API key not configured")
	}
	if len(tomorrow) == 0 {
		return "", fmt.Errorf("no workouts provided")
	}

	var doneText string
	for _, w := range done {
		doneText += fmt.Sprintf("- %s", w.Title)
		if w.TotalTime > 0 {
			doneText += fmt.Sprintf(" (%.0f min", w.TotalTime*60)
			if w.Calories > 0 {
				doneText += fmt.Sprintf(", %.0f kcal", w.Calories)
			}
			doneText += ")"
		}
		doneText += "\n"
	}

	var tomorrowText string
	for _, w := range tomorrow {
		tomorrowText += fmt.Sprintf("- %s", w.Title)
		if w.Description != "" {
			tomorrowText += fmt.Sprintf(": %s", w.Description)
		}
		if w.TotalTimePlanned > 0 {
			tomorrowText += fmt.Sprintf(" (%.0f min)", w.TotalTimePlanned*60)
		}
		tomorrowText += "\n"
	}

	var userMessage string
	if doneText != "" {
		userMessage = fmt.Sprintf("%s\n\nToday's completed workouts:\n%s\nTomorrow's workouts:\n%s", d.prompt, doneText, tomorrowText)
	} else {
		userMessage = fmt.Sprintf("%s\n\nTomorrow's workouts:\n%s", d.prompt, tomorrowText)
	}

	reqBody := dsRequest{
		Model:    d.model,
		Messages: []dsMessage{{Role: "user", Content: userMessage}},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	log.Printf("DeepSeek: requesting nutrition advice (done: %d, tomorrow: %d)", len(done), len(tomorrow))

	req, err := http.NewRequest("POST", d.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("DeepSeek: request failed: %s", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("DeepSeek: response status %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek API error: %s", resp.Status)
	}

	var dsResp dsResponse
	if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil {
		return "", err
	}
	if len(dsResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from DeepSeek")
	}
	log.Printf("DeepSeek: received advice (%d chars)", len(dsResp.Choices[0].Message.Content))
	return dsResp.Choices[0].Message.Content, nil
}
