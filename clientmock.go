package main

type clientMock struct {
	Data []Workout
}

func (c clientMock) GetDataFromServer(token string, tpId int) ([]Workout, error) {
	return c.Data, nil
}

func (c clientMock) GetTodayWorkouts(token string, userId int) ([]Workout, error) {
	data, err := c.GetDataFromServer(token, userId)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (c clientMock) GetRemainOnWeekWorkouts(token string, userId int) ([]Workout, error) {
	data, err := c.GetDataFromServer(token, userId)
	if err != nil {
		return nil, err
	}
	return data, nil
}
