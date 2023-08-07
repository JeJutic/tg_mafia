package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	game "github.com/jejutic/tg_mafia/pkg"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const (
	hider = ".\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n.\n"
)

var bundle *i18n.Bundle

func init() {
	bundle = i18n.NewBundle(language.English)
}

var roleToName = map[game.Role]string{
	game.Mafia:    "мафия",
	game.Peaceful: "мирный",
	game.Doctor:   "врач",
	game.Witness:  "свидетельница",
	game.Sheriff:  "комиссар",
	game.Maniac:   "маньяк",
}

var sideToName = map[game.Side]string{
	game.MafiaSide:    "лагерь мафии",
	game.PeacefulSide: "лагерь мирных",
	game.ManiacSide:   "лагерь маньяка",
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

func (s *server) localizerFromUser(user int64) *i18n.Localizer {
	member, err := s.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: user,
			UserID: user,
		},
	})
	if err == nil {
		log.Println("couldn't get user's language code: ", err)
	}
	if err != nil && member.User.LanguageCode == "ru" {
		return i18n.NewLocalizer(bundle, language.Russian.String(), language.English.String())
	} else {
		return i18n.NewLocalizer(bundle, language.Russian.String(), language.English.String())
	}
}

func (s *server) messageIDToText(user int64, messageID string, a ...any) string {
	config := i18n.LocalizeConfig{
		MessageID: messageID,
	}
	text, err := s.localizerFromUser(user).Localize(&config)
	if err != nil {
		log.Println("couldn't find localization for ", messageID)
		text = "internal server error: unable to find messageID"
	} else {
		text = fmt.Sprintf(text, a)
	}
	return text
}

// func newMessage(user int64, text string) tgbotapi.MessageConfig {
// 	msg := tgbotapi.NewMessage(user, text)
// 	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
// 	return msg
// }

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
		msg := newMessageWithoutKeyboard(user, s.messageIDToText(user, "vote"))

		var keyboard [][]tgbotapi.KeyboardButton
		for _, c := range candidates {
			keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(c)))
		}
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)

		s.sendMessage(msg)
	}
}

func (s *server) HandleUnableToVote(e game.UnableToVoteEvent) {
	s.sendMessage(newMessageWithoutKeyboard(e.User, s.messageIDToText(e.User, "unable_to_vote")))
}

func (s *server) HandleAlreadyVoted(e game.AlreadyVotedEvent) {
	s.sendMessage(newMessageWithoutKeyboard(e.User, s.messageIDToText(e.User, "already_voted")))
}

func (s *server) HandleVotingEnded(e game.VotingEndedEvent) {
	var message string

	for user, voted := range e.UserToVoted {
		message += e.UserToNick[user]
		if voted != -1 {
			message += " проголосовал за " + s.messageIDToText(e.User, "already_voted") + e.UserToNick[voted] + "\n"
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
	case game.Sheriff:
		choose = "Выберите двух игроков, чтобы проверить, из разных ли они лагерей"
	case game.Maniac:
		if e.MafiaAlive {
			choose = "Пока мафия жива, вы не можете никого выбрать"
		} else {
			choose = "Выбирайте жертву"
		}
	default:
		choose = "Выбирайте"
	}

	msg := tgbotapi.NewMessage(e.Player.User, choose)

	var keyboard [][]tgbotapi.KeyboardButton
	for _, v := range e.Victims {
		keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(v)))
	}
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)

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
	message := "Победил " + sideToName[e.Side] + "\n\n"

	for _, nick := range e.Winners {
		message += nick + "\n"
	}

	s.sendAll(e.Users, message)
}

// cleaning references for garbage collection
func (s *server) HandleNotifyStopGame(e game.NotifyStopGameEvent) {
	s.sendAll(e.Users, "Игра прервана")
}
