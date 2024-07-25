package main

import (
	"errors"
	"strconv"
	"strings"
)

const (
	messageTypeInit string = "init"
	messageTypePos  string = "pos"
	messageTypeMove string = "move"
)

var (
	errInvalidMessage       = errors.New("message is invalid!")
	errInvalidType          = errors.New("message type is invalid!")
	errInvalidMessageParameter          = errors.New("message parameter is invalid!")
	errDifferentMessageType = errors.New("message different of what expected!")
)

type message struct {
	messageType string
	data        []string
}

type messageMove struct {
	position point
}

func newMoveMessage(msg *message) (*messageMove, error) {
    if msg.messageType != messageTypeMove {
        return nil, errDifferentMessageType
    }
	px, convErr := strconv.ParseInt(msg.data[0], 10, 64)
    if convErr != nil {
        return nil, errInvalidMessageParameter
    }
    py, convErr := strconv.ParseInt(msg.data[1], 10, 64)
    if convErr != nil {
        return nil, errInvalidMessageParameter
    }
    return &messageMove{
        position: point{int(px), int(py)},
    }, nil

}

type messageInit struct {
	name            string
	color           string
	initialPosition point
}

func newInitMessage(msg *message) (*messageInit, error) {
    if msg.messageType != messageTypeInit {
        return nil, errDifferentMessageType
    }
    name := msg.data[0]
    color := msg.data[1]
	px, convErr := strconv.ParseInt(msg.data[2], 10, 64)
    if convErr != nil {
        return nil, errInvalidMessageParameter
    }
    py, convErr := strconv.ParseInt(msg.data[3], 10, 64)
    if convErr != nil {
        return nil, errInvalidMessageParameter
    }
    return &messageInit{
        name: name,
        color: color,
        initialPosition: point{int(px), int(py)},
    }, nil

}

func newMessage(data string) (*message, error) {
	dataSlice := strings.Split(data, ";")
	if len(dataSlice) < 1 {
		return nil, errInvalidMessage
	}
	var messageType string
	switch dataSlice[0] {
	case messageTypeInit:
		messageType = messageTypeInit
	case messageTypePos:
		messageType = messageTypePos
	case messageTypeMove:
		messageType = messageTypeMove
	default:
		return nil, errInvalidMessage
	}
	if messageType == messageTypeInit && len(dataSlice) != 5 {
		return nil, errInvalidMessage
	}
	if messageType == messageTypeMove && len(dataSlice) != 3 {
		return nil, errInvalidMessage
	}
	return &message{
		messageType: messageType,
		data:        dataSlice[1:],
	}, nil
}
