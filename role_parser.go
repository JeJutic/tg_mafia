package main

const (
	MAFIA_ROLE    = iota
	PEACEFUL_ROLE = iota
	DOCTOR_ROLE   = iota
	WITNESS_ROLE  = iota
)

var roleToName = map[int]string{
	MAFIA_ROLE:    "мафия",
	PEACEFUL_ROLE: "мирный",
	DOCTOR_ROLE:   "врач",
	WITNESS_ROLE:  "свидетельница",
}

func areValidRoles(roles []int) string {
	return ""
}

func parseRoles(tokens []string) ([]int, string) {
	roles := make([]int, len(tokens))
	for i, token := range tokens {
		if token == "мафия" || token == "маф" {
			roles[i] = MAFIA_ROLE
		} else if token == "мирный" || token == "мир" {
			roles[i] = PEACEFUL_ROLE
		} else if token == "врач" || token == "доктор" {
			roles[i] = DOCTOR_ROLE
		} else if token == "свидетельница" || token == "свид" {
			roles[i] = WITNESS_ROLE
		} else {
			err := "Неизвестный токен роли: " + token
			return roles, err
		}
	}
	return roles, areValidRoles(roles)
}

func outputRoles(roles []int) string {
	output := ""
	for i, role := range roles {
		output += roleToName[role]
		if i != len(roles)-1 {
			output += ", "
		}
	}
	return output
}
