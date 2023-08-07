package game

import (
	// "os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newVoting(t *testing.T) {
	ga := &gameActive{}
	ga.newVoting(false)

	if ga.voting.userToVoted == nil {
		t.Error("Voting map field isn't initialized properly")
	}
}

func Test_everyoneVoted(t *testing.T) {
	v := voting{
		gActive: &gameActive{
			pQueue: []Player{
				{1, Mafia},
				{2, Peaceful},
				{3, Peaceful},
			},
		},
		userToVoted: map[int64]int64{
			1: 2,
			3: -1,
		},
	}
	if v.everyoneVoted() {
		t.Error("Expected voting not ended but everyoneVoted() is true")
	}
	v.userToVoted[2] = 1
	if !v.everyoneVoted() {
		t.Error("Expected voting ended but everyoneVoted() is false")
	}
}

func Test_skippedCnt(t *testing.T) {
	v := voting{
		userToVoted: map[int64]int64{
			1: 2,
			2: -1,
			3: -1,
			4: 3,
		},
	}

	assert.Equal(t, 2, v.skippedCnt(), "Count of skipped")
}

func Test_bestCandidate(t *testing.T) {
	assert := assert.New(t)

	v := voting{
		userToVoted: map[int64]int64{
			1: 2,
			2: -1,
			3: -1,
			4: 3,
		},
	}
	assert.EqualValues(0, v.bestCandidate(), "Expected no best candidate")

	v.userToVoted[1] = 3
	assert.EqualValues(0, v.bestCandidate(), "Expected no best candidate")

	v.userToVoted[2] = 3
	assert.EqualValues(3, v.bestCandidate(), "Expected another best candidate")
}
