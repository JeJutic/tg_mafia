package game

import "errors"

type Role int

const (
	_ Role = iota
	Mafia
	Peaceful
	Doctor
	Witness
)

type Side int

const (
	_ Side = iota
	MafiaSide
	PeacefulSide
)

var roleToSide = map[Role]Side{ //mb not best that map isn't const
	Mafia:    MafiaSide,
	Peaceful: PeacefulSide,
	Doctor:   PeacefulSide,
	Witness:  PeacefulSide,
}

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
	default:
		return nil
	}
}
