package game

import "errors"

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
