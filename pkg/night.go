package game

import "strings"

// night represents night phase of an active mafia game
type night struct {
	gActive   *gameActive
	offset    int
	shot      int64
	healed    int64
	witnessed int64
}

// newNight sets night field in gActive ga
func (ga *gameActive) newNight() {
	ga.night = &night{
		gActive: ga,
		offset:  -1,
	}
}

func (n *night) playerCanAct(member Player) []string {
	list := make([]string, 0)
	if member.Role == Peaceful || (member.Role == Maniac && n.gActive.mafiaAlive()) {
		list = append(list, "сделать ничего")
		return list
	}
	for _, player := range n.gActive.pQueue {
		switch member.Role {
		case Mafia:
			if player.Role != Mafia {
				list = append(list, n.gActive.UserToNick[player.User])
			}
		case Doctor:
			if n.gActive.healed != player.User {
				list = append(list, n.gActive.UserToNick[player.User])
			}
		case Witness:
			if n.gActive.witnessed != player.User {
				list = append(list, n.gActive.UserToNick[player.User])
			}
		case Sheriff:
			for _, p2 := range n.gActive.pQueue {
				if member.User != player.User && member.User != p2.User && player.User < p2.User { //can make 2x faster with i, j
					list = append(list,
						n.gActive.UserToNick[player.User]+
							" "+
							n.gActive.UserToNick[p2.User],
					)
				}
			}
		case Maniac:
			if player.User != member.User {
				list = append(list, n.gActive.UserToNick[player.User])
			}
		}
	}
	return list
}

func (n *night) handleAct(user int64, victim string) {

	if player := n.gActive.pQueue[n.offset]; player.User == user {

		if contains(n.playerCanAct(player), victim) {
			e := ActEndedEvent{
				Player:  player,
				Success: true,
			}

			switch player.Role {
			case Mafia:
				if shot := n.gActive.NickToUser[victim]; n.shot != 0 && n.shot != shot {
					e.Success = false
					n.shot = 0
				} else {
					n.shot = shot
				}
			case Doctor:
				n.healed = n.gActive.NickToUser[victim]
			case Witness:
				n.witnessed = n.gActive.NickToUser[victim]
			case Sheriff:
				victims := strings.Split(victim, " ")
				vRole := [2]Role{}
				for i, v := range victims {
					vRole[i] = n.gActive.userToPlayer(n.gActive.NickToUser[v]).Role
				}
				e.Success = roleToSide[vRole[0]] == roleToSide[vRole[1]]
			case Maniac:
				if !n.gActive.mafiaAlive() {
					n.shot = n.gActive.NickToUser[victim]
				}
			}

			if n.offset+1 < len(n.gActive.pQueue) {
				e.Next = n.gActive.UserToNick[n.gActive.pQueue[n.offset+1].User]
			}
			n.gActive.EOutput.HandleActEnded(e)
			n.next()
		} else {
			n.gActive.EOutput.HandleUnsupportedAct(UnsupportedActEvent{
				user,
			})
		}
	} else {
		n.gActive.EOutput.HandleUnexpectedActTrial(UnexpectedActTrialEvent{
			user,
		})
	}
}

func (n *night) next() {
	n.offset++
	if n.offset < len(n.gActive.pQueue) {
		player := n.gActive.pQueue[n.offset]

		n.gActive.EOutput.HandleNightAct(NightActEvent{
			player,
			n.playerCanAct(player),
			n.gActive.mafiaAlive(),
		})
	} else {
		e := NightEndedEvent{
			Users: n.gActive.GetUsers(),
		}
		if n.shot != 0 && n.shot != n.healed {
			e.Killed = n.gActive.UserToNick[n.shot]
			n.gActive.removePlayer(n.shot)
		}
		n.gActive.healed = n.healed
		n.gActive.witnessed = n.witnessed

		n.gActive.EOutput.HandleNightEnded(e)
		n.gActive.startDay()
	}
}
