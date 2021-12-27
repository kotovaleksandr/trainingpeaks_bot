package main

import (
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
	var store usersstore = UserStore{}
	err := store.ClearStorage()
	if err != nil {
		t.Error(err)
	}
	user := User{
		Token:           "tttt",
		ChatId:          23234,
		TraininkPeaksId: 111,
	}

	err = store.CreateUser(user)
	if err != nil {
		t.Errorf("%s", err)
	}
	newUser, err := store.FindUserByChat(user.ChatId)
	if err != nil {
		t.Error(err)
	}

	if user.ChatId != newUser.ChatId {
		t.Error("Chatid not equal")
	}

	if user.Token != newUser.Token {
		t.Error("Token not equal")
	}

	if user.TraininkPeaksId != newUser.TraininkPeaksId {
		t.Error("TP id noq equal")
	}
}

func TestUpdateTokenForUser(t *testing.T) {
	var store usersstore = UserStore{}
	err := store.ClearStorage()
	if err != nil {
		t.Error(err)
	}
	user := User{
		Token:           "123123132",
		ChatId:          2323423,
		TraininkPeaksId: 111111,
	}

	err = store.CreateUser(user)
	if err != nil {
		t.Errorf("%s", err)
	}

	newToken := "12312312333333"
	user.Token = newToken
	err = store.UpdateUser(user)
	if err != nil {
		t.Errorf("%s", err)
	}

	newUser, err := store.FindUserByChat(user.ChatId)
	if err != nil {
		t.Errorf("%s", err)
	}

	if newUser.Token != newToken {
		t.Errorf("Token not updated")
	}
}

func TestUpdateUserWorkouts(t *testing.T) {
	var store usersstore = UserStore{}
	err := store.ClearStorage()
	if err != nil {
		t.Error(err)
	}
	user := User{
		Token:           "123123132",
		ChatId:          2323423,
		TraininkPeaksId: 111111,
	}

	err = store.CreateUser(user)
	if err != nil {
		t.Errorf("%s", err)
	}

	workouts := make([]Workout, 0)
	workouts = append(workouts, Workout{
		Title: "test1",
		WorkoutDay: CustomDate{
			Time: time.Now(),
		},
	})

	user.Workouts = workouts
	err = store.UpdateUser(user)
	if err != nil {
		t.Errorf("%s", err)
	}

	userWorkouts, err := store.FindUserByChat(user.ChatId)
	if err != nil {
		t.Errorf("%s", err)
	}

	if len(userWorkouts.Workouts) != 1 || userWorkouts.Workouts[0].Title != workouts[0].Title {
		t.Error()
	}
}

func TestNullFields(t *testing.T) {
	var store usersstore = UserStore{}
	err := store.ClearStorage()
	if err != nil {
		t.Error(err)
	}
	user := User{
		ChatId: 2323423,
		Status: "",
		Token:  "",
	}

	err = store.CreateUser(user)
	if err != nil {
		t.Errorf("%s", err)
	}

	user, err = store.FindUserByChat(user.ChatId)
	if err != nil {
		t.Errorf("%s", err)
	}

	users, err := store.GetAllUsers()
	if err != nil || len(users) == 0 {
		t.Errorf("%s", err)
	}
}
