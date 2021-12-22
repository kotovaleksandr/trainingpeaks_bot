package main

type clientMock struct {
	Data []Workout
}

func (c clientMock) GetDataFromServer(token string) ([]Workout, error) {
	return c.Data, nil
}
