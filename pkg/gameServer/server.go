package gameServer

import (
	game "github.com/jejutic/tg_mafia/pkg/game"
)

type UserMessage struct {
	User    int64
	Text    string
	Command bool
}

type ServerMessage struct {
	User    int64
	Text    string
	Options []string
}

func newMessageKeepOptions(user int64, text string) ServerMessage {
	return ServerMessage{
		User: user,
		Text: text,
	}
}

func newMessageRemoveOptions(user int64, text string) ServerMessage {
	return ServerMessage{
		User:    user,
		Text:    text,
		Options: make([]string, 0),
	}
}

type Server[T any] interface {
	GetUpdatesChan() <-chan T
	UpdateToMessage(T) *UserMessage // didn't want to make an extra goroutine for casting of updates from chan
	SendMessage(ServerMessage)
	GetDefaultNick(int64) string
}

func sendAll[T any](s Server[T], users []int64, text string, removeOptions bool) {
	for _, user := range users {
		s.SendMessage(newMessageRemoveOptions(user, text))
	}
}

type mafiaServer[T any] struct {
	Server[T]
	userToGame map[int64]*game.Game
	codeToGame map[int]*game.Game
}

func NewMafiaServer[T any](s Server[T]) mafiaServer[T] {
	return mafiaServer[T]{
		Server:     s,
		userToGame: make(map[int64]*game.Game),
		codeToGame: make(map[int]*game.Game),
	}
}

func Run[T any](ms mafiaServer[T]) {

	for update := range ms.GetUpdatesChan() {
		msg := ms.UpdateToMessage(update)
		if msg == nil {
			continue
		}

		if msg.Command {
			handleCommand(ms, *msg)
		} else {
			if game := ms.userToGame[msg.User]; game != nil && game.Started() {
				game.GActive.Handle(msg.User, msg.Text)
			} else if game == nil {
				handleCommand(ms, UserMessage{
					User: msg.User,
					Text: "/join " + msg.Text,
				})
			} else {
				ms.SendMessage(ServerMessage{
					User: msg.User,
					Text: "Дождитесь начала игры",
				})
			}
		}
	}
}
