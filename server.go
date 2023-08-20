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
