package game

type Night struct {
	gActive   *GameActive
	offset    int
	shot      int64
	healed    int64
	witnessed int64
}

func (ga *GameActive) NewNight() {
	ga.night = &Night{
		gActive: ga,
		offset:  -1,
	}
}

func (n *Night) playerCanAct(member Player) []string {
	list := make([]string, 0)
	if member.Role == Peaceful {
		list = append(list, "сделать ничего")
	}
	for _, player := range n.gActive.pQueue {
		if member.Role == Mafia {
			if player.Role != Mafia {
				list = append(list, n.gActive.game.UserToNick[player.User])
			}
		} else if member.Role == Doctor {
			if n.gActive.healed != player.User {
				list = append(list, n.gActive.game.UserToNick[player.User])
			}
		} else if member.Role == Witness {
			if n.gActive.witnessed != player.User {
				list = append(list, n.gActive.game.UserToNick[player.User])
			}
		}
	}
	return list
}

func (n *Night) handleAct(user int64, victim string) {

	if player := n.gActive.pQueue[n.offset]; player.User == user {

		if contains(n.playerCanAct(player), victim) {
			e := ActEndedEvent{
				Player:  player,
				Success: true,
			}

			switch player.Role {
			case Mafia:
				if shot := n.gActive.game.NickToUser[victim]; n.shot != 0 && n.shot != shot { //TODO
					e.Success = false
					n.shot = 0
				} else {
					n.shot = shot
				}
			case Doctor:
				n.healed = n.gActive.game.NickToUser[victim]
			case Witness:
				n.witnessed = n.gActive.game.NickToUser[victim]
			}

			if n.offset+1 < len(n.gActive.pQueue) {
				e.Next = n.gActive.game.UserToNick[n.gActive.pQueue[n.offset+1].User]
			}
			n.gActive.game.EOutput.HandleActEnded(e)
			n.next()
		} else {
			n.gActive.game.EOutput.HandleUnsupportedAct(UnsupportedActEvent{
				user,
			})
		}
	} else {
		n.gActive.game.EOutput.HandleUnexpectedActTrial(UnexpectedActTrialEvent{
			user,
		})
	}
}

func (n *Night) next() {
	n.offset++
	if n.offset < len(n.gActive.pQueue) {
		player := n.gActive.pQueue[n.offset]

		n.gActive.game.EOutput.HandleNightAct(NightActEvent{
			player,
			n.playerCanAct(player),
		})
	} else {
		e := NightEndedEvent{
			Users: n.gActive.game.GetUsers(),
		}
		if n.shot != 0 && n.shot != n.healed {
			e.Killed = n.gActive.game.UserToNick[n.shot]
			n.gActive.removePlayer(n.shot)
		}
		n.gActive.healed = n.healed
		n.gActive.witnessed = n.witnessed

		n.gActive.game.EOutput.HandleNightEnded(e)
		n.gActive.startDay()
	}
}
