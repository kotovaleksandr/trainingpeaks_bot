package main

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

//var token = "V001My--Hwk_3k2cTFhtpcZyafS1RyUegr0pfHP54fQPDwWVaiFniSSd1WhjlIjIjPXqUBSuFNLvPFtNV7NoLL9z0Pda3U-X2FFvTOEcPllHjos9HbkB8IpfAN5wiSSPBt5lFgKY7KaGB__6aJUjdgo3VedWsw9r2ISOZfcjZz6rDv6Ik3ArpWibzfUD5_y3F1rfM6wYFt59soPEBk-lqdFqM5b397ZSNdy9o2mcHDfW_asSaOhwPL1ktWt9V0ZMI43LmPiFypD2lK7ClYrdfxREKauoy29VsK5g6qZBIeLkh65WmcjjvatxLtByEq57v8w-G5784anKS4RMUkXFjnizng7Y6sYHWMi9CygBs5fgYscUclkhAO8OWKmTfjssgbruBwQdIdsOeQ9sKCgGtYRezTCLfD-BeDFBxE_3SFD0be1y5ZCWHfkG1P54cgYOyAwPwVn4BPWWZ-T52tktmo3soyAr_8c_DGVEEE5NLNtd5jUY5owYJNLw7TAcE7yw2W8U2TbSzoeugP6AUhXzLiy_hdFrbSgAoRybDVTPk7zUsTepzq4vsv7GyaJL4UC7HH5v03dfpOGHBcU0aggukdlLj3_8czCyxNwUfR0H6_-ikXVucxHJRtetfbmTvCPCZEY93GMKUQ_BBLBkiJIdcmSPU-umAy8t2hinhQ68tCtH99_4TUfuAqVZjOjoSPkLmsF8MZLOKHl9I4gpXGGwUs0zwuQG-_qIx2xE65nmThmoxnR7DqSt8JHI-mWHt3al7AKWuuNs8wcXJO9x3i3wRrRAKfqmzUTfSODAwuM7SYOr4frum_YyDrZuS1At5OnsKtKyudkZt_RzQ7HrN2wfnkjIXWVEUALOeT29Lnk5An1Bde7zcUk9kCGJR2mw7DDJizRajY5meb4XAhp0Ons2fw9hyamIfMv7jqWRN2079SBZ6nQEUsnw-aze7LDLB_BKA-7qqLhCsV7xtZMF4LGytHBleGoLCeJEmG3R8MG4F8LV50TK7U3N5xfaq8PSREjsPNiNOblRRxFbOpudPCvgBYhKg9GvVKc8njxU5gABaupLR0ugZ43FqfC3vIOeMsHrS56LPAbbeNiAsK44Jf09iBZKW9EX_tOD9W-k6GWgRCUSzerJevJv6GsI8lxoOFKmBpzIb5AjbQhB4Z4doWGBk1AGtw-gmV38F_eumawTQTKTmYhgy9nMd-1SkTINMCcUkE7P1nF-lsIH4-Td6YNFyrQJNoBrpqxWUxzN4l7h4FmLakV794COs-zx3dDEWVKBgwztGAmWOHf1t1b_w_Pk1Zv55qB1eClRjdRZzhqfnBIeguk9ZT-227QRuEAIfWkbre4yvcDuWwWOpCnnTxL7iR6ilAA2538wn0U9SuKrhfQKbWPzWXqVi7R-10GfjmeKj-HIafHHJGkPLaRaNZNAcEq6jq0Cv0elMyDIcb1J4KwXaei09Wt5oNX2riuiHn7abNmPGQD9EE1vDnaotrqRoUnezzcr7Aq_xf0Fi82cAefz1pI045QO1XQf_rDYlyI0lbLh23VWMeSk1itQ6QWNgX9B4jXb_quGj6CUTePuXfRIVM6sJn5yMkUfDIOk7I2R_kr1zCqHKREf8M-BFy1TQNp0boWktGMYeb6UD8E9twDcCvPOODu4DHp8JNXnDXvBbaZgGG-eShqwjAr4TY6a_rLGfbYvNqh73pu4zzzkBFoIvmtHt2dGQkPICnKWpyDFzP9x0"

func TestGetDataFromServer(t *testing.T) {
	token := getTokenFromFile(tp_token, "")
	data, err := client{}.GetDataFromServer(token)
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

func TestFindDiffsBetweenTwoSetsFindChangedText(t *testing.T) {
	content, err := ioutil.ReadFile("data_samples/workouts.json")
	if err != nil {
		t.Errorf("File read error: %s", err)
	}
	t.Logf("File first lines: %s", string(content)[0:100])
	workouts1, err := GetWorkoutsFromData(string(content))
	if err != nil {
		t.Errorf("File read error: %s", err)
	}

	content, err = ioutil.ReadFile("data_samples/workouts_1.json")
	if err != nil {
		t.Errorf("File read error: %s", err)
	}
	t.Logf("File first lines: %s", string(content)[0:100])
	workouts2, err := GetWorkoutsFromData(string(content))
	if err != nil {
		t.Errorf("File read error: %s", err)
	}
	t.Logf("Len first set: %d, len second set: %d", len(workouts1), len(workouts2))
	diff := CompareTwoSets(workouts1, workouts2)
	if len(diff) != 1 {
		t.Error("Diffs is empty")
		t.Fail()
	}
	if diff[0].WorkoutId != 1269137526 {
		t.Error("Found not valid changed workout")
	}
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
	work(&bot, client)
	if !bot.Status {
		t.Errorf("Send method not called")
	}
	if Workouts == nil {
		t.Error()
	}
}

func TestAddNewWorkouts(t *testing.T) {
	TestInitWorkouts(t)
	bot := sendMessageMock{}
	client := clientMock{}
	client.Data = make([]Workout, 0)
	client.Data = append(client.Data, Workout{
		Title: "test",
	})
	work(&bot, client)
	if !bot.Status {
		t.Errorf("Send method not called")
	}
	if !strings.Contains(bot.Message, "test") {
		t.Errorf("Message sended with incorrect text")
	}
	if Workouts == nil {
		t.Error()
	}
	if Workouts[0].Title != "test" {
		t.Error()
	}
}
