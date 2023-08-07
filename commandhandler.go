package main

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jejutic/tg_mafia/pkg"
)

func validNick(nick string) bool {
	return !strings.Contains(nick, "\n") && nick != ""
}

func getDefaultNick(bot *tgbotapi.BotAPI, chatID int64) string {
	chat, err := bot.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
	})
	if err != nil {
		return "unspecified"
	}
	return chat.UserName
}

func parseRoles(tokens []string) ([]game.Role, error) {

	roles := make([]game.Role, len(tokens))
	for i, token := range tokens {
		switch token {
		case "мафия", "маф":
			roles[i] = game.Mafia
		case "мирный", "мир":
			roles[i] = game.Peaceful
		case "врач", "доктор", "док":
			roles[i] = game.Doctor
		case "свидетельница", "свид":
			roles[i] = game.Witness
		case "комиссар", "ком", "шериф":
			roles[i] = game.Sheriff
		case "маньяк", "ман", "убийца":
			roles[i] = game.Maniac
		default:
			err := errors.New("Неизвестный токен роли: " + token)
			return roles, err
		}
	}
	return roles, game.ValidRoles(roles)
}

func (s *server) handleCommand(chatID int64, command string) {

	switch words := strings.Split(command, " "); words[0][1:] {
	case "create":
		var code int
		for  {
			code = 1_000 + rand.Intn(9_000)
			if _, exists := s.codeToGame[code]; !exists {
				break
			}
		}

		roles, err := parseRoles(words[1:])
		if err != nil {
			s.sendMessage(tgbotapi.NewMessage(chatID, "Не получилось распарсить роли: "+err.Error()))
			return
		}

		close := func(game *game.Game) { //closure
			s.codeToGame[code] = nil
			for _, user := range game.NickToUser {
				s.userToGame[user] = nil
			}
		}
		game := game.NewGame(s, code, chatID, roles, close)
		s.codeToGame[code] = game

		s.sendMessage(tgbotapi.NewMessage(chatID, "Игра успешно создана. Чтобы присоединиться введите\n/join "+strconv.Itoa(code)+" /никнейм/"))

	case "join":
		if len(words) < 2 {
			s.sendMessage(tgbotapi.NewMessage(chatID, "В команде не представлен Ваш код"))
			return
		}
		code, err := strconv.Atoi(words[1])
		if err != nil {
			s.sendMessage(tgbotapi.NewMessage(chatID, "У кода невалидный формат"))
			return
		}
		if s.userToGame[chatID] != nil {
			s.sendMessage(tgbotapi.NewMessage(chatID, "Вы уже в игре"))
			return
		}
		if game := s.codeToGame[code]; game != nil {
			if game.GActive == nil {

				var nick string
				switch {
				case len(words) < 3:
					nick = getDefaultNick(s.bot, chatID)
				case len(words) > 3:
					s.sendMessage(tgbotapi.NewMessage(chatID, "Ник может состоять только из одного слова"))
					return
				default:
					nick = words[2]
				}
				if !validNick(nick) {
					s.sendMessage(tgbotapi.NewMessage(chatID, "Ник не валиден"))
					return
				}

				if err := game.AddMember(chatID, nick); err != nil {
					s.sendMessage(tgbotapi.NewMessage(chatID, 
						"Кажется, в игре уже есть человек с таким ником: " + err.Error(),
					))
					return
				}

				s.userToGame[chatID] = game
				message := "Вы успешно присоединились. Роли: " + rolesToString(game.Roles) + "\n\n"
				for nick := range game.NickToUser {
					message += nick + "\n"
				}
				s.sendMessage(tgbotapi.NewMessage(chatID, message))
				game.Start(game.RandomPlayerQueue())	// tries to start, ignores the error
			} else {
				s.sendMessage(tgbotapi.NewMessage(chatID, "Игра уже началась"))
			}
		} else {
			s.sendMessage(tgbotapi.NewMessage(chatID, "Код не валиден"))
		}

	case "stop":
		if game := s.userToGame[chatID]; game != nil {
			game.StopGame(true)
		} else {
			s.sendMessage(tgbotapi.NewMessage(chatID, "Вы не в игре"))
		}

	default:
		s.sendMessage(tgbotapi.NewMessage(chatID, "Неизвестная команда: "+words[0]))
	}

}
