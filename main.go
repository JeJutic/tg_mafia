package main

import (
	"log"
	"os"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jejutic/tg_mafia/pkg"
)

type server struct {
	bot        *tgbotapi.BotAPI	//TODO: use interface
	userToGame map[int64]*game.Game
	codeToGame map[int]*game.Game
}

func (s *server) run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)

	for update := range updates {
		message := update.Message
		if message == nil { // ignore non-Message updates
			continue
		}

		text := update.Message.Text
		utfEncodedText := utf16.Encode([]rune(text))
		runeText := utf16.Decode(utfEncodedText)
		text = string(runeText)

		user := update.Message.Chat.ID
		if message.IsCommand() {
			s.handleCommand(user, text)
		} else {
			if game := s.userToGame[user]; game != nil && game.GActive != nil {
				game.GActive.Handle(user, text)
			} else if game == nil {
				s.handleCommand(user, "/join "+text)
				// s.sendMessage(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы отправили не команду, и при этом Вы не в активной игре"))
			} else {
				s.sendMessage(tgbotapi.NewMessage(update.Message.Chat.ID, "Дождитесь начала игры"))
			}
		}
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	s := server{bot, make(map[int64]*game.Game), make(map[int]*game.Game)}
	s.run()
}
