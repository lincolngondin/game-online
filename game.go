package main

import (
	"fmt"
	"strings"
	"sync"
)

type game struct {
	players map[playerId]*player
	mu      sync.RWMutex
}

func (g *game) initPlayer(playerId playerId, player *player) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.players[player.id] = player
}

func (g *game) removePlayer(playerId playerId) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.players, playerId)
}

func (g *game) updatePlayer(playerId playerId, newPosition point) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.players[playerId] != nil {
		g.players[playerId].position = newPosition
	}
}

func (g *game) getPositionMessage(state *playerState) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	builder := strings.Builder{}
	builder.WriteString("pos")
	for key, value := range g.players {
		if state.player != nil && key == state.player.id {
			continue
		}
		builder.WriteString(fmt.Sprintf(";%s;%s;%d;%d", value.name, value.color, value.position.x, value.position.y))
	}
	return builder.String()
}
