package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"

	"./quotes"
	"github.com/G-Node/marvin/mensa"
)

type Bot struct {
	conn *websocket.Conn
	id   string
	quit chan bool
}

type Message struct {
	Id       uint64 `json:"id"`
	Type     string `json:"type"`
	Channel  string `json:"channel"`
	Text     string `json:"text"`
	Presence string `json:"presence"`
	User     string `json:"user"`
}

type responseSelf struct {
	Id string `json:"id"`
}

type responseRtmStart struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	Url   string       `json:"url"`
	Self  responseSelf `json:"self"`
}

type responseUserProfileGet struct {
	Ok      bool             `json:"ok"`
	Error   string           `json:"error"`
	Profile slackUserProfile `json:"profile"`
}

type slackUserProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

var counter uint64

func (b *Bot) getMessage() (m Message, err error) {
	err = websocket.JSON.Receive(b.conn, &m)
	return
}

func (b *Bot) postMessage(m Message) error {
	m.Id = atomic.AddUint64(&counter, 1)
	err := websocket.JSON.Send(b.conn, m)
	if err != nil {
		log.Printf("Could nor send message:%s becuase:%s", m.Text, err)
	}
	return err
}

func (b *Bot) postText(text string, channelID string) error {
	log.Printf("Try to post:%s to channel:%s", text, channelID)
	m := Message{}
	m.Text = text
	m.Type = "message"
	m.Id = atomic.AddUint64(&counter, 1)
	m.Channel = channelID
	log.Println(m)
	err := websocket.JSON.Send(b.conn, m)
	if err != nil {
		log.Printf("Could not send message:%s becuase:%s", m.Text, err)
	}
	return err
}

func (b *Bot) mensa(channelID string, text string) {

	mc := &mensa.Client{Address: "http://openmensa.org/api/v2"}

	var meals []mensa.Meal
	if strings.Contains(text, "tomorrow") {
		meals, _ = mc.MealsForTomorrow("134")
		// ignored error for now
	} else {
		meals, _ = mc.MealsForToday("134")
	}
	if len(meals) == 0 {
		b.postText(channelID, "No milk today, my love has gone away...")
		return
	}
	if strings.Contains(text, "beilagen") {
		b.postMeals(channelID, meals, []string{"Beilagen"})
		return
	}
	b.postMeals(channelID, meals, []string{"Tagesgericht", "Aktionsessen", "Biogericht", "Aktion"})
	return
}

func (b *Bot) postMeals(channelID string, meals []mensa.Meal, catToUse []string) {
	messageFilter := strings.NewReplacer("[]", "")
	for _, meal := range meals {
		category := meal.Category
		if !stringInSlice(category, catToUse) {
			continue
		}
		var prices []string
		for key, value := range meal.Prices {
			if value != 0. {
				prices = append(prices, fmt.Sprintf("%s:%.2f€",
					key, value))
			}
		}
		notes := mensa.Emojify(strings.Join(meal.Notes, ", "))
		message := fmt.Sprintf("%s [%s] [%s]", meal.Name, notes,
			mensa.Emojify(strings.Join(prices, ", ")))
		b.postText(messageFilter.Replace(message), channelID)
	}
}

func (b *Bot) quote(channelId string) {
	quote, err := quotes.GetRandomQuote()
	if err != nil {
		log.Println(err)
	}
	b.postText(fmt.Sprintf("Did you know that %s said:  \"%s\"", quote.Author, quote.Txt), channelId)
}

func (b *Bot) getUserProfile(userID string) slackUserProfile {
	profile := slackUserProfile{}
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s&user=%s",
		os.Getenv("MARVIN_TOKEN"), userID)
	resp, err := http.Get(url)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return profile
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {

	}
	var respObj responseUserProfileGet
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return profile
	}

	if !respObj.Ok {
		err = fmt.Errorf("Slack error: %s", respObj.Error)
		return profile
	}
	return respObj.Profile
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(a, b) {
			return true
		}
	}
	return false
}

func slackStart(token string) (wsurl, id string, err error) {
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", token)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	var respObj responseRtmStart
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return
	}

	if !respObj.Ok {
		err = fmt.Errorf("Slack error: %s", respObj.Error)
		return
	}

	wsurl = respObj.Url
	id = respObj.Self.Id
	return
}

func slackConnect(token string) (*websocket.Conn, string) {
	wsurl, id, err := slackStart(token)
	if err != nil {
		log.Fatal(err)
	}

	ws, err := websocket.Dial(wsurl, "", "https://api.slack.com/")
	if err != nil {
		log.Fatal(err)
	}

	return ws, id
}

func NewBot() *Bot {
	c, id := slackConnect(os.Getenv("MARVIN_TOKEN"))
	b := &Bot{conn: c, id: id}
	return b
}
func main() {
	bot := NewBot()
	log.Printf("Got a new bot id %s", bot.id)
	rand.Seed(time.Now().UTC().UnixNano())

	for {
		// read each incoming message
		m, err := bot.getMessage()
		log.Println(m.Channel)
		log.Println(m.Text)
		if m.Type == "message" {
			if strings.Contains(m.Text, "nix") {
				m.Type = "message"
				m.Text = "https://www.youtube.com/watch?v=Go4SI5ie7qE"
				bot.postMessage(m)
			}
			if strings.Contains(m.Text, "mensa") {
				bot.mensa(m.Channel, m.Text)
			}
			if strings.Contains(m.Text, "marvin") {
				bot.postText("I am depressed", m.Channel)
			}

			if err != nil {
				log.Fatal(err)
			}
		}
		if m.Type == "channel_join" || strings.Contains(m.Text, "quote") {
			bot.quote(m.Channel)
		}

		if m.Type == "presence_change" {
			log.Println(m)
			if m.Presence == "active" {
				bot.quote("C087SE58E")
			}
		}
	}
	//b.SetupGithub()

	//<-b.quit
}
