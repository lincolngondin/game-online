package main

import "math/rand"

type playerId int64

type player struct {
	id       playerId
	name     string
	color    string
	position point
}

type point struct {
	x int
	y int
}

func newPlayer(name, color string, position point) *player {
	return &player{
		id:       playerId(rand.Int63()),
		name:     name,
		color:    color,
		position: position,
	}
}
