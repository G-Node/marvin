package mensa

import (
	"encoding/json"
	"fmt"
    "time"
	"io/ioutil"
	"net/http"
)

type Client struct {
    Address string
}


type Meal struct {
    Id int
    Name string
    Category string
    Notes []string
}

func (mensa *Client) Meals(canteen string, day time.Time) ([]Meal, error) {
    client := &http.Client{}

    meals := make([]Meal, 0)

	url := fmt.Sprintf("%s/canteens/%s/days/%d-%02d-%02d/meals",
                        mensa.Address, canteen,
                        day.Year(), day.Month(), day.Day())

	res, err := client.Get(url)
	if err != nil {
		return meals, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return meals, err
	}

	if err = json.Unmarshal(body, &meals); err != nil {
		return meals, err
	}

	return meals, nil
}

func (mensa *Client) MealsForToday(canteen string) ([]Meal, error) {
    return mensa.Meals(canteen, time.Now())
}

func (mensa *Client) MealsForTomorrow(canteen string) ([]Meal, error) {
    now := time.Now()
    tomorrow := now.AddDate(0, 0, 1)
    return mensa.Meals(canteen, tomorrow)
}

func Emojify(notes []string) []string {
    emojis := map[string]string {
        "Gericht mit Schweinefleisch": "🐖",
        "mit Fleisch": "🍖",
        "veganes Gericht": "🌿",
        "fleischloses Gericht": "🍄",
    }   
    
    for i, note := range notes {
        emoji, ok := emojis[note]
        if ok {
            notes[i] = emoji
        }
    } 
    
    return notes
}

func main() {
    client := &Client{Address: "http://openmensa.org/api/v2"}
    res, err := client.Meals("134", time.Now())
    if err != nil {
        println(err)
    }
    
    fmt.Printf("%v\n", res)
}
