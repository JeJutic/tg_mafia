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
)

// Side represents a side in mafia game. Only one can win in a game
type Side int

const (
	_ Side = iota
	MafiaSide
	PeacefulSide
	ManiacSide
)

var roleToSide = map[Role]Side{ //mb not best that map isn't const
	Mafia:    MafiaSide,
	Peaceful: PeacefulSide,
	Doctor:   PeacefulSide,
	Witness:  PeacefulSide,
	Sheriff:  PeacefulSide,
	Maniac:   ManiacSide,
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
	default:
		return nil
	}
}
