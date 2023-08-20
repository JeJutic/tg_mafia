package gameServer

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"

	game "github.com/jejutic/tg_mafia/pkg/game"
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

func handleCommand[T any](ms mafiaServer[T], msg UserMessage) {

	switch words := strings.Split(msg.Text, " "); words[0][1:] {
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
			ms.SendMessage(ServerMessage{
				User: msg.User,
				Text: "Не получилось распарсить роли: " + err.Error(),
			})
			return
		}

		close := func(game *game.Game) { //closure
			ms.codeToGame[code] = nil
			for _, user := range game.NickToUser {
				ms.userToGame[user] = nil
			}
		}
		game := game.NewGame(ms, code, msg.User, roles, close)
		ms.codeToGame[code] = game

		ms.SendMessage(ServerMessage{
			User: msg.User,
			Text: "Игра успешно создана. Чтобы присоединиться введите\n/join " + strconv.Itoa(code) + " /никнейм/",
		})

	case "join":
		if len(words) < 2 {
			ms.SendMessage(newMessageKeepOptions(msg.User, "В команде не представлен Ваш код"))
			return
		}
		code, err := strconv.Atoi(words[1])
		if err != nil {
			ms.SendMessage(newMessageKeepOptions(msg.User, "У кода невалидный формат"))
			return
		}
		if ms.userToGame[msg.User] != nil {
			ms.SendMessage(newMessageKeepOptions(msg.User, "Вы уже в игре"))
			return
		}
		if game := ms.codeToGame[code]; game != nil {
			if !game.Started() {

				var nick string
				switch {
				case len(words) < 3:
					nick = ms.GetDefaultNick(msg.User)
				case len(words) > 3:
					ms.SendMessage(newMessageKeepOptions(msg.User, "Ник может состоять только из одного слова"))
					return
				default:
					nick = words[2]
				}
				if !validNick(nick) {
					ms.SendMessage(newMessageKeepOptions(msg.User, "Ник не валиден"))
					return
				}

				if err := game.AddMember(msg.User, nick); err != nil {
					ms.SendMessage(newMessageKeepOptions(msg.User,
						"Кажется, в игре уже есть человек с таким ником: "+err.Error(),
					))
					return
				}

				ms.userToGame[msg.User] = game
				message := "Вы успешно присоединились. Роли: " + rolesToString(game.Roles) + "\n\n"
				for nick := range game.NickToUser {
					message += nick + "\n"
				}
				ms.SendMessage(newMessageKeepOptions(msg.User, message))
				go game.Start(game.RandomPlayerQueue()) // tries to start, ignores the error
			} else {
				ms.SendMessage(newMessageKeepOptions(msg.User, "Игра уже началась"))
			}
		} else {
			ms.SendMessage(newMessageKeepOptions(msg.User, "Код не валиден"))
		}

	case "stop":
		if game := ms.userToGame[msg.User]; game != nil {
			game.StopGame(true)
		} else {
			ms.SendMessage(newMessageKeepOptions(msg.User, "Вы не в игре"))
		}

	default:
		ms.SendMessage(newMessageKeepOptions(msg.User, "Неизвестная команда: "+words[0]))
	}
}