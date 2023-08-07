package game

// EventOutput is used as view model for mafia game. Outputs events
// happening during mafia game
type EventOutput interface {
	HandleFirstDay(FirstDayEvent)	// outputs FirstDayEvent

	HandleVotingStarted(VotingStartedEvent)	// outputs VotingStartedEvent
	HandleUnableToVote(UnableToVoteEvent)	// outputs UnableToVoteEvent
	HandleAlreadyVoted(AlreadyVotedEvent)	// outputs AlreadyVotedEvent
	HandleVotingEnded(VotingEndedEvent)	// outputs VotingEndedEvent

	HandleNightStarted(NightStartedEvent)	// outputs NightStartedEvent
	HandleNightAct(NightActEvent)	// outputs NightActEvent
	HandleUnexpectedActTrial(UnexpectedActTrialEvent)	// outputs UnexpectedActTrialEvent
	HandleUnsupportedAct(UnsupportedActEvent)	// outputs UnsupportedActEvent
	HandleActEnded(ActEndedEvent)	// outputs ActEndedEvent
	HandleNightEnded(NightEndedEvent)	// outputs NightEndedEvent

	HandleWin(WinEvent)	// outputs WinEvent
	HandleNotifyStopGame(NotifyStopGameEvent)	// outputs NotifyStopGameEvent
}

// FirstDayEvent represents starting of the first day in the game
type FirstDayEvent struct {
	UserToNick map[int64]string	// mapping from player's id to nick in the game
	Players    []Player	// slice of players in the game
}

// VotingStartedEvent represents starting of the voting and holds the information
// which candidates are availible for players for voting
type VotingStartedEvent struct {
	UserToCandidates map[int64][]string	// mapping from player's id to list of nicks of his candidates for voting
}
//TODO: should refactor to make 1 mess/sec by combining with NightEnded

// UnableToVoteEvent represents a situation where player tries to vote for not availible
// for him candidate
type UnableToVoteEvent struct {
	User int64	// id of the player
}

// AlreadyVotedEvent represents a situation where player tries to vote more than once
type AlreadyVotedEvent struct {
	User int64	// id of the player
}

// VotingEndedEvent represents result of the voting
type VotingEndedEvent struct {
	Users       []int64	// slice of all players' (including dead) ids
	UserToNick  map[int64]string	// mapping from player's id to his nick in the game
	UserToVoted map[int64]int64	// mapping from player's id to id of player he voted for or -1 if skipped
	Candidate   int64	// id of a player who was selected in the voting or 0 in case of draw
	Witness     bool	// true if selected player had been witnessed
}

// NightStartedEvent represents start of the night
type NightStartedEvent struct {
	Users       []int64	// slice of all players' (including dead) ids
	FirstToWake string	// nick of the player who should wake up first
}

// NightActEvent represents the result of action of a player during the night
type NightActEvent struct {
	Player     Player // player acted
	Victims    []string	// list of victim selected by player
	MafiaAlive bool	// true iff mafia is alive
}

// UnexpectedActTrialEvent represents a situation where one of the players tries
// to act when he is not supposed to
type UnexpectedActTrialEvent struct {
	User int64	// id of the player
}

// UnsupportedActEvent represents a situation where player tries to select
// victims which is not supposed to select
type UnsupportedActEvent struct {
	User int64	// id of the player
}

// ActEndedEvent represents result of player's night act
type ActEndedEvent struct {
	Player  Player	// player acted
	Success bool	// if the action was successful, semantics varies for different roles
	Next    string	// nick of a player who should wake up next
}

// NightEndedEvent represents results of night ended
type NightEndedEvent struct {
	Users  []int64	// slice of all players' (including dead) ids
	Killed string	// nick of player who died this night
}

// WinEvent represents win of one of the sides
type WinEvent struct {
	Users   []int64	// slice of all players' (including dead) ids
	Side    Side	// side winned
	Winners []string	// slice of nicks of players winned
}

// NotifyStopGameEvent represents request for notifying all players about stopping the game
type NotifyStopGameEvent struct {
	Users []int64	// slice of all players' (including dead) ids
}
