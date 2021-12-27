package main

type UserStoreMock struct {
	Users []User
}

func (u UserStoreMock) UpdateToken(trainingpeaksId int, token string) error {
	return nil
}

func (u UserStoreMock) CreateUser(user User) error {
	return nil
}

func (u UserStoreMock) FindUserByChat(chatId int64) (User, error) {
	return u.Users[0], nil
}

func (u UserStoreMock) UpdateUserWorkouts(trainingpeaksId int, workouts []Workout) error {
	u.Users[0].Workouts = workouts
	return nil
}

func (u UserStoreMock) GetUserWorkouts(trainingpeaksId int) ([]Workout, error) {
	return u.Users[0].Workouts, nil
}

func (u UserStoreMock) ClearStorage() error {
	return nil
}

func (u UserStoreMock) GetAllUsers() ([]User, error) {
	return u.Users, nil
}

func (u UserStoreMock) UpdateUser(user User) error {
	u.Users[0] = user
	return nil
}
