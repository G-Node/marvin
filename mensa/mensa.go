package mensa

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Address string
}

type Meal struct {
	Id       int
	Name     string
	Category string
	Notes    []string
	Prices   map[string]float32
}

func (mensa *Client) Meals(canteen string, day time.Time) ([]Meal, error) {
	client := &http.Client{}

	meals := make([]Meal, 0)

	url := fmt.Sprintf("%s/canteens/%s/days/%d-%02d-%02d/meals",
		mensa.Address, canteen,
		day.Year(), day.Month(), day.Day())
	log.Printf("calling Mensa with: %s", url)

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

func Emojify(notes string) string {
	repl := strings.NewReplacer(
		"Gericht mit Schweinefleisch", "🐖",
		"mit Fleisch", "🍖",
		"veganes Gericht", "🌿",
		"fleischloses Gericht", "🍄",
		"Gericht mit Rindfleisch", "🐂",
		"Gericht mit Alkohol", "🍷",
		"students", "♿",
		"employees", "👷",
		"others", "⛄",
		"mit Antioxidationsmittel", "🍋",
		"mit Konservierungsstoff", "🐢",
		"mit Süßungsmitteln", "🍯",
		"mit Phosphat", "☠",
		"mit einer Zuckerart und Süßungsmitteln", "🍯",
		"enthält eine Phenylalaninquelle", "⌬",
		"mit Farbstoff", "🖌")

	return repl.Replace(notes)
}
