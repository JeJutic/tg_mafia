package gameserver

import (
	_ "embed"
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

//go:embed startText.txt
var startText string

func handleCommand[T any](ms mafiaServer[T], msg UserMessage) {

	switch words := strings.Split(msg.Text, " "); words[0][1:] {
	case "create":
		ms.mu.Lock()
		defer ms.mu.Unlock()

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
				delete(ms.userToGameCode, user)
			}
		}
		game := game.NewGame(ms, code, msg.User, roles, close)
		ms.codeToGame[code] = game

		ms.SendMessage(ServerMessage{
			User: msg.User,
			Text: "Игра успешно создана. Чтобы присоединиться введите\n/join " + strconv.Itoa(code) + " /никнейм/",
		})

	case "join":
		ms.mu.Lock()
		defer ms.mu.Unlock()

		if len(words) < 2 {
			ms.SendMessage(newMessage(msg.User, "В команде не представлен Ваш код или название группы", false))
			return
		}
		code, err := strconv.Atoi(words[1])
		if err != nil {
			ms.SendMessage(newMessage(msg.User, "У кода невалидный формат", false))
			return
		}
		if ms.userToGame(msg.User) != nil {
			ms.SendMessage(newMessage(msg.User, "Вы уже в игре", false))
			return
		}
		if game := ms.codeToGame[code]; game != nil {
			if !game.Started() {

				var nick string
				switch {
				case len(words) < 3:
					nick = ms.GetDefaultNick(msg.User)
				case len(words) > 3:
					ms.SendMessage(newMessage(msg.User, "Ник может состоять только из одного слова", false))
					return
				default:
					nick = words[2]
				}
				if !validNick(nick) {
					ms.SendMessage(newMessage(msg.User, "Ник не валиден", false))
					return
				}

				if err := game.AddMember(msg.User, nick); err != nil {
					ms.SendMessage(newMessage(msg.User,
						"Кажется, в игре уже есть человек с таким ником: "+err.Error(),
						false,
					))
					return
				}

				ms.userToGameCode[msg.User] = code
				text := "Вы успешно присоединились. Роли: " + rolesToString(game.Roles) + "\n\n"
				for nick := range game.NickToUser {
					text += nick + "\n"
				}
				ms.SendMessage(newMessage(msg.User, text, true))
				game.Start(game.RandomPlayerQueue()) // tries to start, ignores the error
			} else {
				ms.SendMessage(newMessage(msg.User, "Игра уже началась", false))
			}
		} else {
			ms.SendMessage(newMessage(msg.User, "Код не валиден", false))
		}

	case "stop":
		ms.mu.Lock()
		defer ms.mu.Unlock()

		if game := ms.userToGame(msg.User); game != nil {
			game.StopGame(true)
		} else {
			ms.SendMessage(newMessage(msg.User, "Вы не в игре", true))
		}

	case "start":
		ms.SendMessage(newMessage(msg.User, startText, false))

	case "newgroup":
		if len(words) < 2 {
			ms.SendMessage(newMessage(msg.User, "Вы не указали название группы", false))
			return
		}
		group := words[1]
		err := ms.groups.createGroup(msg.User, group)
		if err != nil {
			ms.SendMessage(newMessage(msg.User, err.Error(), false))
			return
		}
		ms.SendMessage(newMessage(msg.User, "Вы успешно создали группу "+group, false))

	case "group":
		if len(words) < 2 {
			ms.SendMessage(newMessage(msg.User, "В команде не представлена группа", false))
			return
		}
		group := words[1]
		err := ms.groups.joinGroup(msg.User, group)
		if err != nil {
			ms.SendMessage(newMessage(msg.User, "Не удалось присоединиться к группе: "+err.Error(), false))
		} else {
			ms.SendMessage(newMessage(
				msg.User,
				"Вы успешно присоединились к группе "+group+"\n\nНапишите /members "+
					group+", чтобы узнать участников",
				false,
			))
		}
		return

	case "invite":
		if code := ms.userToGameCode[msg.User]; code != 0 {
			if len(words) < 2 {
				ms.SendMessage(newMessage(msg.User, "В команде не представлена группа", false))
				return
			}
			group := words[1]
			users, err := ms.groups.getGroupMembers(msg.User, group)
			if err != nil {
				ms.SendMessage(newMessage(msg.User, err.Error(), false))
				return
			}
			if game := ms.codeToGame[code]; game != nil && game.Started() { // game != nil just for safety
				ms.SendMessage(newMessage(msg.User, "Игра уже началась", false))
				return
			}
			for _, user := range users {
				if user == msg.User {
					continue
				}

				ms.SendMessage(ServerMessage{
					User: user,
					Text: ms.GetDefaultNick(user) + " как участник " + group +
						" приглашает Вас в игру " + strconv.Itoa(code),
					Options: []string{
						"/join" + strconv.Itoa(code),
					},
				})
			}
		} else {
			ms.SendMessage(newMessage(msg.User, "Вы должны находиться в игре", false))
		}

	case "members":
		if len(words) < 2 {
			ms.SendMessage(newMessage(msg.User, "В команде не представлена группа", false))
			return
		}

		members, err := ms.groups.getGroupMembers(msg.User, words[1])
		if err != nil {
			ms.SendMessage(newMessage(msg.User, "Не получилось получить состав группы: "+err.Error(), false))
			return
		}
		var nicks []string
		for _, member := range members {
			nicks = append(nicks, ms.GetDefaultNick(member))
		}
		ms.SendMessage(newMessage(
			msg.User,
			"Группа "+words[1]+" состоит из:\n\n"+strings.Join(nicks, "\n"),
			false,
		))

	default:
		ms.SendMessage(newMessage(msg.User, "Неизвестная команда: "+words[0], false))
	}
}
