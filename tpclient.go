package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jinzhu/now"
)

type Workout struct {
	Title       string
	WorkoutDay  CustomDate
	Description string
	WorkoutId   int64
	TotalTime   float64
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

func (c client) GetDataFromServer(token string, userId int) ([]Workout, error) {
	client := resty.New()

	req := client.R()
	url := fmt.Sprintf("https://tpapi.trainingpeaks.com/fitness/v6/athletes/%d/workouts/%s/%s", userId, time.Now().Format("2006-01-02"), time.Now().AddDate(0, 0, 7).Format("2006-01-02"))

	log.Printf("Get workouts from url: %s", url)

	resp, err := req.
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

func (c client) GetTodayWorkouts(token string, userId int) ([]Workout, error) {
	data, err := c.GetDataFromServer(token, userId)
	if err != nil {
		return nil, err
	}
	workouts := make([]Workout, 0)
	for _, w := range data {
		if w.WorkoutDay.Year() == time.Now().Year() && w.WorkoutDay.Month() == time.Now().Month() && w.WorkoutDay.Day() == time.Now().Day() {
			workouts = append(workouts, w)
		}
	}
	return workouts, nil
}
func (c client) GetRemainOnWeekWorkouts(token string, userId int) ([]Workout, error) {
	from := time.Now()
	to := now.EndOfWeek()
	data, err := c.GetDataFromServer(token, userId)
	if err != nil {
		return nil, err
	}
	workouts := make([]Workout, 0)
	for _, w := range data {
		if w.WorkoutDay.After(from) && w.WorkoutDay.Before(to) {
			workouts = append(workouts, w)
		}
	}
	return workouts, nil
}

func CompareTwoSets(one []Workout, two []Workout) []Workout {
	changedWorkouts := make([]Workout, 0)
	if len(one) < len(two) {
		for i := len(one); i < len(two); i++ {
			changedWorkouts = append(changedWorkouts, two[i])
		}
	} else if len(one) == len(two) {
		for i := range one {
			if one[i].Description != two[i].Description || one[i].TotalTime != two[i].TotalTime {
				changedWorkouts = append(changedWorkouts, two[i])
			}
		}
	}
	return changedWorkouts
}

func RefreshToken(token string) (string, string, error) {
	client := resty.New()
	req := client.R()
	resp, err := req.
		EnableTrace().
		SetCookies(getCookies(token)).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0").
		Get("https://home.trainingpeaks.com/refresh")

	if err != nil {
		log.Printf("Error on get workouts: %s", err)
		return "", "", err
	}

	if resp.StatusCode() != 200 {
		log.Printf("Server return non success status code: %d", resp.StatusCode())
		return "", "", err
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

	log.Printf("Cookies readed. Token: %s, userId %s", newToken, userId)

	if newToken == "" {
		newToken = token
	}

	return newToken, userId, err
}

func getCookies(token string) []*http.Cookie {
	return []*http.Cookie{{
		Name:   "TPtosAgreed_Production",
		Value:  "true",
		Path:   "/",
		Domain: ".trainingpeaks.com",
	}, {
		Name:   "Production_tpAuth",
		Value:  token,
		Path:   "/",
		Domain: ".trainingpeaks.com",
	}}
}
