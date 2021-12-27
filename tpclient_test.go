package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	tp_token string = "tp_token"
)

func TestGetDataFromServer(t *testing.T) {
	token := getTokenFromFile(tp_token, "")
	if os.Getenv("USER_ID") == "" {
		return
	}
	userId, _ := strconv.Atoi(os.Getenv("USER_ID"))
	data, err := client{}.GetDataFromServer(token, userId)
	if err != nil {
		t.Errorf("File read error: %s", err)
	}

	if len(data) == 0 {
		t.Error("Workouts is empty")
	}
}

func TestGetWorkoutsFromData(t *testing.T) {
	content, err := ioutil.ReadFile("data_samples/workouts.json")
	if err != nil {
		t.Errorf("File read error: %s", err)
	}
	t.Logf("File first lines: %s", string(content)[0:100])
	GetWorkoutsFromData(string(content))
}

func TestFindNewWorkout(t *testing.T) {
	workouts1 := make([]Workout, 0)
	workouts1 = append(workouts1, Workout{
		Title: "test1",
	})
	workouts2 := make([]Workout, 0)
	workouts2 = append(workouts2, Workout{
		Title: "test1",
	})
	workouts2 = append(workouts2, Workout{
		Title: "test2",
	})
	workouts2 = append(workouts2, Workout{
		Title: "test3",
	})
	diff := CompareTwoSets(workouts1, workouts2)
	if len(diff) != 2 {
		t.Error("Diff not found")
	}
	if diff[0].Title != "test2" {
		t.Error("Diff not contains new workout")
	}
	if diff[1].Title != "test3" {
		t.Error("Diff not contains new workout")
	}
}

func TestFindChangedWorkout(t *testing.T) {
	workouts1 := make([]Workout, 0)
	workouts1 = append(workouts1, Workout{
		Title:       "test1",
		Description: "old",
	})
	workouts2 := make([]Workout, 0)
	workouts2 = append(workouts2, Workout{
		Title:       "test1",
		Description: "new",
	})

	diff := CompareTwoSets(workouts1, workouts2)
	if len(diff) != 1 {
		t.Error("Diff not found")
	}
	if diff[0].Title != "test1" {
		t.Error("Diff not contains updated workout")
	}
	if diff[0].Description != "new" {
		t.Error("Diff not contains updated workout")
	}
}

func TestWorkoutDeleted(t *testing.T) {
	workouts1 := make([]Workout, 0)
	workouts1 = append(workouts1, Workout{
		Title: "test1",
	})
	workouts1 = append(workouts1, Workout{
		Title: "test2",
	})

	workouts2 := make([]Workout, 0)
	workouts2 = append(workouts2, Workout{
		Title: "test1",
	})

	diff := CompareTwoSets(workouts1, workouts2)
	if len(diff) != 0 {
		t.Errorf("Deleted workout is not diff")
	}
}

func TestRefreshToken(t *testing.T) {
	token := getTokenFromFile(tp_token, "")
	updatedToken, _, err := RefreshToken(token)
	if err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		if updatedToken == "" {
			t.Error("Empty token")
			t.FailNow()
		} else {
			t.Logf("Recieved token: %s", updatedToken)
		}
	}
}

func TestDate(t *testing.T) {
	date := time.Now()
	date1 := time.Now()
	date1 = date1.AddDate(0, 0, -7)
	t.Log(date.String())
	t.Logf(date.Format("2006-01-02"))
	t.Logf(date1.Format("2006-01-02"))
}

func TestInitWorkouts(t *testing.T) {
	bot := sendMessageMock{}
	client := clientMock{}
	client.Data = make([]Workout, 0)
	client.Data = append(client.Data, Workout{Title: "test"})
	store := UserStoreMock{
		Users: make([]User, 0),
	}
	store.Users = append(store.Users, User{Token: "token", ChatId: 333, TraininkPeaksId: 1111})
	work(&bot, client, store)
	if !bot.Status {
		t.Errorf("Send method not called")
	}
	workouts, _ := store.GetUserWorkouts(store.Users[0].TraininkPeaksId)
	if workouts == nil || len(workouts) != 1 {
		t.Error()
	}
}

func TestAddNewWorkouts(t *testing.T) {
	bot := sendMessageMock{}
	client := clientMock{}
	client.Data = make([]Workout, 0)

	store := UserStoreMock{
		Users: make([]User, 0),
	}
	store.Users = append(store.Users, User{Token: "token", ChatId: 333, TraininkPeaksId: 1111})
	work(&bot, client, store)
	client = clientMock{}
	client.Data = make([]Workout, 0)
	client.Data = append(client.Data, Workout{
		Title: "test",
	})

	work(&bot, client, store)
	if !bot.Status {
		t.Errorf("Send method not called")
	}
	if !strings.Contains(bot.Message, "test") {
		t.Errorf("Message sended with incorrect text")
	}
	workouts, _ := store.GetUserWorkouts(store.Users[0].TraininkPeaksId)
	if workouts == nil {
		t.Error()
	}
	if workouts[0].Title != "test" {
		t.Error()
	}
}
