package gameserver

import (
	"sync"

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

func newMessage(user int64, text string, removeOptions bool) ServerMessage {
	msg := ServerMessage{
		User: user,
		Text: text,
	}
	if removeOptions {
		msg.Options = make([]string, 0)
	}
	return msg
}

type Server[T any] interface {
	GetUpdatesChan() <-chan T
	UpdateToMessage(T) *UserMessage // didn't want to make an extra goroutine for casting of updates from chan
	SendMessage(ServerMessage)
	GetDefaultNick(int64) string
}

func sendAll[T any](s Server[T], users []int64, text string, removeOptions bool) {
	for _, user := range users {
		s.SendMessage(newMessage(user, text, removeOptions))
	}
}

type mafiaServer[T any] struct {
	Server[T]
	groups         groupStorage
	userToGameCode map[int64]int
	codeToGame     map[int]*game.Game
	mu             *sync.Mutex
}

func NewMafiaServer[T any](s Server[T], driverName string, dbUrl string) mafiaServer[T] {
	return mafiaServer[T]{
		Server: s,
		groups: &groupsDb{
			driverName,
			dbUrl,
		},
		userToGameCode: make(map[int64]int),
		codeToGame:     make(map[int]*game.Game),
		mu:             &sync.Mutex{},
	}
}

func (ms mafiaServer[T]) userToGame(user int64) *game.Game {
	return ms.codeToGame[ms.userToGameCode[user]]
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
			if game := ms.userToGame(msg.User); game != nil && game.Started() {
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
