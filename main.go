package main

import (
	"log"
	"os"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	run(newMafiaServer[tgbotapi.Update](s))
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

func (tbs tgBotServer) getDefaultNick(user int64) string {
	chat, err := tbs.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: user},
	})
	if err != nil {
		return "unspecified"
	}
	return chat.UserName
}
