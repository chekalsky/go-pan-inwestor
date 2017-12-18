package main

import (
	"fmt"
	"github.com/chekalskiy/go-coinmarketcap"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/configor"
	"github.com/nanobox-io/golang-scribble"
	"log"
	"sort"
	"strings"
	"time"
)

var cachedList []string
var Config = struct {
	ChatId   int64  `required:"true"`
	BotToken string `required:"true"`
}{}

var db *scribble.Driver
var bot *tgbotapi.BotAPI

func init() {
	var err error

	err = configor.Load(&Config, "config.yml")
	if err != nil {
		log.Fatalln("Config Error", err)
	}

	log.Println("Parameters loaded")

	db, err = scribble.New("./db", nil)
	if err != nil {
		log.Fatalln("DB Error", err)
	}

	db.Read("coins", "all", &cachedList)

	log.Println(len(cachedList), "cached coins loaded from DB")

	bot, err = tgbotapi.NewBotAPI(Config.BotToken)
	if err != nil {
		log.Fatalln("Telegram Error", err)
	}

	log.Println("Telegram Bot initialized")
}

func main() {
	for {
		go work()

		time.Sleep(time.Minute * 5)
	}
}

func work() {
	allCoins, err := coinmarketcap.GetAllCoinData(0)
	if err != nil {
		log.Println(err)

		return
	}

	var actualCoins []string

	for _, coin := range allCoins {
		actualCoins = append(actualCoins, coin.ID)
	}

	sort.Strings(actualCoins)

	err = db.Write("coins", "all", actualCoins)
	if err != nil {
		log.Println(err)
	}

	ah := strings.Join(actualCoins, "")
	ch := strings.Join(cachedList, "")

	isNew := false
	if len(cachedList) == 0 {
		isNew = true
	}

	if ah != ch && !isNew {
		diff := difference(cachedList, actualCoins)
		cachedList = actualCoins

		log.Println("Actual:", ah)
		log.Println("Cached:", ch)
		log.Println("Difference:", len(diff), diff)

		for _, d := range diff {
			log.Println("New coins found:", d, "Price:", allCoins[d].PriceUsd)

			sendMessage(fmt.Sprintf("Нашёл новые монеты: %s (%s). Стоимость: $%.4f\n\nhttps://coinmarketcap.com/currencies/%s/", d, allCoins[d].Symbol, allCoins[d].PriceUsd, d))
		}
	} else {
		log.Println("No new coins")
	}

	log.Println("Sleeping for 5 minutes")
}

// ---------------------------------------------------------------

func sendMessage(s string) {
	msg := tgbotapi.NewMessage(Config.ChatId, s)
	msg.DisableWebPagePreview = true

	bot.Send(msg)
}

/**
Find new elements in slice2 comparing to slice1
*/
func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	for _, s2 := range slice2 {
		found := false
		for _, s1 := range slice1 {
			if s1 == s2 {
				found = true
				break
			}
		}

		if !found {
			diff = append(diff, s2)
		}
	}

	return diff
}
