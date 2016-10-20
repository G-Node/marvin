package quotes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type quote struct {
	Txt    string
	Author string
}

func GetRandomQuote() (quote, error) {
	client := &http.Client{}
	req, err := client.Get("http://localhost:8090/getquote")
	if err != nil {
		return quote{"Nothing", "Nobody"}, err
	}

	if req.StatusCode == http.StatusOK {
		decoder := json.NewDecoder(req.Body)
		randQuote := quote{}
		decoder.Decode(&randQuote)
		req.Body.Close()
		return randQuote, nil
	}
	return quote{"Nothing", "Nobody2"}, fmt.Errorf("status:%s, Code:%d", req.Status, req.StatusCode)
}

func LearnQuote(text string) error {
	if strings.ContainsAny(text, "~") {
		substr := strings.Split(text, "~")
		newQuote := quote{strings.Trim(substr[0], " "), strings.Trim(substr[1], " ")}
		newQuoteJSON, err := json.Marshal(newQuote)
		if err != nil {
			fmt.Print(err)
			return err
		}
		client := &http.Client{}
		req, err := http.NewRequest("POST", "http://localhost:8090/learnquote",
			bytes.NewBuffer(newQuoteJSON))
		req.Header.Set("Content-Type", "application/json")
		client.Do(req)
		return err
	}
	return errors.New("No Author provided")
}
