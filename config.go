package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	TelegramToken   string `json:"telegram_token"`
	TelegramChatID  int64  `json:"telegram_chat_id"`
	TPToken         string `json:"tp_token"`
	TPUserID        int    `json:"tp_user_id"`
	DeepSeekAPIKey  string `json:"deepseek_api_key"`
	DeepSeekPrompt  string `json:"deepseek_prompt"`
	DeepSeekModel   string `json:"deepseek_model"`
	DeepSeekBaseURL string `json:"deepseek_base_url"`
}

func saveConfigField(path string, key string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	raw[key] = value
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

func SaveChatID(path string, chatID int64) error {
	return saveConfigField(path, "telegram_chat_id", chatID)
}

func SaveTPToken(path string, token string) error {
	return saveConfigField(path, "tp_token", token)
}

func LoadConfig(path string) Config {
	cfg := Config{
		DeepSeekModel:   "deepseek-chat",
		DeepSeekBaseURL: "https://api.deepseek.com",
		DeepSeekPrompt:  "You are a nutritionist. Based on the workouts planned for tomorrow, give a brief recommendation on calorie intake and key nutrition tips. Be concise.",
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		log.Fatalf("Config file %s not found", path)
	}
	if err != nil {
		log.Fatalf("Error opening config: %s", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatalf("Error parsing config: %s", err)
	}
	return cfg
}
