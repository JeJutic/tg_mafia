package game_test

import (
	// "os"
	// "runtime"
	"strings"
	"testing"
	"time"

	. "github.com/jejutic/tg_mafia/mocks/github.com/jejutic/tg_mafia/pkg/game"
	. "github.com/jejutic/tg_mafia/pkg/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func pbool(b bool) *bool {
	return &b
}

func userToNick(nickToUser map[string]int64) map[int64]string {
	result := make(map[int64]string)
	for nick, user := range nickToUser {
		result[user] = nick
	}
	return result
}

func rolesFromPlayers(players []Player) []Role {
	var roles []Role
	for _, player := range players {
		roles = append(roles, player.Role)
	}
	return roles
}

// func TestMain(m *testing.M) {
// 	runtime.GOMAXPROCS(1)
// 	code := m.Run()
// 	os.Exit(code)
// }

func handle(g *Game, user int64, request string) {
	g.GActive.Handle(user, request)
	time.Sleep(2 * time.Millisecond)
}

func Test_scenario1(t *testing.T) {
	assert := assert.New(t)
	mockEOutput := NewMockEventOutput(t)
	var closed *bool = pbool(false)

	nickToUser := map[string]int64{
		"a": 1,
		"b": 2,
		"c": 3,
		"d": 4,
		"e": 5,
		"f": 6,
	}
	players := []Player{
		{1, Mafia},
		{2, Mafia},
		{3, Peaceful},
		{4, Doctor},
		{5, Witness},
		{6, Peaceful},
	}

	g := NewGame(mockEOutput, 1, 1, rolesFromPlayers(players), func(g *Game) { *closed = true })
	for nick, user := range nickToUser {
		assert.Nil(g.AddMember(user, nick), "expected no error when inserting players")
	}
	assert.NotNil(g.AddMember(7, "e"), "expected to fail when inserting repeated nick")

	mockEOutput.EXPECT().HandleFirstDay(FirstDayEvent{
		UserToNick: userToNick(nickToUser),
		Players:    players,
	}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(6, len(candidates), "candidates for voting number")
			}
		}).Once()
	g.Start(players)

	mockEOutput.EXPECT().HandleUnableToVote(UnableToVoteEvent{
		User: 1,
	}).Once()
	handle(g, 1, "a")

	handle(g, 1, "c")
	handle(g, 2, "a")
	handle(g, 3, "skip")
	mockEOutput.EXPECT().HandleAlreadyVoted(AlreadyVotedEvent{
		User: 3,
	})
	handle(g, 3, "a")
	handle(g, 4, "a")
	handle(g, 5, "b")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.Equal(userToNick(nickToUser), e.UserToNick, "nick to user mapping has changed")
			assert.Equal(map[int64]int64{
				1: 3,
				2: 1,
				3: -1,
				4: 1,
				5: 2,
				6: -1,
			}, e.UserToVoted)
			assert.EqualValues(0, e.Candidate, "draw expected")
			assert.Equal(false, e.Witness, "witness couldn't do anything at the first night")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).
		Run(func(e NightStartedEvent) {
			assert.Equal(6, len(e.Users), "unexpected number of users")
			assert.Equal(1, len(e.FirstToWake), "unrecognized user to wake")
		}).Once()

	acted := pbool(false)
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victimCnt int
			var victim string
			switch e.Player.User {
			case 1:
				victimCnt = 4
				victim = "c"
			case 2:
				victimCnt = 4
				victim = "d"
			case 3:
				victimCnt = 1
				victim = e.Victims[0]
			case 4:
				victimCnt = 6
				victim = "de"
			case 5:
				victimCnt = 6
				victim = "a"
			case 6:
				victimCnt = 1
				victim = e.Victims[0]
			}
			assert.Equal(victimCnt, len(e.Victims), "unexpected number of victims (buttons)")

			if *acted == false { // UnexpectedActTrialEvent
				if e.Player.User == 1 {
					handle(g, 2, "c")
				} else {
					handle(g, 1, "c")
				}
				*acted = true
			}
			handle(g, e.Player.User, victim)
		}).Times(6)
	mockEOutput.EXPECT().HandleUnexpectedActTrial(mock.Anything).Once()
	mockEOutput.EXPECT().HandleUnsupportedAct(UnsupportedActEvent{
		User: 4,
	}).Run(func(e UnsupportedActEvent) {
		handle(g, 4, "e")
	}).Once()

	unsuccessfulShot := pbool(false)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			switch e.Player.Role {
			case Mafia:
				if !e.Success {
					*unsuccessfulShot = true
				}
			default:
				assert.Equal(true, e.Success, "action expected to be successful")
			}
		}).Times(6)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(0, len(e.Died), "unexpected player dead")
		}).Once()

	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(6, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	assert.Equal(true, *unsuccessfulShot, "shot expected to be unsuccessful")

	handle(g, 1, "skip")
	handle(g, 2, "a")
	handle(g, 3, "a")
	handle(g, 4, "a")
	handle(g, 5, "a")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(1, e.Candidate, "unexpected candidate won")
			assert.Equal(true, e.Witness, "witness should have protected")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).
		Run(func(e NightStartedEvent) {
			assert.Equal(1, len(e.FirstToWake), "unrecognized user to wake")
		}).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			var victimCnt int
			switch e.Player.Role {
			case Mafia:
				victimCnt = 4
				victim = "c"
			case Doctor:
				assert.NotContains(e.Victims, "e")
				victimCnt = 5
				victim = "a"
			case Witness:
				assert.NotContains(e.Victims, "a")
				victimCnt = 5
				victim = "b"
			default:
				victimCnt = 1
				victim = e.Victims[0]
			}
			assert.Equal(victimCnt, len(e.Victims), "unexpected number of victims (buttons)")
			handle(g, e.Player.User, victim)
		}).Times(6)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			assert.Equal(true, e.Success, "all actions are supposed to be successful")
		}).Times(6)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(1, len(e.Died), "unexpected number of people dead")
		}).Once()

	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(5, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 1, "b")
	handle(g, 2, "a")
	handle(g, 4, "b")
	handle(g, 5, "b")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(2, e.Candidate, "unexpected candidate won")
			assert.Equal(true, e.Witness, "witness should have protected")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Mafia:
				victim = "d"
			case Doctor:
				victim = "d"
			case Witness:
				victim = "e"
			default:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(5)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			assert.Equal(true, e.Success, "all actions are supposed to be successful")
		}).Times(5)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(0, len(e.Died), "the person should have been saved by doctor")
		}).Once()

	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(5, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 1, "b")
	handle(g, 2, "a")
	handle(g, 4, "b")
	handle(g, 5, "b")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(2, e.Candidate, "unexpected candidate won")
			assert.Equal(false, e.Witness, "witness should not have protected")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Mafia:
				victim = "f"
			case Doctor:
				victim = "e"
			case Witness:
				victim = "d"
			default:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(4)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			assert.Equal(true, e.Success, "all actions are supposed to be successful")
		}).Times(4)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(1, len(e.Died), "unexpected number of people dead")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(3, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 1, "skip")
	handle(g, 4, "skip")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(0, e.Candidate, "expected draw")
			assert.Equal(false, e.Witness, "witness can't protect draw")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Mafia:
				victim = "d"
			case Doctor:
				victim = "d"
			case Witness:
				victim = "e"
			default:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(3)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			assert.Equal(true, e.Success, "all actions are supposed to be successful")
		}).Times(3)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(0, len(e.Died), "doctor was supposed to save himself")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(3, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 5, "skip")

	handle(g, 1, "skip")
	handle(g, 4, "a")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(1, e.Candidate, "unexpected candidate")
			assert.Equal(false, e.Witness, "witness wasn't supposed to protect")
		}).Once()
	mockEOutput.EXPECT().HandleWin(mock.Anything).
		Run(func(e WinEvent) {
			assert.Equal(PeacefulSide, e.Side, "unexpected side won")
			assert.Equal(2, len(e.Winners), "unexpected game winners count")
		}).Once()

	assert.Equal(false, *closed, "game should not have been closed yet")
	handle(g, 5, "a")

	time.Sleep(100 * time.Millisecond)

	// mockEOutput.EXPECT().HandleNotifyStopGame(mock.Anything).Once()
	// g.StopGame(true)
	assert.Equal(true, *closed, "game should have been closed")
}

func containsIn(slice []string, s string) int {
	var res int
	for _, t := range slice {
		if strings.Contains(t, s) {
			res++
		}
	}
	return res
}

func Test_scenario2(t *testing.T) {
	assert := assert.New(t)
	mockEOutput := NewMockEventOutput(t)

	nickToUser := map[string]int64{
		"a": 1,
		"b": 2,
		"c": 3,
		"d": 4,
		"e": 5,
		"f": 6,
	}
	players := []Player{
		{1, Mafia},
		{2, Mafia},
		{3, Maniac},
		{4, Doctor},
		{5, Witness},
		{6, Sheriff},
	}

	g := NewGame(mockEOutput, 1, 1, rolesFromPlayers(players), func(g *Game) {})
	for nick, user := range nickToUser {
		assert.Nil(g.AddMember(user, nick), "expected no error when inserting players")
	}

	mockEOutput.EXPECT().HandleFirstDay(mock.Anything).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(6, len(candidates), "candidates for voting number")
			}
		}).Once()
	g.Start(players)

	handle(g, 1, "c")
	handle(g, 2, "a")
	handle(g, 3, "skip")
	handle(g, 4, "a")
	handle(g, 5, "b")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(0, e.Candidate, "expected draw")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Mafia:
				victim = "d"
			case Doctor:
				victim = "d"
			case Witness:
				victim = "e"
			case Sheriff:
				assert.Equal(11, len(e.Victims), "expected another cnt of sheriff's options")
				for nick, user := range nickToUser {
					if user != e.Player.User {
						assert.Equal(4, containsIn(e.Victims, nick), "expected another sheriff's options")
					}
				}
				victim = "a b"
			case Maniac:
				assert.Equal(1, len(e.Victims), "maniac has victims when mafia alive")
				victim = e.Victims[0]
			default:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(6)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			assert.Equal(true, e.Success, "all actions are supposed to be successful")
		}).Times(6)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(0, len(e.Died), "doctor was supposed to save himself")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(6, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 1, "c")
	handle(g, 2, "a")
	handle(g, 3, "skip")
	handle(g, 4, "a")
	handle(g, 5, "b")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(0, e.Candidate, "expected draw")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Mafia:
				victim = "e"
			case Doctor:
				victim = "a"
			case Witness:
				victim = "d"
			case Sheriff:
				victim = "d e"
			case Maniac:
				assert.Equal(1, len(e.Victims), "maniac has victims when mafia alive")
				victim = e.Victims[0]
			default:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(6)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			assert.Equal(true, e.Success, "all actions are supposed to be successful")
		}).Times(6)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(1, len(e.Died), "unexpected number of people dead")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(5, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 1, "c")
	handle(g, 2, "a")
	handle(g, 3, "skip")
	handle(g, 4, "a")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(0, e.Candidate, "expected draw")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Mafia:
				victim = "f"
			case Doctor:
				victim = "f"
			case Sheriff:
				victim = "a c"
			case Maniac:
				assert.Equal(1, len(e.Victims), "maniac has victims when mafia alive")
				victim = e.Victims[0]
			default:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(5)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			if e.Player.Role == Sheriff {
				assert.Equal(false, e.Success, "Mafia and maniac are supposed to be in different sides")
			} else {
				assert.Equal(true, e.Success, "all actions are supposed to be successful")
			}
		}).Times(5)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(0, len(e.Died), "unexpected player dead")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(5, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 1, "c")
	handle(g, 2, "a")
	handle(g, 3, "a")
	handle(g, 4, "a")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(1, e.Candidate, "unexpected candidate won")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Mafia:
				victim = "d"
			case Doctor:
				victim = "d"
			case Sheriff:
				assert.Equal(4, len(e.Victims), "expected another cnt of sheriff's options")
				victim = "b c"
			case Maniac:
				assert.Equal(1, len(e.Victims), "maniac has victims when mafia alive")
				victim = e.Victims[0]
			default:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(4)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			if e.Player.Role == Sheriff {
				assert.Equal(false, e.Success, "Mafia and maniac are supposed to be in different sides")
			} else {
				assert.Equal(true, e.Success, "all actions are supposed to be successful")
			}
		}).Times(4)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(0, len(e.Died), "doctor should have saved himself")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(4, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 2, "d")
	handle(g, 3, "b")
	handle(g, 4, "b")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(2, e.Candidate, "unexpected candidate won")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Doctor:
				victim = "f"
			case Sheriff:
				assert.Equal(2, len(e.Victims), "expected another cnt of sheriff's options")
				victim = "c d"
			case Maniac:
				assert.Equal(2, len(e.Victims), "unexpected number of maniac potential victims")
				victim = "f"
			}
			handle(g, e.Player.User, victim)
		}).Times(3)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			if e.Player.Role == Sheriff {
				assert.Equal(false, e.Success, "Doctor and maniac are supposed to be in different sides")
			} else {
				assert.Equal(true, e.Success, "all actions are supposed to be successful")
			}
		}).Times(3)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(0, len(e.Died), "doctor should have saved sheriff")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(3, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "b")

	time.Sleep(100 * time.Millisecond)

	handle(g, 3, "d")
	handle(g, 4, "c")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(4, e.Candidate, "unexpected candidate")
		}).Once()
	mockEOutput.EXPECT().HandleWin(mock.Anything).
		Run(func(e WinEvent) {
			assert.Equal(ManiacSide, e.Side, "unexpected side won")
			assert.Equal(1, len(e.Winners), "unexpected game winners count")
		}).Once()

	handle(g, 6, "d")

	time.Sleep(100 * time.Millisecond)
}

func Test_scenario3(t *testing.T) {
	assert := assert.New(t)
	mockEOutput := NewMockEventOutput(t)

	nickToUser := map[string]int64{
		"a": 1,
		"b": 2,
		"c": 3,
		"d": 4,
		"e": 5,
		"f": 6,
	}
	players := []Player{
		{1, Mafia},
		{2, Maniac},
		{3, Guesser},
		{4, Doctor},
		{5, Peaceful},
		{6, Sheriff},
	}

	g := NewGame(mockEOutput, 1, 1, rolesFromPlayers(players), func(g *Game) {})
	for nick, user := range nickToUser {
		assert.Nil(g.AddMember(user, nick), "expected no error when inserting players")
	}

	mockEOutput.EXPECT().HandleFirstDay(mock.Anything).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(6, len(candidates), "candidates for voting number")
			}
		}).Once()
	g.Start(players)

	handle(g, 1, "c")
	handle(g, 2, "a")
	handle(g, 3, "skip")
	handle(g, 4, "a")
	handle(g, 5, "b")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(0, e.Candidate, "expected draw")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Mafia:
				victim = "e"
			case Doctor:
				victim = "d"
			case Sheriff:
				victim = "a c"
			case Guesser:
				assert.Equal(26, len(e.Victims), "expected another cnt of guesser's options")
				for nick, user := range nickToUser {
					if user != e.Player.User {
						assert.Equal(5, containsIn(e.Victims, nick), "expected another guesser's options")
					}
				}
				victim = e.Victims[0]
			default:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(6)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			if e.Player.Role == Sheriff {
				assert.Equal(false, e.Success, "mafia and guesser are supposed to be in different sides")
			} else {
				assert.Equal(true, e.Success, "all actions are supposed to be successful")
			}
		}).Times(6)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(1, len(e.Died), "unexpected number of people dead")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(5, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 1, "c")
	handle(g, 2, "a")
	handle(g, 3, "skip")
	handle(g, 4, "a")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(0, e.Candidate, "expected draw")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Mafia:
				victim = "f"
			case Doctor:
				victim = "f"
			case Sheriff:
				victim = "b c"
			case Guesser:
				assert.Equal(21, len(e.Victims), "expected another cnt of guesser's options")
				victim = "a мафия"
			default:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(5)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			if e.Player.Role == Sheriff {
				assert.Equal(false, e.Success, "maniac and guesser are supposed to be in different sides")
			} else {
				assert.Equal(true, e.Success, "all actions are supposed to be successful")
			}
		}).Times(5)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(1, len(e.Died), "unexpected number of people dead")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(4, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 2, "c")
	handle(g, 3, "skip")
	handle(g, 4, "skip")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(0, e.Candidate, "expected draw")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Maniac:
				victim = "d"
			case Doctor:
				victim = "b"
			case Sheriff:
				victim = "c d"
			case Guesser:
				victim = e.Victims[0]
			}
			handle(g, e.Player.User, victim)
		}).Times(4)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			if e.Player.Role == Sheriff {
				assert.Equal(false, e.Success, "doctor and guesser are supposed to be in different sides")
			} else {
				assert.Equal(true, e.Success, "all actions are supposed to be successful")
			}
		}).Times(4)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(1, len(e.Died), "unexpected number of people dead")
		}).Once()
	mockEOutput.EXPECT().HandleVotingStarted(mock.Anything).
		Run(func(e VotingStartedEvent) {
			for _, candidates := range e.UserToCandidates {
				assert.Equal(3, len(candidates), "candidates for voting number")
			}
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)

	handle(g, 2, "c")
	handle(g, 3, "skip")

	mockEOutput.EXPECT().HandleVotingEnded(mock.Anything).
		Run(func(e VotingEndedEvent) {
			assert.EqualValues(0, e.Candidate, "expected draw")
		}).Once()
	mockEOutput.EXPECT().HandleNightStarted(mock.Anything).Once()
	mockEOutput.EXPECT().HandleNightAct(mock.Anything).
		Run(func(e NightActEvent) {
			var victim string
			switch e.Player.Role {
			case Maniac:
				victim = "c"
			case Sheriff:
				victim = "b c"
			case Guesser:
				assert.Equal(11, len(e.Victims), "expected another cnt of guesser's options")
				victim = "b маньяк"
			}
			handle(g, e.Player.User, victim)
		}).Times(3)
	mockEOutput.EXPECT().HandleActEnded(mock.Anything).
		Run(func(e ActEndedEvent) {
			if e.Player.Role == Sheriff {
				assert.Equal(false, e.Success, "doctor and guesser are supposed to be in different sides")
			} else {
				assert.Equal(true, e.Success, "all actions are supposed to be successful")
			}
		}).Times(3)
	mockEOutput.EXPECT().HandleNightEnded(mock.Anything).
		Run(func(e NightEndedEvent) {
			assert.Equal(1, len(e.Died), "unexpected number of people dead")
		}).Once()

	mockEOutput.EXPECT().HandleWin(mock.Anything).
		Run(func(e WinEvent) {
			assert.Equal(ManiacSide, e.Side, "unexpected side won")
			assert.Equal(1, len(e.Winners), "unexpected game winners count")
		}).Once()

	handle(g, 6, "skip")

	time.Sleep(100 * time.Millisecond)
}
