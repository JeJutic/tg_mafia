package game

type EventOutput interface {
	HandleFirstDay(FirstDayEvent)

	HandleVotingStarted(VotingStartedEvent)
	HandleUnableToVote(UnableToVoteEvent)
	HandleAlreadyVoted(AlreadyVotedEvent)
	HandleVotingEnded(VotingEndedEvent)

	HandleNightStarted(NightStartedEvent)
	HandleNightAct(NightActEvent)
	HandleUnexpectedActTrial(UnexpectedActTrialEvent)
	HandleUnsupportedAct(UnsupportedActEvent)
	HandleActEnded(ActEndedEvent)
	HandleNightEnded(NightEndedEvent)

	HandleWin(WinEvent)
	HandleNotifyStopGame(NotifyStopGameEvent) //unnecessary to call when stopping
}

type FirstDayEvent struct {
	UserToNick map[int64]string
	Players    []Player
}

type VotingStartedEvent struct { //should refactor to make 1 mess/sec by combining with NightEnded
	UserToCandidates map[int64][]string
}

type UnableToVoteEvent struct {
	User int64
}

type AlreadyVotedEvent struct {
	User int64
}

type VotingEndedEvent struct {
	Users       []int64
	UserToNick  map[int64]string
	UserToVoted map[int64]int64
	Candidate   int64
	Witness     bool
}

type NightStartedEvent struct {
	Users       []int64
	FirstToWake string
}

type NightActEvent struct {
	Player  Player //player's role just to name acting right way
	Victims []string
}

type UnexpectedActTrialEvent struct {
	User int64
}

type UnsupportedActEvent struct {
	User int64
}

type ActEndedEvent struct {
	Player  Player
	Success bool
	Next    string
}

type NightEndedEvent struct { //notifying what happened during night
	Users  []int64
	Killed string
}

type WinEvent struct {
	Users   []int64
	Role    Role
	Winners []string
}

type NotifyStopGameEvent struct {
	Users []int64
}
