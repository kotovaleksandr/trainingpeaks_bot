package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type tpclient interface {
	GetDataFromServer(token string, userId int) ([]Workout, error)
	GetTodayWorkouts(token string, userId int) ([]Workout, error)
	GetTomorrowWorkouts(token string, userId int) ([]Workout, error)
	GetRemainOnWeekWorkouts(token string, userId int) ([]Workout, error)
	GetWeekWorkouts(token string, userId int) ([]Workout, error)
}

// appState holds mutable runtime state for the single user.
type appState struct {
	mu       sync.RWMutex
	token    string
	workouts []Workout
}

func (s *appState) getToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.token
}

func (s *appState) setToken(t string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = t
}

func (s *appState) getWorkouts() []Workout {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workouts
}

func (s *appState) setWorkouts(w []Workout) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workouts = w
}

func main() {
	cfg := LoadConfig("config.json")

	if cfg.TelegramToken == "" {
		log.Fatal("telegram_token is not set in config.json")
	}
	if cfg.TPToken == "" {
		log.Fatal("tp_token is not set in config.json")
	}

	// Resolve TrainingPeaks user ID: use config value or auto-detect from profile API
	tpUserID := cfg.TPUserID
	initialToken := cfg.TPToken
	if tpUserID == 0 {
		log.Print("tp_user_id not set, fetching from TrainingPeaks...")
		var err error
		tpUserID, err = GetCurrentUserID(cfg.TPToken)
		if err != nil {
			log.Fatalf("Could not auto-detect tp_user_id: %s", err)
		}
		log.Printf("Auto-detected tp_user_id: %d", tpUserID)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	if cfg.TelegramChatID == 0 {
		log.Printf("telegram_chat_id is not set — send any message to the bot to discover your chat ID")
	}

	ai := deepseekClient{
		apiKey:  cfg.DeepSeekAPIKey,
		baseURL: cfg.DeepSeekBaseURL,
		model:   cfg.DeepSeekModel,
		prompt:  cfg.DeepSeekPrompt,
	}

	tpClient := client{}
	state := &appState{token: initialToken}

	// Initial workout load
	work(bot, tpClient, state, tpUserID, cfg.TelegramChatID)

	// Polling every 10 minutes
	go func() {
		for range time.Tick(time.Minute * 10) {
			work(bot, tpClient, state, tpUserID, cfg.TelegramChatID)
		}
	}()

	// Token refresh every 9 minutes
	go func() {
		for range time.Tick(time.Minute * 9) {
			refreshToken(state)
		}
	}()

	// Daily stats at 19:00 Moscow time
	go runDailyAt(19, 0, moscowLocation(), func() {
		if cfg.TelegramChatID != 0 {
			sendDailyStats(bot, tpClient, ai, state.getToken(), tpUserID, cfg.TelegramChatID)
		}
	})

	// Weekly plan every Monday at 19:00 Moscow time
	go runWeeklyAt(time.Monday, 19, 0, moscowLocation(), func() {
		if cfg.TelegramChatID != 0 {
			sendWeeklyPlan(bot, tpClient, state.getToken(), tpUserID, cfg.TelegramChatID)
		}
	})

	// Telegram command handler
	go handleCommands(bot, tpClient, ai, state, tpUserID, &cfg)

	select {}
}

// runDailyAt fires action every day at the given hour:minute in loc timezone.
func runDailyAt(hour, minute int, loc *time.Location, action func()) {
	for {
		now := time.Now().In(loc)
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		log.Printf("Next daily schedule at %v", next)
		time.Sleep(next.Sub(time.Now()))
		action()
	}
}

// runWeeklyAt fires action every week on the given weekday at hour:minute in loc timezone.
func runWeeklyAt(weekday time.Weekday, hour, minute int, loc *time.Location, action func()) {
	for {
		now := time.Now().In(loc)
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		daysUntil := int(weekday) - int(now.Weekday())
		if daysUntil < 0 {
			daysUntil += 7
		} else if daysUntil == 0 && !next.After(now) {
			daysUntil = 7
		}
		next = next.AddDate(0, 0, daysUntil)
		log.Printf("Next weekly schedule at %v", next)
		time.Sleep(next.Sub(time.Now()))
		action()
	}
}

func work(bot sender, c tpclient, state *appState, userID int, chatID int64) {
	if userID == 0 {
		return
	}
	token := state.getToken()
	log.Printf("Polling workouts for userID: %d", userID)

	actualWorkouts, err := c.GetDataFromServer(token, userID)
	if err != nil {
		log.Printf("Error receiving data from server: %s", err)
		return
	}
	log.Printf("Found %d workouts", len(actualWorkouts))

	prev := state.getWorkouts()
	if prev == nil {
		log.Print("Init workouts")
		if chatID != 0 {
			bot.Send(tgbotapi.NewMessage(chatID, "Bot connected, workouts loaded."))
		}
	} else {
		diff := CompareTwoSets(prev, actualWorkouts)
		for _, w := range diff {
			if chatID != 0 {
				bot.Send(tgbotapi.NewMessage(chatID, formatWorkoutUpdate(w)))
			}
		}
		if len(diff) == 0 {
			log.Print("No new data")
		}
	}
	state.setWorkouts(actualWorkouts)
}

func refreshToken(state *appState) {
	newToken, _, err := RefreshToken(state.getToken())
	if err != nil {
		log.Printf("Token refresh error: %s", err)
		return
	}
	if newToken != "" && newToken != state.getToken() {
		state.setToken(newToken)
		log.Print("Token refreshed")
	}
}

func handleCommands(bot *tgbotapi.BotAPI, c tpclient, ai aiAdvisor, state *appState, tpUserID int, cfg *Config) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// Auto-save chat ID on first message
		if cfg.TelegramChatID == 0 {
			if err := SaveChatID("config.json", chatID); err != nil {
				log.Printf("Failed to save chat ID to config: %s", err)
				bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(
					"Your chat ID is: %d\nCould not save automatically — add it to config.json as \"telegram_chat_id\".",
					chatID,
				)))
			} else {
				cfg.TelegramChatID = chatID
				log.Printf("Chat ID %d saved to config.json", chatID)
				bot.Send(tgbotapi.NewMessage(chatID, "Chat ID saved. Bot is ready — try /today, /digest or /plan."))
			}
			continue
		}

		// Ignore messages from other chats
		if chatID != cfg.TelegramChatID {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		token := state.getToken()
		switch update.Message.Command() {
		case "today":
			w, err := c.GetTodayWorkouts(token, tpUserID)
			if err != nil {
				log.Printf("Failed to get today workouts: %s", err)
				bot.Send(tgbotapi.NewMessage(chatID, "Failed to fetch workouts."))
			} else {
				sendWorkoutList(bot, chatID, w, "Today:")
			}
		case "week":
			w, err := c.GetRemainOnWeekWorkouts(token, tpUserID)
			if err != nil {
				log.Printf("Failed to get week workouts: %s", err)
				bot.Send(tgbotapi.NewMessage(chatID, "Failed to fetch workouts."))
			} else {
				sendWorkoutList(bot, chatID, w, "Remaining this week:")
			}
		case "digest":
			sendDailyStats(bot, c, ai, token, tpUserID, chatID)
		case "plan":
			sendWeeklyPlan(bot, c, token, tpUserID, chatID)
		}
	}
}

func sendDailyStats(bot sender, c tpclient, ai aiAdvisor, token string, userID int, chatID int64) {
	moscow := moscowLocation()
	today := time.Now().In(moscow)

	todayWorkouts, err := c.GetTodayWorkouts(token, userID)
	if err != nil {
		log.Printf("Error getting today workouts: %s", err)
		bot.Send(tgbotapi.NewMessage(chatID, "Failed to fetch today's workouts."))
		return
	}
	tomorrowWorkouts, err := c.GetTomorrowWorkouts(token, userID)
	if err != nil {
		log.Printf("Error getting tomorrow workouts: %s", err)
		bot.Send(tgbotapi.NewMessage(chatID, "Failed to fetch tomorrow's workouts."))
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📅 *%s*\n", today.Format("2 January 2006")))

	done := filterDone(todayWorkouts, true)
	planned := filterDone(todayWorkouts, false)

	if len(done) > 0 {
		sb.WriteString("\n✅ *Done today:*\n")
		for _, w := range done {
			sb.WriteString(formatWorkoutShort(w))
		}
	}
	if len(planned) > 0 {
		sb.WriteString("\n⏳ *Planned today:*\n")
		for _, w := range planned {
			sb.WriteString(formatWorkoutShort(w))
		}
	}
	if len(todayWorkouts) == 0 {
		sb.WriteString("\nNo workouts today.\n")
	}

	tomorrow := today.AddDate(0, 0, 1)
	if len(tomorrowWorkouts) > 0 {
		sb.WriteString(fmt.Sprintf("\n🗓 *Tomorrow, %s:*\n", tomorrow.Format("January 2")))
		for _, w := range tomorrowWorkouts {
			sb.WriteString(formatWorkoutShort(w))
		}
		advice, err := ai.GetNutritionAdvice(tomorrowWorkouts)
		if err != nil {
			log.Printf("DeepSeek error: %s", err)
		} else {
			sb.WriteString("\n🍽 *Nutrition advice for tomorrow:*\n")
			sb.WriteString(advice)
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString(fmt.Sprintf("\n🗓 *Tomorrow, %s:* no workouts.\n", tomorrow.Format("January 2")))
	}

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending daily stats: %s", err)
	}
}

func sendWeeklyPlan(bot sender, c tpclient, token string, userID int, chatID int64) {
	moscow := moscowLocation()
	now := time.Now().In(moscow)

	daysSinceMonday := int(now.Weekday()) - int(time.Monday)
	if daysSinceMonday < 0 {
		daysSinceMonday += 7
	}
	monday := now.AddDate(0, 0, -daysSinceMonday)
	sunday := monday.AddDate(0, 0, 6)

	workouts, err := c.GetWeekWorkouts(token, userID)
	if err != nil {
		log.Printf("Error getting week workouts: %s", err)
		bot.Send(tgbotapi.NewMessage(chatID, "Failed to fetch weekly workouts."))
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📅 *Week plan %s – %s:*\n\n",
		monday.Format("2 Jan"),
		sunday.Format("2 Jan"),
	))

	if len(workouts) == 0 {
		sb.WriteString("No workouts scheduled for this week.")
	} else {
		for _, w := range workouts {
			d := w.WorkoutDay.Time.In(moscow)
			sb.WriteString(fmt.Sprintf("*%s %s* — %s\n",
				weekdayShort(d.Weekday()),
				d.Format("01/02"),
				formatWorkoutLine(w),
			))
		}
	}

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending weekly plan: %s", err)
	}
}

func filterDone(workouts []Workout, done bool) []Workout {
	result := make([]Workout, 0)
	for _, w := range workouts {
		if w.IsDone() == done {
			result = append(result, w)
		}
	}
	return result
}

func formatWorkoutShort(w Workout) string {
	s := fmt.Sprintf("• *%s*", w.Title)
	if w.Description != "" {
		s += fmt.Sprintf(" — %s", w.Description)
	}
	return s + "\n"
}

func formatWorkoutLine(w Workout) string {
	if w.Description != "" {
		return fmt.Sprintf("%s — %s", w.Title, w.Description)
	}
	return w.Title
}

func formatWorkoutUpdate(w Workout) string {
	return fmt.Sprintf("🔔 Workout updated: *%s* (%s)\n%s",
		w.Title,
		w.WorkoutDay.Format("02.01.2006"),
		w.Description,
	)
}

func weekdayShort(d time.Weekday) string {
	names := map[time.Weekday]string{
		time.Monday:    "Mon",
		time.Tuesday:   "Tue",
		time.Wednesday: "Wed",
		time.Thursday:  "Thu",
		time.Friday:    "Fri",
		time.Saturday:  "Sat",
		time.Sunday:    "Sun",
	}
	return names[d]
}

func sendWorkoutList(bot sender, chatID int64, workouts []Workout, header string) {
	if len(workouts) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, header+" no workouts."))
		return
	}
	var sb strings.Builder
	sb.WriteString(header + "\n")
	for _, w := range workouts {
		sb.WriteString(formatWorkoutShort(w))
	}
	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
