package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type tpclient interface {
	GetDataFromServer(token string, userId int) ([]Workout, error)
	GetTodayWorkouts(token string, userId int) ([]Workout, error)
	GetRemainOnWeekWorkouts(token string, userId int) ([]Workout, error)
}

func main() {
	telegramToken := getTokenFromFile("telegram_token", "Telegram token (telegram_token)")
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	client := client{}
	store := UserStore{}
	work(bot, client, store)

	go func() {
		for range time.Tick(time.Minute * 10) {
			work(bot, client, store)
		}
	}()

	go func() {
		for range time.Tick(time.Minute * 9) {
			err := getUserIdAndRefreshToken(store)
			if err != nil {
				log.Fatalf("User id not found")
			}
		}
	}()
	go waitNewUsers(bot, store, client)

	select {}
}

func work(bot sender, client tpclient, store usersstore) {

	users, err := store.GetAllUsers()
	if err != nil {
		log.Printf("Error on read users from store %s", err)
	}
	for _, u := range users {
		if u.TraininkPeaksId == 0 {
			log.Printf("Skip user %d with TP id = 0", u.ChatId)
			continue
		}
		log.Printf("Try get data from server. UserID: %d", u.TraininkPeaksId)
		actualWorkouts, err := client.GetDataFromServer(u.Token, u.TraininkPeaksId)
		if err != nil {
			log.Printf("Error on recieving data from server: %s", err)
			continue
		}

		log.Printf("Found %d workouts", len(actualWorkouts))
		if u.Workouts == nil {
			log.Print("Init workouts")
			u.Workouts = actualWorkouts
			store.UpdateUser(u)

			msg := tgbotapi.NewMessage(int64(u.ChatId), "Ok, initial workouts loaded")
			log.Printf("Send message to user %v", u.ChatId)
			bot.Send(msg)

		} else {
			log.Printf("User workouts %d, recived workouts %d", len(u.Workouts), len(actualWorkouts))
			diff := CompareTwoSets(u.Workouts, actualWorkouts)

			if len(diff) > 0 {
				for _, currentWorkout := range diff {
					msg := tgbotapi.NewMessage(int64(u.ChatId), fmt.Sprintf("Updated workout: %v at %v, description: %v\n", currentWorkout.Title, currentWorkout.WorkoutDay.Format("2006-01-02"), currentWorkout.Description))
					log.Printf("Send message to user %v", u.ChatId)
					bot.Send(msg)
					u.Workouts = actualWorkouts
					store.UpdateUser(u)
				}
			} else {
				log.Print("No new data")
			}
		}
	}

}

func getTokenFromFile(fileName string, tokenKind string) string {
	tokenFile, err := os.Open(fileName)
	if os.IsNotExist(err) {
		log.Fatalf("%s token file not found", tokenKind)
		panic(err)
	}

	scanner := bufio.NewScanner(tokenFile)
	scanner.Scan()
	token := scanner.Text()
	if scanner.Err() != nil {
		panic(scanner.Err())
	}

	return token
}

func waitNewUsers(bot *tgbotapi.BotAPI, store usersstore, c client) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		user, err := store.FindUserByChat(update.Message.Chat.ID)
		if err != nil || user.ChatId == 0 {
			log.Printf("Error on get user: %d, %s", update.Message.Chat.ID, err)
			log.Printf("User not found, try create them")
			user = User{
				ChatId: update.Message.Chat.ID,
			}
			store.CreateUser(user)
		} else {
			log.Printf("Found user %d", user.ChatId)
		}

		var messageText string
		if update.Message.IsCommand() {
			log.Printf("Recived command %s", update.Message.Text)
			if update.Message.Text == "/token" {
				user.Status = "token"
				messageText = "Enter TP token value"
			} else if update.Message.Text == "/id" {
				user.Status = "id"
				messageText = "Enter TP id"
			} else if update.Message.Text == "/today" {
				w, err := c.GetTodayWorkouts(user.Token, user.TraininkPeaksId)
				if err != nil {
					log.Printf("Failed to get today workouts: %s", err)
					continue
				} else {
					for _, currentWorkout := range w {
						msg := tgbotapi.NewMessage(int64(user.ChatId), fmt.Sprintf("Workout: %v at %v, description: %v\n", currentWorkout.Title, currentWorkout.WorkoutDay.Format("2006-01-02"), currentWorkout.Description))
						log.Printf("Send message to user %v", user.ChatId)
						bot.Send(msg)
					}
					continue
				}
			} else if update.Message.Text == "/week" {
				w, err := c.GetRemainOnWeekWorkouts(user.Token, user.TraininkPeaksId)
				if err != nil {
					log.Printf("Failed to get week workouts: %s", err)
					continue
				} else {
					for _, currentWorkout := range w {
						msg := tgbotapi.NewMessage(int64(user.ChatId), fmt.Sprintf("Workout: %v at %v, description: %v\n", currentWorkout.Title, currentWorkout.WorkoutDay.Format("2006-01-02"), currentWorkout.Description))
						log.Printf("Send message to user %v", user.ChatId)
						bot.Send(msg)
					}
					continue
				}
			}
		} else {
			if user.Status == "token" {
				user.Token = update.Message.Text
				messageText = "Token saved!"
				user.Status = ""
			} else if user.Status == "id" {
				user.TraininkPeaksId, err = strconv.Atoi(update.Message.Text)
				if err != nil {
					messageText = "Id has incorrect format, please enter again"
				} else {
					messageText = "Id saved!"
					user.Status = ""
				}
			}
		}

		err = store.UpdateUser(user)
		if err != nil {
			log.Printf("Error on user status save %s", err)
		}
		message := tgbotapi.NewMessage(user.ChatId, messageText)
		_, err = bot.Send(message)
		if err != nil {
			log.Printf("Error on message send: %s", err)
		}
	}
}

func getUserIdAndRefreshToken(store usersstore) error {
	users, err := store.GetAllUsers()
	if err != nil {
		log.Printf("Error on read users, try lateer")
		return err
	}
	for _, u := range users {
		token, _, err := RefreshToken(u.Token)
		if err != nil {
			log.Printf("Error on token refresh %s", err)
			return err
		}
		if token != "" {
			u.Token = token
			store.UpdateUser(u)
		}
	}

	return nil
}
