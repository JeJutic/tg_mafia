package main

import (
	game "github.com/jejutic/tg_mafia/pkg"
)

type userMessage struct {
	user    int64
	text    string
	command bool
}

type serverMessage struct {
	user    int64
	text    string
	options []string
}

func newMessageKeepOptions(user int64, text string) serverMessage {
	return serverMessage{
		user: user,
		text: text,
	}
}

func newMessageRemoveOptions(user int64, text string) serverMessage {
	return serverMessage{
		user:    user,
		text:    text,
		options: make([]string, 0),
	}
}

type server[T any] interface {
	getUpdatesChan() <-chan T
	updateToMessage(T) *userMessage // didn't want to make an extra goroutine for casting of updates from chan
	sendMessage(serverMessage)
	getDefaultNick(int64) string
}

func sendAll[T any](s server[T], users []int64, text string) {
	for _, user := range users {
		s.sendMessage(newMessageRemoveOptions(user, text))
	}
}

type mafiaServer[T any] struct {
	server[T]
	userToGame map[int64]*game.Game
	codeToGame map[int]*game.Game
}

func newMafiaServer[T any](s server[T]) mafiaServer[T] {
	return mafiaServer[T]{
		server:     s,
		userToGame: make(map[int64]*game.Game),
		codeToGame: make(map[int]*game.Game),
	}
}

func run[T any](ms mafiaServer[T]) {

	for update := range ms.getUpdatesChan() {
		msg := ms.updateToMessage(update)
		if msg == nil {
			continue
		}

		if msg.command {
			handleCommand(ms, *msg)
		} else {
			if game := ms.userToGame[msg.user]; game != nil && game.Started() {
				game.GActive.Handle(msg.user, msg.text)
			} else if game == nil {
				handleCommand(ms, userMessage{
					user: msg.user,
					text: "/join " + msg.text,
				})
			} else {
				ms.sendMessage(serverMessage{
					user: msg.user,
					text: "Дождитесь начала игры",
				})
			}
		}
	}
}
