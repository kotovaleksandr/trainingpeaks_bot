package main

type clientMock struct {
	Data []Workout
}

func (c clientMock) GetDataFromServer(token string, tpId int) ([]Workout, error) {
	return c.Data, nil
}

func (c clientMock) GetTodayWorkouts(token string, userId int) ([]Workout, error) {
	return c.Data, nil
}

func (c clientMock) GetTomorrowWorkouts(token string, userId int) ([]Workout, error) {
	return c.Data, nil
}

func (c clientMock) GetRemainOnWeekWorkouts(token string, userId int) ([]Workout, error) {
	return c.Data, nil
}

func (c clientMock) GetWeekWorkouts(token string, userId int) ([]Workout, error) {
	return c.Data, nil
}
