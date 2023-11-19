package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	yandexToken   = ""
	telegramToken = ""
	yandexURL     = "https://300.ya.ru/api/sharing-url"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	updates, err := bot.GetUpdatesChan(tgbotapi.UpdateConfig{Timeout: 60})
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		bot.Send(tgbotapi.NewMessage(chatID, "Обрабатываю"))

		message, err := sender(update.Message.Text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Error: %v", err)))
		}
		bot.Send(tgbotapi.NewMessage(chatID, message))
	}
}

func sender(msg string) (string, error) {
	requestBody, err := json.Marshal(map[string]string{"article_url": msg})
	if err != nil {
		return "", err
	}

	resp, err := sendRequest(requestBody)
	if err != nil {
		return "", err
	}

	result, err := parseResponse(resp)
	if err != nil {
		return "", err
	}

	return result, nil
}

func sendRequest(requestBody []byte) (string, error) {
	req, err := http.NewRequest("POST", yandexURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "OAuth "+yandexToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return result["sharing_url"], nil
}

func parseResponse(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	fmt.Println("Ссылка: ", url)
	title := doc.Find("title").Text()
	//fmt.Println(title)
	result := title

	doc.Find("span.text-wrapper").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		result += "\n" + text
	})

	return result, nil
}
