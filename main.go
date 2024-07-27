package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const timeSendUpdates = time.Millisecond * 10

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var gameState game = game{
	players: make(map[playerId]*player, 1000),
}

type playerState struct {
	player *player
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	var thisPlayer *playerState = &playerState{player: nil}
	// receive message of player
	go receiveMessagesPlayer(thisPlayer, wsConn)
	// write messages to player
	sendMessagesPlayer(thisPlayer, wsConn)
    log.Println("exiting websocket handler")
}

func receiveMessagesPlayer(state *playerState, wsConn *websocket.Conn) {
	for {
		// receive messages from player (movements)
		readedMessageType, readedMessage, readErr := wsConn.ReadMessage()
		if readErr != nil {
			log.Println("error reading message: ", readErr)
			wsConn.Close()
			gameState.removePlayer(state.player.id)
			break
		}
		if readedMessageType == websocket.CloseMessage {
			log.Println("client has closed the connection!", readErr)
			gameState.removePlayer(state.player.id)
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
			if state.player != nil {
				log.Println("player already created")
				continue
			}

			initMessage, errInitMessage := newInitMessage(msg)
			if errInitMessage != nil {
				log.Println(errInitMessage)
				continue
			}

			state.player = newPlayer(initMessage.name, initMessage.color, initMessage.initialPosition)
			gameState.initPlayer(state.player.id, state.player)
			continue
		}

		if msg.messageType == messageTypeMove {
			if state.player == nil {
				log.Println("player not created yet!")
				continue
			}

			moveMessage, errMoveMessage := newMoveMessage(msg)
			if errMoveMessage != nil {
				log.Println(errMoveMessage)
				continue
			}
			gameState.updatePlayer(state.player.id, moveMessage.position)
		}
	}
    log.Println("exiting receive messages players!")
}

func sendMessagesPlayer(state *playerState, wsConn *websocket.Conn) {
	for {
		msg := gameState.getPositionMessage(state)
		err := wsConn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("error writing message: ", err)
			break
		}
		time.Sleep(timeSendUpdates)
	}
    log.Println("exiting send messages players!")
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

    var ch chan os.Signal = make(chan os.Signal, 1)
    var chExit chan struct{} = make(chan struct{})
    signal.Notify(ch, os.Interrupt)

    go func(){
        err := server.ListenAndServe()
        if err == http.ErrServerClosed {
            log.Println("server closed!")
        }
        chExit<-struct{}{}
    }()

    <-ch
    log.Println("closing server!")
    server.Shutdown(context.Background())
    <-chExit
}
