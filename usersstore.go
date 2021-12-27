package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

type User struct {
	Token           string
	ChatId          int64
	TraininkPeaksId int
	Workouts        []Workout
	Status          string
}

type usersstore interface {
	CreateUser(user User) error
	FindUserByChat(chatId int64) (User, error)
	ClearStorage() error
	GetAllUsers() ([]User, error)
	UpdateUser(u User) error
}

var fileName = "users.database"

type UserStore struct {
}

func (store UserStore) CreateUser(user User) error {
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer db.Close()

	return db.Update(func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists([]byte("Users"))
		if err != nil {
			return err
		}
		userData, err := json.Marshal(user)
		if err != nil {
			return err
		}
		key := []byte(fmt.Sprintf("%d", user.ChatId))
		return b.Put(key, userData)
	})
}

func (store UserStore) FindUserByChat(chatId int64) (User, error) {
	user := User{}
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
		return user, err
	}
	defer db.Close()

	err = db.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte("Users"))
		if b == nil {
			return fmt.Errorf("bucket doesn't exists")
		}
		if err != nil {
			return err
		}

		key := []byte(fmt.Sprintf("%d", chatId))
		data := b.Get(key)
		err = json.Unmarshal(data, &user)
		return err
	})
	return user, err
}

func (store UserStore) GetAllUsers() ([]User, error) {
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	users := make([]User, 0)

	err = db.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte("Users"))
		if b == nil {
			return fmt.Errorf("bucket doesn't exists")
		}
		if err != nil {
			return err
		}

		err = b.ForEach(func(k, v []byte) error {
			user := User{}
			err = json.Unmarshal(v, &user)
			if err == nil {
				users = append(users, user)
			} else {
				log.Printf("Error on unmarshall user from db: %s. Key: %s", err, string(k))
			}
			return nil
		})
		return err
	})
	return users, err
}

func (store UserStore) ClearStorage() error {
	os.Remove(fileName)
	return nil
}

func (store UserStore) UpdateUser(u User) error {
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return db.Update(func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists([]byte("Users"))
		if err != nil {
			return err
		}
		key := []byte(fmt.Sprintf("%d", u.ChatId))

		userData, err := json.Marshal(u)
		if err != nil {
			return err
		}

		return b.Put(key, userData)
	})
}
