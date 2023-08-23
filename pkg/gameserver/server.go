package gameserver

import (
	"log"
	"sync"

	game "github.com/jejutic/tg_mafia/pkg/game"
)

// UserMessage represents message sent from user to server
type UserMessage struct {
	User    int64  // user sent
	Text    string // text
	Command bool   // if message is a command
}

// ServerMessage represents message sent from server to user
type ServerMessage struct {
	User    int64    // receiver of message
	Text    string   // text
	Options []string // options for user; if Options isn't nil and have 0 len, means previous options should be hidden
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

// Server represents any endpoint user can chat with. Type T is an update from user format
// which is suitable for sending in chan
type Server[T any] interface {
	GetUpdatesChan() <-chan T       // channel of updates from user
	UpdateToMessage(T) *UserMessage // transforms update to standard format
	SendMessage(ServerMessage)      // sends message to user
	GetDefaultNick(int64) string    // gets default user's nickname
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
	mu             *sync.RWMutex
}

// NewMafiaServer creates new instance of server for mafia game.
// s is a server that will be used internally.
// driverName and dbUrl specify connection to database/sql.DB.
// initDb specifies if "CREATE TABLE IF NOT EXISTS" should be called for tables needed internally.
func NewMafiaServer[T any](s Server[T], driverName string, dbUrl string, initDb bool) mafiaServer[T] {
	groups := &groupsDb{
		driverName,
		dbUrl,
	}
	if initDb {
		if err := groups.initDb(); err != nil {
			log.Println(err)
		}
	}
	return mafiaServer[T]{
		Server:         s,
		groups:         groups,
		userToGameCode: make(map[int64]int),
		codeToGame:     make(map[int]*game.Game),
		mu:             &sync.RWMutex{},
	}
}

func (ms mafiaServer[T]) userToGame(user int64) *game.Game {
	return ms.codeToGame[ms.userToGameCode[user]]
}

// Run starts listening mafiaServer
func Run[T any](ms mafiaServer[T]) {

	for update := range ms.GetUpdatesChan() {
		msg := ms.UpdateToMessage(update)
		if msg == nil {
			continue
		}

		if msg.Command {
			go ms.handleCommand(*msg)
		} else {
			if game := ms.userToGame(msg.User); game != nil && game.Started() {
				game.GActive.Handle(msg.User, msg.Text)
			} else if game == nil {
				go ms.handleCommand(UserMessage{
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
