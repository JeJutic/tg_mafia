package game

import "errors"

// Role represents a role in mafia game
type Role int

const (
	_ Role = iota
	Mafia
	Peaceful
	Doctor
	Witness
	Sheriff
	Maniac
	Guesser
)

// Side represents a side in mafia game. Only one can win in a game
type Side int

const (
	_ Side = iota
	MafiaSide
	PeacefulSide
	ManiacSide
	GuesserSide
)

var roleToName = map[Role]string{
	Mafia:    "мафия",
	Peaceful: "мирный",
	Doctor:   "врач",
	Witness:  "свидетельница",
	Sheriff:  "комиссар",
	Maniac:   "маньяк",
	Guesser:  "разгадыватель",
}

var roleToSide = map[Role]Side{ //mb not best that map isn't const
	Mafia:    MafiaSide,
	Peaceful: PeacefulSide,
	Doctor:   PeacefulSide,
	Witness:  PeacefulSide,
	Sheriff:  PeacefulSide,
	Maniac:   ManiacSide,
	Guesser:  GuesserSide,
}

// ValidRoles returns nil error iff list of role-cards roles is valid for mafia game
func ValidRoles(roles []Role) error {
	roleToCnt := make(map[Role]int)
	for _, role := range roles {
		roleToCnt[role]++
	}

	switch {
	case roleToCnt[Doctor] > 1:
		return errors.New("не может быть более одного доктора")
	case roleToCnt[Witness] > 1:
		return errors.New("не может быть более одной свидетельницы")
	case roleToCnt[Maniac] > 1:
		return errors.New("не может быть более одного маньяка")
	case roleToCnt[Guesser] > 1:
		return errors.New("не может быть более одного разгадывателя")
	default:
		return nil
	}
}
