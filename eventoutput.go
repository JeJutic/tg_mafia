package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jejutic/tg_mafia/pkg"
)

const (
	hider = ".\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n"
)

var roleToName = map[game.Role]string{
	game.Mafia:    "мафия",
	game.Peaceful: "мирный",
	game.Doctor:   "врач",
	game.Witness:  "свидетельница",
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

func newMessageWithoutKeyboard(user int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(user, text)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	return msg
}

func (s *server) sendMessage(msg tgbotapi.MessageConfig) { //TODO: panicing if haven't sent?
	for i := 0; i < 2; i++ {
		if _, err := s.bot.Send(msg); err == nil {
			return
		} else {
			log.Println("Unable to send from ", i, " trials: ", err)
		}
	}
}

func (s *server) sendAll(users []int64, text string) {
	for _, user := range users {
		s.sendMessage(newMessageWithoutKeyboard(user, text))
	}
}

func (s *server) HandleFirstDay(e game.FirstDayEvent) {
	for _, player := range e.Players {
		message := "Ваша роль: " + roleToName[player.Role] + "\n"

		var sidemates []string
		if player.Role == game.Mafia {
			for _, p := range e.Players {
				if p.User != player.User && p.Role == game.Mafia {
					sidemates = append(sidemates, e.UserToNick[p.User])
				}
			}
		}
		if len(sidemates) != 0 {
			message += "Ваши напарники: "
			for i, mate := range sidemates {
				message += mate
				if i != len(sidemates)-1 {
					message += ", "
				}
			}
			message += "\n"
		}

		message += hider
		s.sendMessage(newMessageWithoutKeyboard(player.User, message))
	}
}

func (s *server) HandleVotingStarted(e game.VotingStartedEvent) {
	for user, candidates := range e.UserToCandidates {
		msg := tgbotapi.NewMessage(user, "Голосуйте")

		var keyboard []tgbotapi.KeyboardButton
		for _, c := range candidates {
			keyboard = append(keyboard, tgbotapi.NewKeyboardButton(c))
		}
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)

		s.sendMessage(msg)
	}
}

func (s *server) HandleUnableToVote(e game.UnableToVoteEvent) {
	s.sendMessage(tgbotapi.NewMessage(e.User, "Вы не можете за него проголосовать"))
}

func (s *server) HandleAlreadyVoted(e game.AlreadyVotedEvent) {
	s.sendMessage(tgbotapi.NewMessage(e.User, "Вы уже проголосовали"))
}

func (s *server) HandleVotingEnded(e game.VotingEndedEvent) {
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

	s.sendAll(e.Users, message)
}

func (s *server) HandleNightStarted(e game.NightStartedEvent) {
	s.sendAll(e.Users, "Город засыпает. Просыпается "+e.FirstToWake)
}

func (s *server) HandleNightAct(e game.NightActEvent) {
	var choose string
	switch e.Player.Role {
	case game.Mafia:
		choose = "Выбирайте жертву"
	case game.Peaceful:
		choose = "Просто нажмите на кнопку"
	case game.Doctor:
		choose = "Выберите, кого лечить"
	case game.Witness:
		choose = "Выберите, кому вы доверяете - его не смогут ошибочно выгнать"
	default:
		choose = "Выбирайте"
	}

	msg := tgbotapi.NewMessage(e.Player.User, choose)

	var keyboard []tgbotapi.KeyboardButton
	for _, v := range e.Victims {
		keyboard = append(keyboard, tgbotapi.NewKeyboardButton(v))
	}
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)

	s.sendMessage(msg)
}

func (s *server) HandleUnexpectedActTrial(e game.UnexpectedActTrialEvent) {
	s.sendMessage(tgbotapi.NewMessage(e.User, "Какого фига, ты спать должен"))
}

func (s *server) HandleUnsupportedAct(e game.UnsupportedActEvent) {
	s.sendMessage(tgbotapi.NewMessage(e.User, "Вы не можете его выбрать"))
}

func (s *server) HandleActEnded(e game.ActEndedEvent) {
	var message string
	if e.Success {
		message = "Цель выбрана"
	} else {
		message = "Промах! Другая мафия выбрала другого"
	}
	message += "\n\n"

	if e.Next == "" {
		message += "Просыпается город"
	} else {
		message += "Просыпается " + e.Next
	}
	s.sendMessage(newMessageWithoutKeyboard(e.Player.User, message))
}

func (s *server) HandleNightEnded(e game.NightEndedEvent) {

	message := hider + "Этой ночью\n"
	if e.Killed == "" {
		message += "Никого не убили\n"
	} else {
		message += "Убили " + e.Killed + "\n"
	}
	s.sendAll(e.Users, message)
}

func (s *server) HandleWin(e game.WinEvent) {
	message := "Победили " + roleToName[e.Role] + "\n\n"

	for _, nick := range e.Winners {
		message += nick + "\n"
	}

	s.sendAll(e.Users, message)
}

// cleaning references for garbage collection
func (s *server) HandleNotifyStopGame(e game.NotifyStopGameEvent) {
	s.sendAll(e.Users, "Игра прервана")
}
