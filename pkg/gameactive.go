package game

import (
	"log"
)

type GameActive struct {
	game      *Game
	pQueue    []Player //will be deleted from slice if dead
	witnessed int64
	healed    int64
	voting    *Voting
	night     *Night
}

func (g *Game) NewGameActive(pQueue []Player) {
	g.GActive = &GameActive{
		game:   g,
		pQueue: pQueue,
	}
	g.GActive.NewVoting(true)
}

func (ga *GameActive) checkForEnd() bool { //TODO: different for day and night
	mafiaCnt := 0
	for _, player := range ga.pQueue {
		if player.Role == Mafia {
			mafiaCnt++
		}
	}

	var roleWinned Role
	switch {
	case mafiaCnt >= len(ga.pQueue)-mafiaCnt:
		roleWinned = Mafia
	case mafiaCnt == 0:
		roleWinned = Peaceful
	default:
		roleWinned = -1
	}

	if roleWinned != -1 {
		e := WinEvent{
			Users: ga.game.GetUsers(),
			Role:  roleWinned,
		}

		for _, player := range ga.pQueue {
			if roleToSide[player.Role] == roleToSide[roleWinned] {
				e.Winners = append(e.Winners, ga.game.UserToNick[player.User])
			}
		}
		ga.game.EOutput.HandleWin(e)
		ga.game.StopGame(false)

		return true
	}
	return false
}

// func (gameActive *GameActive) updatePlayerQueue() { //TODO

// }

// func (ga *GameActive) getPlayerFromUser(user int64) player {
// 	for _, player := range ga.pQueue {
// 		if player.user == user {
// 			return player
// 		}
// 	}
// 	log.Fatal("Unable to find the user")
// 	return player{}
// }

func remove(slice []Player, s int) []Player {
	return append(slice[:s], slice[s+1:]...)
}

func (ga *GameActive) removePlayer(user int64) {
	for i, player := range ga.pQueue {
		if player.User == user {
			ga.pQueue = remove(ga.pQueue, i)
			return
		}
	}
	log.Fatal("nobody to remove")
}

func (ga *GameActive) startDay() {
	ga.night = nil
	if ga.voting == nil {
		ga.NewVoting(false)
	}

	if ga.checkForEnd() {
		return
	}

	e := VotingStartedEvent{
		make(map[int64][]string),
	}
	for _, player := range ga.pQueue {
		e.UserToCandidates[player.User] = ga.voting.memberCanVote(player.User)
	}
	ga.game.EOutput.HandleVotingStarted(e)
}

func (ga *GameActive) startNight() {
	ga.voting = nil
	ga.NewNight()

	if ga.checkForEnd() {
		return
	}

	ga.game.EOutput.HandleNightStarted(NightStartedEvent{
		ga.game.GetUsers(),
		ga.game.UserToNick[ga.pQueue[0].User],
	})
	ga.night.next()
}

func (ga *GameActive) votingConclusion() {

	candidate := ga.voting.bestCandidate()
	e := VotingEndedEvent{
		Users:       ga.game.GetUsers(),
		UserToNick:  ga.game.UserToNick,
		UserToVoted: ga.voting.userToVoted,
		Candidate:   candidate,
	}
	switch {
	case candidate == ga.witnessed:
		e.Witness = true
	case candidate != 0:
		ga.removePlayer(candidate)
	}

	ga.game.EOutput.HandleVotingEnded(e)
	ga.startNight()
}

func contains(elems []string, target string) bool {
	for _, elem := range elems {
		if elem == target {
			return true
		}
	}
	return false
}

func (ga *GameActive) Handle(user int64, request string) {
	if ga.voting != nil {
		ga.voting.handleVote(user, request)
	} else { //night
		ga.night.handleAct(user, request)
	}
}
