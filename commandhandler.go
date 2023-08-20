package main

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"

	game "github.com/jejutic/tg_mafia/pkg"
)

func validNick(nick string) bool {
	return !strings.Contains(nick, "\n") && nick != ""
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
		case "разгадыватель", "раз", "guesser":
			roles[i] = game.Guesser
		default:
			err := errors.New("Неизвестный токен роли: " + token)
			return roles, err
		}
	}
	return roles, game.ValidRoles(roles)
}

func handleCommand[T any](ms mafiaServer[T], msg userMessage) {

	switch words := strings.Split(msg.text, " "); words[0][1:] {
	case "create":
		var code int
		for {
			code = 1_000 + rand.Intn(9_000)
			if _, exists := ms.codeToGame[code]; !exists {
				break
			}
		}

		roles, err := parseRoles(words[1:])
		if err != nil {
			ms.sendMessage(serverMessage{
				user: msg.user,
				text: "Не получилось распарсить роли: " + err.Error(),
			})
			return
		}

		close := func(game *game.Game) { //closure
			ms.codeToGame[code] = nil
			for _, user := range game.NickToUser {
				ms.userToGame[user] = nil
			}
		}
		game := game.NewGame(ms, code, msg.user, roles, close)
		ms.codeToGame[code] = game

		ms.sendMessage(serverMessage{
			user: msg.user,
			text: "Игра успешно создана. Чтобы присоединиться введите\n/join " + strconv.Itoa(code) + " /никнейм/",
		})

	case "join":
		if len(words) < 2 {
			ms.sendMessage(newMessageKeepKeyboard(msg.user, "В команде не представлен Ваш код"))
			return
		}
		code, err := strconv.Atoi(words[1])
		if err != nil {
			ms.sendMessage(newMessageKeepKeyboard(msg.user, "У кода невалидный формат"))
			return
		}
		if ms.userToGame[msg.user] != nil {
			ms.sendMessage(newMessageKeepKeyboard(msg.user, "Вы уже в игре"))
			return
		}
		if game := ms.codeToGame[code]; game != nil {
			if game.GActive == nil {

				var nick string
				switch {
				case len(words) < 3:
					nick = ms.getDefaultNick(msg.user)
				case len(words) > 3:
					ms.sendMessage(newMessageKeepKeyboard(msg.user, "Ник может состоять только из одного слова"))
					return
				default:
					nick = words[2]
				}
				if !validNick(nick) {
					ms.sendMessage(newMessageKeepKeyboard(msg.user, "Ник не валиден"))
					return
				}

				if err := game.AddMember(msg.user, nick); err != nil {
					ms.sendMessage(newMessageKeepKeyboard(msg.user,
						"Кажется, в игре уже есть человек с таким ником: "+err.Error(),
					))
					return
				}

				ms.userToGame[msg.user] = game
				message := "Вы успешно присоединились. Роли: " + rolesToString(game.Roles) + "\n\n"
				for nick := range game.NickToUser {
					message += nick + "\n"
				}
				ms.sendMessage(newMessageKeepKeyboard(msg.user, message))
				game.Start(game.RandomPlayerQueue()) // tries to start, ignores the error
			} else {
				ms.sendMessage(newMessageKeepKeyboard(msg.user, "Игра уже началась"))
			}
		} else {
			ms.sendMessage(newMessageKeepKeyboard(msg.user, "Код не валиден"))
		}

	case "stop":
		if game := ms.userToGame[msg.user]; game != nil {
			game.StopGame(true)
		} else {
			ms.sendMessage(newMessageKeepKeyboard(msg.user, "Вы не в игре"))
		}

	default:
		ms.sendMessage(newMessageKeepKeyboard(msg.user, "Неизвестная команда: "+words[0]))
	}
}
