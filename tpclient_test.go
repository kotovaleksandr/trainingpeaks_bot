package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// getTokenFromFile is a test helper for integration tests that read credentials from files.
func getTokenFromFile(fileName string) string {
	f, err := os.Open(fileName)
	if err != nil {
		log.Printf("Test credential file %s not found: %s", fileName, err)
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	return scanner.Text()
}

const (
	tp_token string = "tp_token"
)

func TestGetDataFromServer(t *testing.T) {
	token := getTokenFromFile(tp_token)
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
	workouts1 = append(workouts1, Workout{Title: "test1", WorkoutId: 1})
	workouts2 := make([]Workout, 0)
	workouts2 = append(workouts2, Workout{Title: "test1", WorkoutId: 1})
	workouts2 = append(workouts2, Workout{Title: "test2", WorkoutId: 2})
	workouts2 = append(workouts2, Workout{Title: "test3", WorkoutId: 3})
	diff := CompareTwoSets(workouts1, workouts2)
	if len(diff) != 2 {
		t.Errorf("Expected 2 diff, got %d", len(diff))
	}
	if diff[0].Title != "test2" {
		t.Error("Diff not contains new workout test2")
	}
	if diff[1].Title != "test3" {
		t.Error("Diff not contains new workout test3")
	}
}

func TestFindChangedWorkout(t *testing.T) {
	workouts1 := make([]Workout, 0)
	workouts1 = append(workouts1, Workout{
		Title:       "test1",
		WorkoutId:   1,
		Description: "old",
	})
	workouts2 := make([]Workout, 0)
	workouts2 = append(workouts2, Workout{
		Title:       "test1",
		WorkoutId:   1,
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
	workouts1 = append(workouts1, Workout{Title: "test1", WorkoutId: 1})
	workouts1 = append(workouts1, Workout{Title: "test2", WorkoutId: 2})

	workouts2 := make([]Workout, 0)
	workouts2 = append(workouts2, Workout{Title: "test1", WorkoutId: 1})

	diff := CompareTwoSets(workouts1, workouts2)
	if len(diff) != 0 {
		t.Errorf("Deleted workout is not diff")
	}
}

func TestRefreshToken(t *testing.T) {
	token := getTokenFromFile(tp_token)
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
	c := clientMock{Data: []Workout{{Title: "test", WorkoutId: 1}}}
	state := &appState{token: "token"}

	work(&bot, c, state, 1111, 333)

	if !bot.Status {
		t.Errorf("Send method not called")
	}
	workouts := state.getWorkouts()
	if workouts == nil || len(workouts) != 1 {
		t.Error("Workouts not stored in state")
	}
}

func TestAddNewWorkouts(t *testing.T) {
	bot := sendMessageMock{}
	c := clientMock{Data: []Workout{}}
	state := &appState{token: "token"}

	// First call — init
	work(&bot, c, state, 1111, 333)

	// Second call — new workout appears
	c.Data = []Workout{{Title: "test", WorkoutId: 1}}
	work(&bot, c, state, 1111, 333)

	if !bot.Status {
		t.Errorf("Send method not called")
	}
	if !strings.Contains(bot.Message, "test") {
		t.Errorf("Message sent with incorrect text: %s", bot.Message)
	}
	workouts := state.getWorkouts()
	if workouts == nil || workouts[0].Title != "test" {
		t.Error("Workouts not updated in state")
	}
}
