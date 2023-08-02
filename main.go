package main

import (
	"log"
	"math/rand"
	"os"
	"time"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	hider = ".\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n"
)

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("1"),
		tgbotapi.NewKeyboardButton("2"),
		tgbotapi.NewKeyboardButton("3"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("4"),
		tgbotapi.NewKeyboardButton("5"),
		tgbotapi.NewKeyboardButton("6"),
	),
)

type server struct {
	bot        *tgbotapi.BotAPI
	userToGame map[int64]*Game
	codeToGame map[int]*Game
}

func (serv *server) sendMessage(msg tgbotapi.MessageConfig) tgbotapi.Message {
	for i := 0; i < 2; i++ {
		if message, err := serv.bot.Send(msg); err == nil {
			return message
			log.Panic(err)
		} else {
			log.Println("Unable to send from ", i, " trials")
			if i == 1 {
				return message
			}
		}
	}
	return tgbotapi.Message{}
}

func (serv *server) run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := serv.bot.GetUpdatesChan(u)

	for update := range updates {
		message := update.Message
		if message == nil { // ignore non-Message updates
			continue
		}

		text := update.Message.Text
		utfEncodedText := utf16.Encode([]rune(text))
		runeText := utf16.Decode(utfEncodedText)
		text = string(runeText)

		user := update.Message.Chat.ID // TODO: private messages
		if message.IsCommand() {
			handleCommand(serv, user, text)
		} else {
			if game := serv.userToGame[user]; game != nil && game.gameActive != nil {
				game.gameActive.handle(user, text)
			} else {
				serv.sendMessage(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы отправили не команду, и при этом Вы не в активной игре"))
			}
		}

		// switch update.Message.Text {
		// case "open":
		// 	msg.ReplyMarkup = numericKeyboard
		// case "close":
		// 	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		// }

		// serv.sendMessage(msg)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	serv := server{bot, make(map[int64]*Game), make(map[int]*Game)}
	serv.run()
}
