package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var Workouts []Workout
var dataFileName = "users.dat"

const (
	tp_token string = "tp_token"
)

type sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type tpclient interface {
	GetDataFromServer(token string) ([]Workout, error)
}

func main() {
	telegramToken := getTokenFromFile("telegram_token", "Telegram token (telegram_token)")
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	userId, err := getUserIdAndRefreshToken()
	if err != nil {
		log.Fatalf("User id not found")
	} else {
		log.Printf("User id %s", userId)
	}
	client := client{
		UserId: os.Getenv("USER_ID"),
	}

	work(bot, client)

	go func() {
		for range time.Tick(time.Minute * 10) {
			work(bot, client)
		}
	}()

	go func() {
		for range time.Tick(time.Minute * 9) {
			_, err := getUserIdAndRefreshToken()
			if err != nil {
				log.Fatalf("User id not found")
			}
		}
	}()
	go waitNewUsers(bot)

	select {}
}

func work(bot sender, client tpclient) {
	tpToken := getTokenFromFile(tp_token, "Training peaks token (tp_token)")
	log.Printf("Try get data from server")
	actualWorkouts, err := client.GetDataFromServer(tpToken)
	if err != nil {
		log.Printf("Error on recieving data from server: %s", err)
		return
	}

	log.Printf("Found %d workouts", len(actualWorkouts))
	if Workouts == nil {
		log.Print("Init workouts")
		Workouts = actualWorkouts
		for _, user := range getUsers() {
			msg := tgbotapi.NewMessage(user, "Ok, initial workouts loaded")
			log.Printf("Send message to user %v", user)
			bot.Send(msg)
		}
	} else {
		diff := CompareTwoSets(Workouts, actualWorkouts)

		if len(diff) > 0 {
			for _, currentWorkout := range diff {
				for _, user := range getUsers() {
					msg := tgbotapi.NewMessage(user, fmt.Sprintf("Updated workout: %v at %v, description: %v\n", currentWorkout.Title, currentWorkout.WorkoutDay.Format("2006-01-02"), currentWorkout.Description))
					log.Printf("Send message to user %v", user)
					bot.Send(msg)
					Workouts = actualWorkouts
				}
			}
		} else {
			log.Print("No new data")
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

func getUsers() []int64 {
	log.Printf("Try read data from file %v", dataFileName)
	file, err := os.Open(dataFileName)

	result := make([]int64, 0)
	if err != nil && !os.IsExist(err) {
		log.Printf("Dat file not found, create them")
		os.Create(dataFileName)
		return result
	}

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Error read users files: %v\n", err)
			break
		}

		in, err := strconv.ParseInt(strings.TrimSpace(line), 10, 64)
		log.Printf("Get line from file: %d. Source line:%v\n", in, line)
		if err != nil {
			log.Printf("Data file corrupted: %v\n", err)
		} else {
			result = append(result, in)
		}
	}

	return result
}

func waitNewUsers(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		log.Printf("Wait messages error: %s", err)
		return
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		addUser(update.Message.Chat.ID)
		bot.Send(msg)
	}
}

func addUser(user int64) {
	log.Printf("Add user %d to file %v", user, dataFileName)
	file, err := os.Open(dataFileName)
	if !os.IsExist(err) {
		log.Printf("Data file not found, create them")
		file, err = os.Create(dataFileName)
		if err != nil {
			panic(err)
		}
	}
	_, err = file.WriteString(fmt.Sprintf("%v\n", user))
	if err != nil {
		panic(err)
	}
	file.Close()
}

func getUserIdAndRefreshToken() (string, error) {
	tpToken := getTokenFromFile("tp_token", "Training peaks token (tp_token)")
	token, userId, err := RefreshToken(tpToken)
	if err != nil {
		log.Panicf("Error on token refresh %s", err)
		return userId, err
	}
	if token != "" {
		writeTokenToFile(token, tp_token)
	}
	return userId, nil
}

func writeTokenToFile(token string, filename string) {
	log.Printf("Write token to file %s", filename)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil && !os.IsExist(err) {
		log.Printf("Data file not found, create them")
		file, err = os.Create(filename)
		if err != nil {
			panic(err)
		}
	}
	file.Truncate(0)
	_, err = file.WriteString(fmt.Sprintf("%v\n", token))
	if err != nil {
		panic(err)
	}
	file.Close()
}
