package main

import (
	"log"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Game struct {
	serv       *server
	code       int
	creator    int64
	nickToUser map[string]int64
	userToNick map[int64]string
	roles      []int
	gameActive *GameActive
}

func NewGame(serv *server, code int, creator int64, roles []int) *Game {
	return &Game{
		serv,
		code,
		creator,
		make(map[string]int64),
		make(map[int64]string),
		roles,
		nil,
	}
}

func (game *Game) addMember(member int64, nick string) {
	game.nickToUser[nick] = member // TODO: check if not busy
	game.userToNick[member] = nick

	if len(game.nickToUser) == len(game.roles) {
		game.NewGameActive(game.RandomPlayerQueue())

		message := "\n\n\nПервый день. Список участников:\n"
		for nick, _ := range game.nickToUser {
			message += nick + "\n"
		}
		for _, user := range game.nickToUser {
			role := game.gameActive.getPlayerFromUser(user).role
			message := "Ваша роль: " + roleToName[role] + "\n"
			if role == MAFIA_ROLE {
				for _, player := range game.gameActive.playerQueue {
					if player.role == MAFIA_ROLE && player.user != user {
						message += "Ваш напарник: " + game.userToNick[player.user] + "\n"
					}
				}
			}
			message += hider
			game.serv.sendMessage(tgbotapi.NewMessage(user, message))
		}
		game.sendAll(message)
		game.gameActive.startDay()
	}
}

func (game *Game) sendAll(message string) {
	for _, user := range game.nickToUser {
		msg := tgbotapi.NewMessage(user, message)
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		game.serv.sendMessage(msg)
	}
}

type player struct {
	user int64
	role int
}

type GameActive struct {
	game        *Game
	firstDay    bool
	playerQueue []player //will be deleted from map if dead
	witnessed   int64
	saved       int64
	userToVoted map[int64]int64
	night       *Night
}

func (game *Game) isMafiaAlive() bool {
	mafiaAlive := false
	for role := range game.roles {
		if role == MAFIA_ROLE {
			mafiaAlive = true
		}
	}
	return mafiaAlive
}

func (game *Game) NewGameActive(playerQueue []player) {
	game.gameActive = &GameActive{
		game,
		true,
		playerQueue,
		0,
		0,
		make(map[int64]int64),
		nil,
	}
}

func (game *Game) RandomPlayerQueue() []player {
	rand.Shuffle(len(game.roles), func(i, j int) {
		game.roles[i], game.roles[j] = game.roles[j], game.roles[i]
	})
	playerQueue := make([]player, 0, len(game.nickToUser))
	for _, user := range game.nickToUser {
		playerQueue = append(playerQueue, player{user, game.roles[len(playerQueue)]})
	}
	rand.Shuffle(len(playerQueue), func(i, j int) {
		playerQueue[i], playerQueue[j] = playerQueue[j], playerQueue[i]
	})
	return playerQueue
}

type Night struct {
	gameActive *GameActive
	offset     int
	shot       int64
	saved      int64
	witnessed  int64
}

func (gameActive *GameActive) NewNight() {
	gameActive.night = &Night{
		gameActive,
		-1,
		0,
		0,
		0,
	}
}

func (night *Night) next() {
	night.offset++
	if night.offset < len(night.gameActive.playerQueue) {
		log.Println("Its time for " + night.gameActive.game.userToNick[night.gameActive.playerQueue[night.offset].user])
		night.gameActive.sendActingMessage(night.gameActive.playerQueue[night.offset])
	} else {
		message := hider + "Этой ночью\n"
		if night.shot != 0 && night.shot != night.saved {
			message += "Убили " + night.gameActive.game.userToNick[night.shot] + "\n"
			night.gameActive.removePlayer(night.shot)
		} else {
			message += "Никого не убили\n"
		}
		night.gameActive.saved = night.saved
		night.gameActive.witnessed = night.witnessed

		night.gameActive.game.sendAll(message)
		night.gameActive.startDay()
	}
}

func (gameActive *GameActive) checkForEnd() int {
	mafiaCnt := 0
	for _, player := range gameActive.playerQueue {
		if player.role == MAFIA_ROLE {
			mafiaCnt++
		}
	}
	if mafiaCnt >= len(gameActive.playerQueue)-mafiaCnt {
		return MAFIA_ROLE
	} else if mafiaCnt == 0 {
		return PEACEFUL_ROLE
	} else {
		return -1
	}
}

func (gameActive *GameActive) updatePlayerQueue() { //TODO

}

func (gameActive *GameActive) getPlayerFromUser(user int64) player {
	for _, player := range gameActive.playerQueue {
		if player.user == user {
			return player
		}
	}
	log.Fatal("Unable to find the user")
	return player{}
}

func (gameActive *GameActive) sendVotingMessage(user int64) {
	msg := tgbotapi.NewMessage(user, "Голосуйте")
	keyboard := make([]tgbotapi.KeyboardButton, 0)
	for _, option := range gameActive.memberCanVote(user) {
		keyboard = append(keyboard, tgbotapi.NewKeyboardButton(option))
	}
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
	gameActive.game.serv.sendMessage(msg)
}

func (gameActive *GameActive) sendActingMessage(player player) {
	msg := tgbotapi.NewMessage(player.user, "Выбирайте")
	keyboard := make([]tgbotapi.KeyboardButton, 0)
	for _, option := range gameActive.playerCanAct(player) {
		keyboard = append(keyboard, tgbotapi.NewKeyboardButton(option))
	}
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
	gameActive.game.serv.sendMessage(msg)
}

func remove(slice []player, s int) []player {
	return append(slice[:s], slice[s+1:]...)
}

func (gameActive *GameActive) removePlayer(user int64) {
	for i, player := range gameActive.playerQueue {
		if player.user == user {
			gameActive.playerQueue = remove(gameActive.playerQueue, i)
			return
		}
	}
	log.Fatal("nobody to remove")
}

func (gameActive *GameActive) startDay() {
	gameActive.night = nil
	gameActive.userToVoted = make(map[int64]int64)
	if roleWinned := gameActive.checkForEnd(); roleWinned != -1 {
		message := "Победили " + roleToName[roleWinned] + "\n"
		for _, player := range gameActive.playerQueue {
			if player.role == roleWinned {
				message += gameActive.game.userToNick[player.user] + "\n"
			}
		}
		gameActive.game.sendAll(message)
		gameActive.game.stopGame()
		return
	}
	for _, player := range gameActive.playerQueue {
		gameActive.sendVotingMessage(player.user)
	}
}

func (gameActive *GameActive) startNight() {
	gameActive.NewNight()
	message := "Город засыпает. Просыпается " + gameActive.game.userToNick[gameActive.playerQueue[0].user]
	gameActive.game.sendAll(message)
	gameActive.night.next()
}

func (gameActive *GameActive) skippedCnt() int {
	skipped := 0
	for _, voted := range gameActive.userToVoted {
		if voted == -1 {
			skipped++
		}
	}
	return skipped
}

func (gameActive *GameActive) votingConclusion() {
	message := ""
	userToVotes := make(map[int64]int)
	for user, voted := range gameActive.userToVoted {
		message += gameActive.game.userToNick[user]
		if voted != -1 {
			message += " проголосовал за " + gameActive.game.userToNick[voted] + "\n"
			userToVotes[voted]++
		} else {
			message += " скипнул\n"
		}
	}
	bestCandidate, cnt := int64(0), 0
	for user, votes := range userToVotes {
		if votes > cnt {
			bestCandidate, cnt = user, votes
		} else if votes == cnt {
			bestCandidate = 0
		}
	}
	if gameActive.skippedCnt() >= cnt {
		bestCandidate = 0
	}

	message += "\n"
	if bestCandidate == 0 {
		message += "Никто не был исключен. "
	} else if bestCandidate == gameActive.witnessed {
		message += gameActive.game.userToNick[bestCandidate] + "был защищен свидетельницей"
	} else {
		message += "Исключили " + gameActive.game.userToNick[bestCandidate]
		gameActive.removePlayer(bestCandidate)
	}
	gameActive.game.sendAll(message)
	//TODO: special check for end
	gameActive.startNight()
}

func (gameActive *GameActive) memberCanVote(member int64) []string {
	list := make([]string, 1)
	list[0] = "skip"
	for _, player := range gameActive.playerQueue {
		if player.user != member {
			list = append(list, gameActive.game.userToNick[player.user])
		}
	}
	return list
}

func (gameActive *GameActive) playerCanAct(member player) []string {
	list := make([]string, 0)
	if member.role == PEACEFUL_ROLE {
		list = append(list, "сделать ничего")
	}
	for _, player := range gameActive.playerQueue {
		if member.role == MAFIA_ROLE {
			if player.role != MAFIA_ROLE {
				list = append(list, gameActive.game.userToNick[player.user])
			}
		} else if member.role == DOCTOR_ROLE {
			if gameActive.saved != player.user {
				list = append(list, gameActive.game.userToNick[player.user])
			}
		} else if member.role == WITNESS_ROLE {
			if gameActive.witnessed != player.user {
				list = append(list, gameActive.game.userToNick[player.user])
			}
		}
	}
	return list
}

func contains(elems []string, target string) bool {
	for _, elem := range elems {
		if elem == target {
			return true
		}
	}
	return false

}

func (gameActive *GameActive) isEveryoneVoted() bool {
	for _, player := range gameActive.playerQueue {
		if _, exists := gameActive.userToVoted[player.user]; !exists {
			return false
		}
	}
	return true
}

func (gameActive *GameActive) handle(user int64, request string) {
	if gameActive.night == nil { //day
		if _, exists := gameActive.userToVoted[user]; !exists {
			if contains(gameActive.memberCanVote(user), request) {
				if request == "skip" {
					gameActive.userToVoted[user] = -1
				} else {
					gameActive.userToVoted[user] = gameActive.game.nickToUser[request]
				}

				if gameActive.isEveryoneVoted() {
					log.Println("Concluding")
					gameActive.votingConclusion()
				} else {
					log.Println("Not Concluding")
				}
			} else {
				gameActive.game.serv.sendMessage(tgbotapi.NewMessage(user, "Вы не можете за него проголосовать"))
			}
		} else {
			gameActive.game.serv.sendMessage(tgbotapi.NewMessage(user, "Вы уже проголосовали"))
		}
	} else {
		if player := gameActive.playerQueue[gameActive.night.offset]; player.user == user {
			if contains(gameActive.playerCanAct(player), request) {
				message := ""
				if player.role == MAFIA_ROLE {
					if gameActive.night.shot != 0 {	//TODO
						message += "Промах! Другая мафия выбрала другого\n"
					} else {
						message += "Цель выбрана\n"
						gameActive.night.shot = gameActive.game.nickToUser[request]
					}
				} else if player.role == DOCTOR_ROLE {
					message += "Цель выбрана"
					gameActive.night.saved = gameActive.game.nickToUser[request]
				} else if player.role == WITNESS_ROLE {
					message += "Цель выбрана"
					gameActive.night.witnessed = gameActive.game.nickToUser[request]
				}
				if gameActive.night.offset+1 < len(gameActive.playerQueue) {
					message += "Просыпается " + gameActive.game.userToNick[gameActive.playerQueue[gameActive.night.offset+1].user]
				} else {
					message += "Просыпается город"
				}
				gameActive.game.serv.sendMessage(tgbotapi.NewMessage(user, message))
				gameActive.night.next()
			} else {
				gameActive.game.serv.sendMessage(tgbotapi.NewMessage(user, "Такая опция недоступна"))
			}
		} else {
			gameActive.game.serv.sendMessage(tgbotapi.NewMessage(user, "Какого фига, ты спать должен"))
		}
	}
}

func (game *Game) stopGame() {
	for _, userID := range game.nickToUser {
		game.serv.userToGame[userID] = nil
	}
	game.serv.codeToGame[game.code] = nil
}
