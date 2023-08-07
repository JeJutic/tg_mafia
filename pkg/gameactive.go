package game

import (
	"log"
)

// gameActive represents a started mafia game
type gameActive struct {
	*Game
	pQueue    []Player //will be deleted from slice if dead
	witnessed int64
	healed    int64
	voting    *voting
	night     *night
}

// newGameActive sets gActive field in Game g with pQueue parameter set
func (g *Game) newGameActive(pQueue []Player) {
	g.GActive = &gameActive{
		Game:   g,
		pQueue: pQueue,
	}
	g.GActive.newVoting(true)
}

func (ga *gameActive) roleToCnt() map[Role]int {
	roleToCnt := make(map[Role]int)
	for _, player := range ga.pQueue {
		roleToCnt[player.Role]++
	}
	return roleToCnt
}

func (ga *gameActive) mafiaAlive() bool {
	return ga.roleToCnt()[Mafia] != 0
}

func checkForEnd(playerCnt int, roleToCnt map[Role]int, isNight bool) Side {
	switch {
	case roleToCnt[Mafia] >= playerCnt-roleToCnt[Mafia] ||
		(isNight && playerCnt-2*roleToCnt[Mafia] == 1 && roleToCnt[Doctor] == 0):
		return MafiaSide
	case roleToCnt[Maniac] == 1 && playerCnt <= 2:
		return ManiacSide
	case roleToCnt[Mafia] == 0 && roleToCnt[Maniac] == 0:
		return PeacefulSide
	default:
		return 0
	}
}

func (ga *gameActive) checkForEnd() bool {

	if sideWinned := checkForEnd(len(ga.pQueue), ga.roleToCnt(), ga.night != nil); sideWinned != 0 {
		e := WinEvent{
			Users: ga.GetUsers(),
			Side:  sideWinned,
		}

		for _, player := range ga.pQueue {
			if roleToSide[player.Role] == sideWinned {
				e.Winners = append(e.Winners, ga.UserToNick[player.User])
			}
		}
		ga.EOutput.HandleWin(e)
		ga.StopGame(false)

		return true
	}
	return false
}

// func (gameActive *GameActive) updatePlayerQueue() { //TODO

// }

func (ga *gameActive) userToPlayer(user int64) Player {
	for _, player := range ga.pQueue {
		if player.User == user {
			return player
		}
	}
	log.Fatal("Unable to find the user")
	return Player{}
}

func remove(slice []Player, s int) []Player {
	return append(slice[:s], slice[s+1:]...)
}

func (ga *gameActive) removePlayer(user int64) {
	for i, player := range ga.pQueue {
		if player.User == user {
			ga.pQueue = remove(ga.pQueue, i)
			return
		}
	}
	log.Fatal("nobody to remove")
}

func (ga *gameActive) startDay() {
	ga.night = nil
	if ga.voting == nil {
		ga.newVoting(false)
	}

	if ga.checkForEnd() {
		return
	}

	e := VotingStartedEvent{
		make(map[int64][]string),
	}
	for _, player := range ga.pQueue {
		e.UserToCandidates[player.User] = ga.voting.userCanVote(player.User)
	}
	ga.EOutput.HandleVotingStarted(e)
}

func (ga *gameActive) startNight() {
	ga.voting = nil
	ga.newNight()

	if ga.checkForEnd() {
		return
	}

	ga.EOutput.HandleNightStarted(NightStartedEvent{
		ga.GetUsers(),
		ga.UserToNick[ga.pQueue[0].User],
	})
	ga.night.next()
}

func (ga *gameActive) votingConclusion() {

	candidate := ga.voting.bestCandidate()
	e := VotingEndedEvent{
		Users:       ga.GetUsers(),
		UserToNick:  ga.UserToNick,
		UserToVoted: ga.voting.userToVoted,
		Candidate:   candidate,
	}
	switch {
	case candidate == ga.witnessed && candidate != 0:
		e.Witness = true
	case candidate != 0:
		ga.removePlayer(candidate)
	}

	ga.EOutput.HandleVotingEnded(e)
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

func (ga *gameActive) Handle(user int64, request string) {
	if ga.voting != nil {
		ga.voting.handleVote(user, request)
	} else { //night
		ga.night.handleAct(user, request)
	}
}
