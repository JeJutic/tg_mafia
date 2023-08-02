package main

import (
	// "log"
	"math/rand"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCommand(serv *server, chatID int64, command string) {

	switch words := strings.Split(command, " "); words[0][1:] {
	case "create":
		code := 100_000 + rand.Intn(900_000) // TODO: check if not busy

		roles, err := parseRoles(words[1:])
		if err != "" {
			serv.sendMessage(tgbotapi.NewMessage(chatID, "Не получилось распарсить роли: "+err))
			return
		}

		game := NewGame(serv, code, chatID, roles)
		serv.codeToGame[code] = game

		serv.sendMessage(tgbotapi.NewMessage(chatID, "Игра успешно создана. Чтобы присоединиться введите\n/join "+strconv.Itoa(code)+" /никнейм/"))

	case "join":
		if len(words) < 2 {
			serv.sendMessage(tgbotapi.NewMessage(chatID, "В команде не представлен Ваш код"))
			return
		}
		code, err := strconv.Atoi(words[1])
		if err != nil {
			serv.sendMessage(tgbotapi.NewMessage(chatID, "У кода невалидный формат"))
			return
		}
		if serv.userToGame[chatID] != nil {
			serv.sendMessage(tgbotapi.NewMessage(chatID, "Вы уже в игре"))
			return
		}
		if game := serv.codeToGame[code]; game != nil {
			if game.gameActive == nil {
				if len(words) < 3 {
					serv.sendMessage(tgbotapi.NewMessage(chatID, "В команде не представлен Ваш ник"))
					return
				}
				nick := words[2]
				if _, exists := game.nickToUser[nick]; exists {
					serv.sendMessage(tgbotapi.NewMessage(chatID, "В игре уже есть человек с таким ником"))
					return
				}
				serv.userToGame[chatID] = game
				serv.sendMessage(tgbotapi.NewMessage(chatID, "Вы успешно присоединились. Роли: "+outputRoles(game.roles)))
				game.addMember(chatID, nick)
			} else {
				serv.sendMessage(tgbotapi.NewMessage(chatID, "Игра уже началась"))
			}
		} else {
			serv.sendMessage(tgbotapi.NewMessage(chatID, "Код не валиден"))
		}

	case "stop":
		if game := serv.userToGame[chatID]; game != nil {
			game.sendAll("Игра прервана")
			game.stopGame()
		} else {
			serv.sendMessage(tgbotapi.NewMessage(chatID, "Вы не в игре"))
		}
	}

}
