package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Workout struct {
	Title            string
	WorkoutDay       CustomDate
	Description      string
	WorkoutId        int64
	TotalTime        float64
	TotalTimePlanned float64
}

type CustomDate struct {
	time.Time
}

type client struct {
}

func (c *CustomDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"") + ".000Z"
	c.Time, _ = time.Parse(time.RFC3339, s)
	return nil
}

func GetWorkoutsFromData(data string) ([]Workout, error) {
	var f []Workout
	err := json.Unmarshal([]byte(data), &f)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return f, err
}

// GetCurrentUserID fetches the authenticated athlete's numeric ID from TrainingPeaks.
func GetCurrentUserID(token string) (int, error) {
	type userResponse struct {
		User struct {
			UserID int `json:"userId"`
		} `json:"user"`
	}

	httpClient := resty.New()
	resp, err := httpClient.R().
		SetCookies(getCookies(token)).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0").
		Get("https://tpapi.trainingpeaks.com/users/v3/user")

	if err != nil {
		return 0, fmt.Errorf("user profile request failed: %w", err)
	}
	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("user profile endpoint returned %d: %s", resp.StatusCode(), resp.String())
	}

	var u userResponse
	if err := json.Unmarshal(resp.Body(), &u); err != nil {
		return 0, fmt.Errorf("failed to parse user profile: %w", err)
	}
	if u.User.UserID == 0 {
		return 0, fmt.Errorf("userId is 0 in response: %s", resp.String())
	}
	return u.User.UserID, nil
}

func moscowLocation() *time.Location {
	return time.FixedZone("Moscow", 3*60*60)
}

func (w Workout) IsDone() bool {
	return w.TotalTime > 0
}

func (c client) getWorkoutsForRange(token string, userId int, from, to time.Time) ([]Workout, error) {
	httpClient := resty.New()
	url := fmt.Sprintf(
		"https://tpapi.trainingpeaks.com/fitness/v6/athletes/%d/workouts/%s/%s",
		userId,
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)
	log.Printf("Get workouts from url: %s", url)

	resp, err := httpClient.R().
		EnableTrace().
		SetCookies(getCookies(token)).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0").
		Get(url)

	if err != nil {
		log.Printf("Error on get workouts: %s", err)
		return nil, err
	}
	if resp.StatusCode() != 200 {
		log.Printf("Server return non success status code: %d: %s", resp.StatusCode(), resp.String())
		return nil, fmt.Errorf("server error: %s", resp.Status())
	}
	return GetWorkoutsFromData(resp.String())
}

func (c client) GetDataFromServer(token string, userId int) ([]Workout, error) {
	return c.getWorkoutsForRange(token, userId, time.Now(), time.Now().AddDate(0, 0, 7))
}

func (c client) GetTodayWorkouts(token string, userId int) ([]Workout, error) {
	data, err := c.GetDataFromServer(token, userId)
	if err != nil {
		return nil, err
	}
	moscow := moscowLocation()
	today := time.Now().In(moscow)
	workouts := make([]Workout, 0)
	for _, w := range data {
		d := w.WorkoutDay.Time.In(moscow)
		if d.Year() == today.Year() && d.Month() == today.Month() && d.Day() == today.Day() {
			workouts = append(workouts, w)
		}
	}
	return workouts, nil
}

func (c client) GetTomorrowWorkouts(token string, userId int) ([]Workout, error) {
	data, err := c.GetDataFromServer(token, userId)
	if err != nil {
		return nil, err
	}
	moscow := moscowLocation()
	tomorrow := time.Now().In(moscow).AddDate(0, 0, 1)
	workouts := make([]Workout, 0)
	for _, w := range data {
		d := w.WorkoutDay.Time.In(moscow)
		if d.Year() == tomorrow.Year() && d.Month() == tomorrow.Month() && d.Day() == tomorrow.Day() {
			workouts = append(workouts, w)
		}
	}
	return workouts, nil
}

func (c client) GetRemainOnWeekWorkouts(token string, userId int) ([]Workout, error) {
	moscow := moscowLocation()
	from := time.Now().In(moscow)
	// End of week (Sunday 23:59:59)
	daysSinceMonday := int(from.Weekday()) - int(time.Monday)
	if daysSinceMonday < 0 {
		daysSinceMonday += 7
	}
	monday := from.AddDate(0, 0, -daysSinceMonday)
	endOfWeek := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, moscow).AddDate(0, 0, 7)

	data, err := c.GetDataFromServer(token, userId)
	if err != nil {
		return nil, err
	}
	workouts := make([]Workout, 0)
	for _, w := range data {
		d := w.WorkoutDay.Time.In(moscow)
		if !d.Before(from) && d.Before(endOfWeek) {
			workouts = append(workouts, w)
		}
	}
	return workouts, nil
}

func (c client) GetWeekWorkouts(token string, userId int) ([]Workout, error) {
	moscow := moscowLocation()
	now := time.Now().In(moscow)
	daysSinceMonday := int(now.Weekday()) - int(time.Monday)
	if daysSinceMonday < 0 {
		daysSinceMonday += 7
	}
	monday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, moscow).AddDate(0, 0, -daysSinceMonday)
	sunday := monday.AddDate(0, 0, 7) // exclusive: up to (but not including) next Monday

	workouts, err := c.getWorkoutsForRange(token, userId, monday, sunday)
	if err != nil {
		return nil, err
	}
	filtered := make([]Workout, 0)
	for _, w := range workouts {
		d := w.WorkoutDay.Time.In(moscow)
		if !d.Before(monday) && d.Before(sunday) {
			filtered = append(filtered, w)
		}
	}
	return filtered, nil
}

// CompareTwoSets returns workouts that are new or changed (by WorkoutId).
func CompareTwoSets(old []Workout, updated []Workout) []Workout {
	changed := make([]Workout, 0)
	oldMap := make(map[int64]Workout, len(old))
	for _, w := range old {
		oldMap[w.WorkoutId] = w
	}
	for _, w := range updated {
		prev, exists := oldMap[w.WorkoutId]
		if !exists || prev.Description != w.Description || prev.TotalTime != w.TotalTime {
			changed = append(changed, w)
		}
	}
	return changed
}

func RefreshToken(token string) (string, string, error) {
	httpClient := resty.New()
	resp, err := httpClient.R().
		EnableTrace().
		SetCookies(getCookies(token)).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0").
		Get("https://home.trainingpeaks.com/refresh")

	if err != nil {
		log.Printf("Error on token refresh: %s", err)
		return "", "", err
	}
	if resp.StatusCode() != 200 {
		log.Printf("Server return non success status code: %d", resp.StatusCode())
		return "", "", fmt.Errorf("server error: %s", resp.Status())
	}

	var newToken, userId string
	for _, cookie := range resp.Cookies() {
		log.Printf("Cookie: %s", cookie.Name)
		if cookie.Name == "Production_tpAuth" {
			newToken = cookie.Value
		}
		if cookie.Name == "ajs_user_id" {
			userId = cookie.Value
		}
	}
	log.Printf("Cookies read. Token: %s, userId %s", newToken, userId)

	if newToken == "" {
		newToken = token
	}
	return newToken, userId, nil
}

func getCookies(token string) []*http.Cookie {
	return []*http.Cookie{
		{
			Name:   "TPtosAgreed_Production",
			Value:  "true",
			Path:   "/",
			Domain: ".trainingpeaks.com",
		},
		{
			Name:   "Production_tpAuth",
			Value:  token,
			Path:   "/",
			Domain: ".trainingpeaks.com",
		},
	}
}
