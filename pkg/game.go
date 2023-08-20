package game

import (
	"errors"
	"math/rand"
)

// Game represents a mafia game either started or not
type Game struct {
	eOutput    EventOutput      // output of events happening during active game
	creator    int64            // user which created the game
	NickToUser map[string]int64 // mapping from nick in the game to user
	UserToNick map[int64]string // mapping from user to nick in the game
	Roles      []Role           // list of all role-cards in the game, its len should be equal to number of players
	GActive    *gameActive      // nil iff is game is not started, represents active game otherwise
	close      func(*Game)      // function which will be callbacked when stopping the game to free resources
}

// NewGame returns instance with fields specified in parameters, nil GActive and empty
// NickToUser and UserToNick maps
func NewGame(eOutput EventOutput, code int, creator int64, roles []Role, close func(*Game)) *Game {
	return &Game{
		eOutput,
		creator,
		make(map[string]int64),
		make(map[int64]string),
		roles,
		nil,
		close,
	}
}

// Started iff Game g has started
func (g *Game) Started() bool {
	return g.GActive != nil
}

// AddMember adds a player in Game g with user id as user and nick as game nickname
// if nick isn't already presented in the game. Returns non-nil error otherwise
func (g *Game) AddMember(user int64, nick string) error {
	if _, exists := g.NickToUser[nick]; exists {
		return errors.New("this nick is already presented in the game")
	}

	g.NickToUser[nick] = user
	g.UserToNick[user] = nick
	return nil
}

// Start starts Game g if number of players joined is as expected,
// returns non-nil error otherwise
func (g *Game) Start(pQueue []Player) error {
	if len(g.NickToUser) != len(g.Roles) {
		return errors.New("not enough members")
	}
	if len(pQueue) != len(g.Roles) {
		return errors.New("not enough players in queue")
	}

	g.initGameActive(pQueue)
	g.eOutput.HandleFirstDay(FirstDayEvent{
		g.UserToNick,
		g.GActive.pQueue,
	})
	go g.GActive.startDay()
	return nil
}

// Player represents player in an active game
type Player struct {
	User int64 // user ID
	Role Role  // player's role in the game
}

// RandomPlayerQueue returns a slice of players formed from current list of joined users
// and roles set in Game g.
func (g *Game) RandomPlayerQueue() []Player {
	rand.Shuffle(len(g.Roles), func(i, j int) {
		g.Roles[i], g.Roles[j] = g.Roles[j], g.Roles[i]
	})
	playerQueue := make([]Player, 0, len(g.NickToUser))
	for _, user := range g.NickToUser {
		playerQueue = append(playerQueue, Player{user, g.Roles[len(playerQueue)]})
	}
	rand.Shuffle(len(playerQueue), func(i, j int) {
		playerQueue[i], playerQueue[j] = playerQueue[j], playerQueue[i]
	})
	return playerQueue
}

// GetUsers returns list of joined users
func (g *Game) GetUsers() (users []int64) {
	for _, user := range g.NickToUser {
		users = append(users, user)
	}
	return users
}

// StopGame callbacks field close in Game g and notifies in the output of game
// if notify is true
func (g *Game) StopGame(notify bool) {
	g.close(g)

	if notify {
		g.eOutput.HandleNotifyStopGame(NotifyStopGameEvent{
			g.GetUsers(),
		})
	}
}
