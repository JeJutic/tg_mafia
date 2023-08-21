package gameServer

import (
	"log"

	game "github.com/jejutic/tg_mafia/pkg/game"
)

const (
	hider = ".\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n"
)

var roleToName = map[game.Role]string{
	game.Mafia:    "мафия",
	game.Peaceful: "мирный",
	game.Doctor:   "врач",
	game.Witness:  "свидетельница",
	game.Sheriff:  "комиссар",
	game.Maniac:   "маньяк",
	game.Guesser:  "разгадыватель",
}

var sideToName = map[game.Side]string{
	game.MafiaSide:    "лагерь мафии",
	game.PeacefulSide: "лагерь мирных",
	game.ManiacSide:   "лагерь маньяка",
	game.GuesserSide:  "лагерь разгадывателя",
}

func rolesToString(roles []game.Role) string {
	output := ""
	for i, role := range roles {
		output += roleToName[role]
		if i != len(roles)-1 {
			output += ", "
		}
	}
	return output
}

func (ms mafiaServer[T]) HandleFirstDay(e game.FirstDayEvent) {
	for _, player := range e.Players {
		text := "Ваша роль: " + roleToName[player.Role] + "\n"

		var sidemates []string
		if player.Role == game.Mafia {
			for _, p := range e.Players {
				if p.User != player.User && p.Role == game.Mafia {
					sidemates = append(sidemates, e.UserToNick[p.User])
				}
			}
		}
		if len(sidemates) != 0 {
			text += "Ваши напарники: "
			for i, mate := range sidemates {
				text += mate
				if i != len(sidemates)-1 {
					text += ", "
				}
			}
			text += "\n"
		}

		text += hider
		ms.SendMessage(ServerMessage{
			User: player.User,
			Text: text,
		})
	}
}

func (ms mafiaServer[T]) HandleVotingStarted(e game.VotingStartedEvent) {
	for user, candidates := range e.UserToCandidates {
		ms.SendMessage(ServerMessage{
			User:    user,
			Text:    "Голосуйте",
			Options: candidates,
		})
	}
}

func (ms mafiaServer[T]) HandleUnableToVote(e game.UnableToVoteEvent) {
	ms.SendMessage(newMessage(e.User, "Вы не можете за него проголосовать", false))
}

func (ms mafiaServer[T]) HandleAlreadyVoted(e game.AlreadyVotedEvent) {
	ms.SendMessage(newMessage(e.User, "Вы уже проголосовали", true))
}

func (ms mafiaServer[T]) HandleVotingEnded(e game.VotingEndedEvent) {
	var message string

	for user, voted := range e.UserToVoted {
		message += e.UserToNick[user]
		if voted != -1 {
			message += " проголосовал за " + e.UserToNick[voted] + "\n"
		} else {
			message += " скипнул\n"
		}
	}
	message += "\n"

	switch {
	case e.Candidate == 0:
		message += "Никто не был исключен. "
	case e.Witness:
		message += e.UserToNick[e.Candidate] + "был защищен свидетельницей"
	default:
		message += "Исключили " + e.UserToNick[e.Candidate]
	}

	sendAll[T](ms, e.Users, message, false)
}

func (ms mafiaServer[T]) HandleNightStarted(e game.NightStartedEvent) {
	sendAll[T](ms, e.Users, "Город засыпает. Просыпается "+e.FirstToWake, false)
}

func (ms mafiaServer[T]) HandleNightAct(e game.NightActEvent) {
	var text string
	switch e.Player.Role {
	case game.Mafia:
		text = "Выбирайте жертву"
	case game.Peaceful:
		text = "Просто нажмите на кнопку"
	case game.Doctor:
		text = "Выберите, кого лечить"
	case game.Witness:
		text = "Выберите, кому вы доверяете - его не смогут ошибочно выгнать"
	case game.Sheriff:
		text = "Выберите двух игроков, чтобы проверить, из разных ли они лагерей"
	case game.Maniac:
		if e.MafiaAlive {
			text = "Пока мафия жива, вы не можете никого выбрать"
		} else {
			text = "Выбирайте жертву"
		}
	case game.Guesser:
		text = `
		Вы можете попытаться угадать роль одного из игроков. 
		Если вы окажетесь верны, он умрет, если его не спасут или не убьют вас. 
		Если вы угадаете неверно, вы умрете, если вас не спасут
		`
	default:
		text = "Выбирайте"
	}

	ms.SendMessage(ServerMessage{
		User:    e.Player.User,
		Text:    text,
		Options: e.Victims,
	})
}

func (ms mafiaServer[T]) HandleUnexpectedActTrial(e game.UnexpectedActTrialEvent) {
	ms.SendMessage(newMessage(e.User, "Какого фига, ты спать должен", false))
}

func (ms mafiaServer[T]) HandleUnsupportedAct(e game.UnsupportedActEvent) {
	ms.SendMessage(newMessage(e.User, "Вы не можете его выбрать", false))
}

func (ms mafiaServer[T]) HandleActEnded(e game.ActEndedEvent) {
	var message string
	if e.Success {
		if e.Player.Role == game.Sheriff {
			message = "Они в одном лагере"
		} else {
			message = "Цель выбрана"
		}
	} else {
		if e.Player.Role == game.Sheriff {
			message = "Они в разных лагерях"
		} else if e.Player.Role == game.Mafia {
			message = "Промах! Другая мафия выбрала другого"
		} else {
			message = "Случился баг"
			log.Println("unexpected unsuccessful action of ", e.Player.Role)
		}
	}
	message += "\n\n"

	if e.Next == "" {
		message += "Просыпается город"
	} else {
		message += "Просыпается " + e.Next
	}
	ms.SendMessage(newMessage(e.Player.User, message, true))
}

func (ms mafiaServer[T]) HandleNightEnded(e game.NightEndedEvent) {

	message := hider + "Этой ночью\n"
	if len(e.Died) == 0 {
		message += "Никого не убили\n"
	} else {
		message += "Убили "
		for i, died := range e.Died {
			message += died
			if i != len(e.Died)-1 {
				message += ", "
			}
		}
		message += "\n"
	}
	sendAll[T](ms, e.Users, message, false)
}

func (ms mafiaServer[T]) HandleWin(e game.WinEvent) {
	message := "Победил " + sideToName[e.Side] + "\n\n"

	for _, nick := range e.Winners {
		message += nick + "\n"
	}

	sendAll[T](ms, e.Users, message, true)
}

// cleaning references for garbage collection
func (ms mafiaServer[T]) HandleNotifyStopGame(e game.NotifyStopGameEvent) {
	sendAll[T](ms, e.Users, "Игра прервана", true)
}
