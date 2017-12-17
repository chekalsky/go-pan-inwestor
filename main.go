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
			if contains(actualCoins, d) {
				log.Println("New coins found:", d, "Price:", allCoins[d].PriceUsd)

				sendMessage(fmt.Sprintf("Нашёл новые монеты: %s (%s). Стоимость: $%.6f", d, allCoins[d].Symbol, allCoins[d].PriceUsd))
			}
		}
	} else {
		log.Println("No new coins")
	}

	log.Println("Sleeping for 5 minutes")
}

// ---------------------------------------------------------------

func sendMessage(s string) {
	msg := tgbotapi.NewMessage(Config.ChatId, s)

	bot.Send(msg)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}
