package game

import (
	"math/rand"
)

type Game struct {
	EOutput    EventOutput
	creator    int64
	NickToUser map[string]int64
	UserToNick map[int64]string
	Roles      []Role
	GActive    *GameActive
	close      func(*Game)
}

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

func (g *Game) AddMember(member int64, nick string) {
	g.NickToUser[nick] = member // TODO: check if not busy
	g.UserToNick[member] = nick

	if len(g.NickToUser) == len(g.Roles) {
		g.NewGameActive(g.RandomPlayerQueue())

		g.EOutput.HandleFirstDay(FirstDayEvent{
			g.UserToNick,
			g.GActive.pQueue,
		})
		g.GActive.startDay()
	}
}

type Player struct {
	User int64
	Role Role
}

// func (g *Game) isMafiaAlive() bool {	//TODO: maniac
// 	mafiaAlive := false
// 	for role := range g.roles {
// 		if role == MAFIA_ROLE {
// 			mafiaAlive = true
// 		}
// 	}
// 	return mafiaAlive
// }

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

func (g *Game) GetUsers() (users []int64) {
	for _, user := range g.NickToUser {
		users = append(users, user)
	}
	return users
}

// can happen both ways implicitly or explicitly
func (g *Game) StopGame(notify bool) {
	g.close(g)

	if notify {
		g.EOutput.HandleNotifyStopGame(NotifyStopGameEvent{
			g.GetUsers(),
		})
	}
}
