package game

const (
	skip = "skip"
)

type voting struct {
	gActive     *gameActive
	firstVoting bool
	userToVoted map[int64]int64
}

func (ga *gameActive) initVoting(firstVoting bool) {
	ga.voting = &voting{
		ga,
		firstVoting,
		make(map[int64]int64),
	}
}

func (v *voting) userCanVote(user int64) []string {
	list := make([]string, 1)
	list[0] = skip
	for _, player := range v.gActive.pQueue {
		if player.User != user {
			list = append(list, v.gActive.UserToNick[player.User])
		}
	}
	return list
}

func (v *voting) everyoneVoted() bool {
	for _, player := range v.gActive.pQueue {
		if _, exists := v.userToVoted[player.User]; !exists {
			return false
		}
	}
	return true
}

func (v *voting) handleVote(user int64, vote string) {
	if _, exists := v.userToVoted[user]; !exists {
		if contains(v.userCanVote(user), vote) {
			if vote == skip {
				v.userToVoted[user] = -1
			} else {
				v.userToVoted[user] = v.gActive.NickToUser[vote]
			}

			if v.everyoneVoted() {
				v.gActive.votingConclusion()
			}
		} else {
			v.gActive.EOutput.HandleUnableToVote(UnableToVoteEvent{
				user,
			})
		}
	} else {
		v.gActive.EOutput.HandleAlreadyVoted(AlreadyVotedEvent{
			user,
		})
	}
}

func (v *voting) skippedCnt() int {
	skipped := 0
	for _, voted := range v.userToVoted {
		if voted == -1 {
			skipped++
		}
	}
	return skipped
}

func (v *voting) bestCandidate() int64 {
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
