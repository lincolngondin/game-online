package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const timeSendUpdates = time.Millisecond * 10

type game struct {
	players map[playerId]*player
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var gameState game = game{
	players: make(map[playerId]*player, 1000),
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	var thisPlayer *player
	// receive message of player
	go receiveMessagesPlayer(thisPlayer, wsConn)
	// write messages to player
	sendMessagesPlayer(thisPlayer, wsConn)
}

func receiveMessagesPlayer(player *player, wsConn *websocket.Conn) {
	for {
		// receive messages from player (movements)
		readedMessageType, readedMessage, readErr := wsConn.ReadMessage()
		if readErr != nil {
			log.Println("error reading message: ", readErr)
			wsConn.Close()
			delete(gameState.players, player.id)
			break
		}
		if readedMessageType == websocket.CloseMessage {
			log.Println("client has closed the connection!", readErr)
			delete(gameState.players, player.id)
			break
		}
		if readedMessageType != websocket.TextMessage {
			log.Println("received non text message, ignoring!", readErr)
			continue
		}

		msg, errMessage := newMessage(string(readedMessage))
		if errMessage != nil {
			log.Println(errMessage)
			continue
		}

		if msg.messageType == messageTypeInit {
			if player != nil {
				log.Println("player already created")
				continue
			}

			initMessage, errInitMessage := newInitMessage(msg)
			if errInitMessage != nil {
				log.Println(errInitMessage)
				continue
			}

			player = newPlayer(initMessage.name, initMessage.color, initMessage.initialPosition)
			gameState.players[player.id] = player
			continue
		}

		if msg.messageType == messageTypeMove {
			if player == nil {
				log.Println("player not created yet!")
				continue
			}

			moveMessage, errMoveMessage := newMoveMessage(msg)
			if errMoveMessage != nil {
				log.Println(errMoveMessage)
				continue
			}
			if gameState.players[player.id] != nil {
				gameState.players[player.id].position.x = moveMessage.position.x
				gameState.players[player.id].position.y = moveMessage.position.y
			}
		}
	}
}

func sendMessagesPlayer(player *player, wsConn *websocket.Conn) {
	for {
		builder := strings.Builder{}
		builder.WriteString("pos")
		for key, value := range gameState.players {
			if player != nil && key == player.id {
				continue
			}
			builder.WriteString(fmt.Sprintf(";%s;%s;%d;%d", value.name, value.color, value.position.x, value.position.y))
		}
		str := builder.String()
		err := wsConn.WriteMessage(websocket.TextMessage, []byte(str))
		if err != nil {
			log.Println("error writing message: ", err)
			break
		}
		time.Sleep(timeSendUpdates)
	}
}

func main() {
	tp, errTemplate := template.ParseGlob("*.html")
	if errTemplate != nil {
		log.Fatal(errTemplate)
	}

	mux := http.DefaultServeMux

	server := http.Server{
		Addr:    ":8089",
		Handler: mux,
	}

	mux.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		tp.ExecuteTemplate(w, "index.html", fmt.Sprintf("ws://%s/server", r.Host))
	})

	mux.HandleFunc("/server", websocketHandler)

	server.ListenAndServe()
}
