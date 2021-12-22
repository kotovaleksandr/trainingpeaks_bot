package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type sendMessageMock struct {
	Status  bool
	Message string
}

func (s *sendMessageMock) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	s.Status = true

	log.Printf("Recieved message %s", c)
	s.Message = fmt.Sprintf("%s", c)
	return tgbotapi.Message{
		Text: "work",
	}, nil
}
