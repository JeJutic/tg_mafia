package main

import (
	"log"
	"os"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	game "github.com/jejutic/tg_mafia/pkg"
)

type userMessage struct {
	user    int64
	text    string
	command bool
}

type serverMessage struct {
	user    int64
	text    string
	options []string
}

func newMessageKeepKeyboard(user int64, text string) serverMessage {
	return serverMessage{
		user: user,
		text: text,
	}
}

func newMessageRemoveKeyboard(user int64, text string) serverMessage {
	return serverMessage{
		user:    user,
		text:    text,
		options: make([]string, 0),
	}
}

type server[T any] interface {
	getUpdatesChan() <-chan T
	updateToMessage(T) *userMessage // didn't want to make an extra goroutine for casting of updates from chan
	sendMessage(serverMessage)
	getDefaultNick(int64) string
}

type mafiaServer[T any] struct {
	server[T]
	userToGame map[int64]*game.Game
	codeToGame map[int]*game.Game
}

func newMafiaServer[T any](s server[T]) mafiaServer[T] {
	return mafiaServer[T]{
		server:     s,
		userToGame: make(map[int64]*game.Game),
		codeToGame: make(map[int]*game.Game),
	}
}

type tgBotServer struct {
	*tgbotapi.BotAPI
}

func (tbs tgBotServer) getUpdatesChan() <-chan tgbotapi.Update {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return tbs.BotAPI.GetUpdatesChan(u)
}

func (tbs tgBotServer) updateToMessage(update tgbotapi.Update) *userMessage {
	if update.Message == nil { // ignore non-Message updates
		return nil
	}

	text := update.Message.Text
	utfEncodedText := utf16.Encode([]rune(text))
	runeText := utf16.Decode(utfEncodedText)
	text = string(runeText)

	return &userMessage{
		user:    update.Message.Chat.ID,
		text:    text,
		command: update.Message.IsCommand(),
	}
}

func (tbs tgBotServer) sendMessage(msg serverMessage) {
	msgConfig := tgbotapi.NewMessage(msg.user, msg.text)

	if msg.options != nil {
		if len(msg.options) == 0 {
			msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		} else {
			var keyboard [][]tgbotapi.KeyboardButton
			for _, c := range msg.options {
				keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(c)))
			}
			msgConfig.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)
		}
	}

	for i := 0; i < 2; i++ {
		if _, err := tbs.Send(msgConfig); err == nil {
			return
		} else {
			log.Println("Unable to send from ", i, " trials: ", err)
		}
	}
}

func sendAll[T any](s server[T], users []int64, text string) {
	for _, user := range users {
		s.sendMessage(newMessageRemoveKeyboard(user, text))
	}
}

func (tbs tgBotServer) getDefaultNick(user int64) string {
	chat, err := tbs.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: user},
	})
	if err != nil {
		return "unspecified"
	}
	return chat.UserName
}

func run[T any](ms mafiaServer[T]) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	for update := range ms.getUpdatesChan() {
		msg := ms.updateToMessage(update)
		if msg == nil {
			continue
		}

		if msg.command {
			handleCommand(ms, *msg)
		} else {
			if game := ms.userToGame[msg.user]; game != nil && game.GActive != nil {
				game.GActive.Handle(msg.user, msg.text)
			} else if game == nil {
				handleCommand(ms, userMessage{
					user: msg.user,
					text: "/join " + msg.text,
				})
			} else {
				ms.sendMessage(serverMessage{
					user: msg.user,
					text: "Дождитесь начала игры",
				})
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

	s := tgBotServer{bot}
	run(newMafiaServer[tgbotapi.Update](s))
}
