package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

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
	initMessageType, initMessage, initMsgErr := wsConn.ReadMessage()
	if initMsgErr != nil || initMessageType != websocket.TextMessage {
		log.Println("error reading message: ", initMsgErr)
		wsConn.Close()
		return
	}
	log.Println("received: ", string(initMessage))
	cmds := strings.Split(string(initMessage), ";")
	if cmds[0] != "init" {
		log.Println("init message not send yet!")
		return
	}
	posStr := strings.Split(cmds[3], ",")
	px, _ := strconv.ParseInt(posStr[0], 10, 64)
	py, _ := strconv.ParseInt(posStr[1], 10, 64)
	var ptn = point{int(px), int(py)}
	thisPlayer = newPlayer(cmds[1], cmds[2], ptn)
	gameState.players[thisPlayer.id] = thisPlayer

    // receive message of player
    go receiveMessagesPlayer(thisPlayer.id, wsConn)
	// write messages to player
	sendMessagesPlayer(thisPlayer.id, wsConn)
}

func receiveMessagesPlayer(playerId playerId, wsConn *websocket.Conn) {
	for {
		// receive messages from player (movements)
		_, moveMessage, moveMessageErr := wsConn.ReadMessage()
		if moveMessageErr != nil {
			log.Println("error reading message: ", moveMessageErr)
			wsConn.Close()
			delete(gameState.players, playerId)
			break
		}
		moveCmdsStr := strings.Split(string(moveMessage), ";")
		posMoveStr := strings.Split(moveCmdsStr[1], ",")
		movepx, _ := strconv.ParseInt(posMoveStr[0], 10, 64)
		movepy, _ := strconv.ParseInt(posMoveStr[1], 10, 64)
		var movePtn = point{int(movepx), int(movepy)}
		if gameState.players[playerId] != nil {
			gameState.players[playerId].position.x = movePtn.x
			gameState.players[playerId].position.y = movePtn.y
		}
	}
}

func sendMessagesPlayer(playerId playerId, wsConn *websocket.Conn) {
	for {
		builder := strings.Builder{}
		builder.WriteString("pos")
		for key, value := range gameState.players {
			if key == playerId {
				continue
			}
			builder.WriteString(fmt.Sprintf(";%s;%s;%d,%d", value.name, value.color, value.position.x, value.position.y))
		}
		str := builder.String()
		err := wsConn.WriteMessage(websocket.TextMessage, []byte(str))
		if err != nil {
			log.Println(err)
			break
		}
		time.Sleep(time.Millisecond * 10)
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
