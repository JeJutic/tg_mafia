package game

const (
	skip = "skip"
)

type Voting struct {
	gActive     *GameActive
	firstVoting bool
	userToVoted map[int64]int64
}

func (ga *GameActive) NewVoting(firstVoting bool) {
	ga.voting = &Voting{
		ga,
		firstVoting,
		make(map[int64]int64),
	}
}

func (v *Voting) memberCanVote(member int64) []string {
	list := make([]string, 1)
	list[0] = skip
	for _, player := range v.gActive.pQueue {
		if player.User != member {
			list = append(list, v.gActive.game.UserToNick[player.User])
		}
	}
	return list
}

func (v *Voting) isEveryoneVoted() bool {
	for _, player := range v.gActive.pQueue {
		if _, exists := v.userToVoted[player.User]; !exists {
			return false
		}
	}
	return true
}

func (v *Voting) handleVote(user int64, vote string) {
	if _, exists := v.userToVoted[user]; !exists {
		if contains(v.memberCanVote(user), vote) {
			if vote == skip {
				v.userToVoted[user] = -1
			} else {
				v.userToVoted[user] = v.gActive.game.NickToUser[vote]
			}

			if v.isEveryoneVoted() {
				v.gActive.votingConclusion()
			}
		} else {
			v.gActive.game.EOutput.HandleUnableToVote(UnableToVoteEvent{
				user,
			})
		}
	} else {
		v.gActive.game.EOutput.HandleAlreadyVoted(AlreadyVotedEvent{
			user,
		})
	}
}

func (v *Voting) skippedCnt() int {
	skipped := 0
	for _, voted := range v.userToVoted {
		if voted == -1 {
			skipped++
		}
	}
	return skipped
}

func (v *Voting) bestCandidate() int64 {
	userToVotes := make(map[int64]int)
	for _, voted := range v.userToVoted {
		if voted != -1 {
			userToVotes[voted]++
		}
	}

	bestCandidate, cnt := int64(0), 0
	for user, votes := range userToVotes {
		if votes > cnt {
			bestCandidate, cnt = user, votes
		} else if votes == cnt {
			bestCandidate = 0
		}
	}

	if v.skippedCnt() >= cnt {
		bestCandidate = 0
	}
	return bestCandidate
}
