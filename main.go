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
	players map[string]*player
}

type player struct {
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
		name:     name,
		color:    color,
		position: position,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var gameState game = game{
	players: make(map[string]*player, 1000),
}

func main() {
	tp, errTemplate := template.ParseGlob("*.html")
	if errTemplate != nil {
		log.Fatal(errTemplate)
	}

	mux := http.DefaultServeMux

	mux.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		tp.ExecuteTemplate(w, "index.html", nil)
	})

	mux.HandleFunc("/server", func(w http.ResponseWriter, r *http.Request) {
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
		gameState.players[thisPlayer.name] = thisPlayer

		go func() {
			for {
				// receive messages from player (movements)
				_, moveMessage, moveMessageErr := wsConn.ReadMessage()
				if moveMessageErr != nil {
					log.Println("error reading message: ", moveMessageErr)
                    wsConn.Close()
                    delete(gameState.players, thisPlayer.name)
                    break
				}
				moveCmdsStr := strings.Split(string(moveMessage), ";")
				posMoveStr := strings.Split(moveCmdsStr[1], ",")
				movepx, _ := strconv.ParseInt(posMoveStr[0], 10, 64)
				movepy, _ := strconv.ParseInt(posMoveStr[1], 10, 64)
				var movePtn = point{int(movepx), int(movepy)}
                if gameState.players[thisPlayer.name]!=nil{
                    gameState.players[thisPlayer.name].position.x = movePtn.x
                    gameState.players[thisPlayer.name].position.y = movePtn.y
                }
			}
		}()

        
        // write messages to player
        for {
            builder := strings.Builder{}
            builder.WriteString("pos")
            for _, value := range gameState.players {
                if value.name == thisPlayer.name {
                    continue
                }
                builder.WriteString(fmt.Sprintf(";%s;%s;%d,%d", value.name, value.color, value.position.x, value.position.y))
            }
            str := builder.String()
            //log.Println("write: ", str)
            err := wsConn.WriteMessage(websocket.TextMessage, []byte(str))
            if err != nil {
                break
            }
            time.Sleep(time.Millisecond*50)
        }

	})


	server := http.Server{
		Addr:    ":8089",
		Handler: mux,
	}


	server.ListenAndServe()

}
