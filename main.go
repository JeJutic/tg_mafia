package main

import (
	"log"
	"os"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	gameServer "github.com/jejutic/tg_mafia/pkg/gameserver"
)

type tgBotServer struct {
	*tgbotapi.BotAPI
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	s := tgBotServer{bot}
	gameServer.Run[tgbotapi.Update](gameServer.NewMafiaServer[tgbotapi.Update](
		s,
		"pgx",
		os.Getenv("POSTGRES_URI"),
		true,
	))
}

func (tbs tgBotServer) GetUpdatesChan() <-chan tgbotapi.Update {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return tbs.BotAPI.GetUpdatesChan(u)
}

func (tbs tgBotServer) UpdateToMessage(update tgbotapi.Update) *gameServer.UserMessage {
	if update.Message == nil { // ignore non-Message updates
		return nil
	}

	text := update.Message.Text
	utfEncodedText := utf16.Encode([]rune(text))
	runeText := utf16.Decode(utfEncodedText)
	text = string(runeText)

	return &gameServer.UserMessage{
		User:    update.Message.Chat.ID,
		Text:    text,
		Command: update.Message.IsCommand(),
	}
}

func (tbs tgBotServer) SendMessage(msg gameServer.ServerMessage) {
	msgConfig := tgbotapi.NewMessage(msg.User, msg.Text)

	if msg.Options != nil {
		if len(msg.Options) == 0 {
			msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		} else {
			var keyboard [][]tgbotapi.KeyboardButton
			for _, c := range msg.Options {
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

func (tbs tgBotServer) GetDefaultNick(user int64) string {
	chat, err := tbs.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: user},
	})
	if err != nil {
		return "unspecified"
	}
	return chat.UserName
}
