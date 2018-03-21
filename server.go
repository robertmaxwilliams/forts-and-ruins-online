package main

import (
	"log"
	"net/http"
	"strings"
	"encoding/json"
	"flag"
	"math/rand"
	"strconv"
	"io/ioutil"
	"github.com/googollee/go-socket.io"
)

/*
 * Runs go server for Forts and Ruins
 * Game logic is handled and verified
 * so clients can't cheat
 *
 * Make sure to run this from the same dir 
 * as all the files. TODO allow for custom serve dir
 */

type Match struct {
	Host string
	Boardsize int
}

// move = {I: i, J: j, Color : pickedColor.toString()}
// i and j are like x and y but for discrete grid
// color will be "0" to "6"
type Move struct {
  I int
  J int
  Color string
}

type Game struct {
  Board [][]string
  Host string
  Guest string
  Room string
  HostMove Move
  GuestMove Move
  KillColor string
}

func getUserList(userSet map[string]bool) []string {
    keys := make([]string, 0, len(userSet))
    for k, v := range userSet{
      if v {
        keys = append(keys, k)
      }
    }
	return keys
}

func emitUserList(so socketio.Socket, userSet map[string]bool) {
	userlist, _ := json.Marshal(getUserList(userSet))
	so.Emit("userlist", string(userlist))
	so.BroadcastTo("lobby", "userlist", string(userlist))
	log.Println(string(userlist))
}

func printBoard(game *Game) {
  return // disables this func
  log.Println(game.Room)
  for _, row := range game.Board {
    for _, cell := range row {
      if cell == "" {
        log.Print(" ")
      } else {
        log.Print(string(cell[0]))
      }
    }
    log.Println()
  }
}

// forts are created in a plus
// around last move.
// Then, forts are added in a plus
// around each fort, iterated twice
// each for is checked for being surrounded
// by something other that ''Dirt and 'B'arren
// and destoyed if it is. Iterate thrice
func growForts(game *Game, move Move, iterations int, placedForts *[]Move ) {
  if iterations == 0 {
    return
  }
  printBoard(game)
  log.Println(move)
  for _, coords := range [][]int{{0,1}, {1,0}, {-1,0}, {0,-1}} {
    i := move.I + coords[0]
    j := move.J + coords[1]
    if (i >= 15 || i < 0 || j >= 15 || j < 0) { continue }
    if (game.Board[i][j] == "") {
      // Color is actually user's name in this case
      game.Board[i][j] = move.Color
      // keep all the forts we've placed
      *placedForts = append(*placedForts, Move{I: i, J: j, Color: move.Color})
      growForts(game, Move{I: i, J: j, Color: move.Color}, iterations-1, placedForts)
    }
  }
}

//detroys forts grown by growForts
func killForts(game *Game, placedForts *[]Move) {
  log.Println(*placedForts)
  printBoard(game)
  for _, move := range *placedForts {
    i := move.I
    j := move.J
    if game.Board[move.I][move.J] == "" || game.Board[move.I][move.J] == "D" { continue }
    // after propogating forts around, if I see dirt or deadfort, 
    // delete myself
    for _, checkCoords := range [][]int{{0,1}, {1,0}, {-1,0}, {0,-1}} {
      checki := i + checkCoords[0]
      checkj := j + checkCoords[1]
      if (checki >= 15 || checki < 0 || checkj >= 15 || checkj < 0) { continue }
      if game.Board[checki][checkj] == "" || game.Board[checki][checkj] == "B" {
        game.Board[i][j] = ""
        break
      }
    }
  }
}


// wrapper func for growForts
func addForts(game *Game, lastMove Move, player string) {
  placedForts := make([]Move, 0, 128)
  growForts(game, Move{I: lastMove.I, J: lastMove.J, Color: player}, 3, &placedForts)
  for n := 0; n < 3; n++ {
    killForts(game, &placedForts)
  }
}

func killRandomColor(game *Game) {
  for i := 0; i < 15; i++ {
    for j := 0; j < 15; j++ {
      if game.Board[i][j] == game.KillColor {
		log.Println("killed a thing")
        game.Board[i][j] = "B"
		for _, checkCoords := range [][]int{{0,1}, {1,0}, {-1,0}, {0,-1}} {
		  checki := i + checkCoords[0]
		  checkj := j + checkCoords[1]
		  if (checki >= 15 || checki < 0 || checkj >= 15 || checkj < 0) { continue }
		  if game.Board[checki][checkj] == game.Host || game.Board[checki][checkj] == game.Guest{
			log.Println("converted fort to deadfort")
			  game.Board[checki][checkj] = "R"
		  }
		}
      }
    }
  }
}

// modifies game's board in place
func makeMove(game *Game, move Move) {
  if game.Board[move.I][move.J] == "" {
    game.Board[move.I][move.J] = move.Color
  }
}

func updateGame(game *Game) bool{
  log.Println("updating game ", game)
  if (game.HostMove != Move{}) && (game.GuestMove != Move{}) {
    log.Println("moves in for ", game.Room, "playing moves")
    if game.HostMove.I != game.GuestMove.I || game.HostMove.J != game.GuestMove.J {
      makeMove(game, game.HostMove)
      makeMove(game, game.GuestMove)
      addForts(game, game.HostMove, game.Host)
      addForts(game, game.GuestMove, game.Guest)
      killRandomColor(game)
    }
    game.HostMove = Move{}
    game.GuestMove = Move{}
    return true
  }
  return false
}

func main() {

	// "--testing" flag to run in local dir
	// "--disable-logs flag for server performance
	isTestingPtr := flag.Bool("testing", false, "set this flag to run on localhost under ./")
	disableLogsPtr := flag.Bool("disable-logs", false, "set this flag for performance improvement by no log.")
    flag.Parse()
	isTesting := *isTestingPtr
	disableLogs := *disableLogsPtr

	workingDir := "/home/public/forts-and-ruins-online/"
	serverAddress := "http://fortsandruins.nfshost.com"

	if isTesting {
		workingDir = "./"
		serverAddress = "http://localhost:8080"
	}

	if disableLogs {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	var userSet = make(map[string]bool)
	var matches []Match
  var games = make(map[string]*Game)

	// use WebSockets for client interaction, and 
	// FileServer for uh you know serving files.
	// Mux ties them together. TODO is Mux even needed?

	mux := http.NewServeMux()

	// copied from https://github.com/googollee/go-socket.io
	// TODO maybe put these in a seperate file or something, this giant block is unsightly
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
		room := "lobby"
		user := "anon"
		so.Join(room)

		so.On("room", func(msg string) {
			log.Println(user + " joined room ", msg)
			so.Join(msg)
			room = msg
		})

		so.On("setname", func(msg string) {
			log.Println("setname: ", msg, ", in room: ", room)
			user = strings.Trim(msg, " ")
			userSet[user] = true
			emitUserList(so, userSet)
			if len(matches) > 0 {
				matchJson, _ := json.Marshal(matches)
				so.Emit("updatedmatches", string(matchJson))
			}
		})

		so.On("unsetname", func(msg string) {
			log.Println("unsetname: ", msg, ", in room: ", room)
			so.BroadcastTo(room, "message", msg + " has logged out\n")
			user = strings.Trim(msg, " ")
			userSet[user] = false
			emitUserList(so, userSet)
		})

		so.On("message", func(msg string) {
			log.Println("message: ", msg)
			so.BroadcastTo(room, "message", msg)
		})

		so.On("makematch", func(msg string) {
			log.Println(msg)
			var newMatch Match
			_ = json.Unmarshal([]byte(msg), &newMatch)
			matches = append(matches, newMatch)
			log.Println("made new match: ", newMatch.Host, newMatch.Boardsize, " from: ", msg)
			matchJson, _ := json.Marshal(matches)
			so.BroadcastTo("lobby", "updatedmatches", string(matchJson))
			log.Println("sending callback to ", user)
			so.Emit("updatedmatches", string(matchJson))
		})

		so.On("deletematch", func(msg string) {
			var index int
			_ = json.Unmarshal([]byte(msg), &index)
			if (index >= len(matches)) {
				log.Println("bad match index from ", user)
				return
			}
			matches = append(matches[:index], matches[index+1:]...)
			log.Println("deleted a match, here's what's left: ", matches)
			matchJson, _ := json.Marshal(matches)
			so.Emit("updatedmatches", string(matchJson))
		})

		so.On("makegame", func(gameRoom string) {
			userlist := strings.SplitN(gameRoom, ":", 2)
			if len(userlist) != 2 {
				log.Println("Bad gamerooom name ", gameRoom)
				return
			}
			host := strings.Trim(userlist[0], " ")
			guest := strings.Trim(userlist[1], " ")
			if guest != user {
				log.Println("Invalid game room, should be host:guest (you are guest)")
				return
			}
			log.Println("making game room for", host, "and", guest)

			// TODO Check if both users are on lobby

			so.Emit("initiategame", gameRoom)
			so.BroadcastTo("lobby", "initiategame", gameRoom)
      boardSize := 15
      newGame := Game{Board: make([][]string, boardSize),
                      Host: host, Guest: guest, Room: gameRoom}
      for i := 0; i < boardSize; i++ {
        newGame.Board[i] = make([]string, boardSize)
      }
      log.Println(newGame.Board)
      games[gameRoom] = &newGame
		})

    so.On("playmove", func (move string) {
      log.Println(user, "played", move)
      var newMove Move
      _ = json.Unmarshal([]byte(move), &newMove)
      log.Println("unmarshalled to", newMove)
      game := games[room]
      if game.Host == user {
        game.HostMove = newMove
      } else if game.Guest == user {
        game.GuestMove = newMove
      } else {
        log.Println(user, "is in an invalid room")
      }
      game.KillColor = strconv.Itoa(rand.Intn(6))
      succeeded := updateGame(game)

      if succeeded {
        gameJson, _ := json.Marshal(game)
        so.Emit("updatedgame", string(gameJson))
        log.Println("sending back json for game: ", string(gameJson))
        so.BroadcastTo(game.Room, "updatedgame", string(gameJson))
      } else {
        so.Emit("message", "moves not valid or waiting on opponent")
      }
    })

    so.On("cancelmove", func(msg string) {
      log.Println(user, "canceled move")
      game := games[room]
      if game.Host == user {
        game.HostMove = Move{}
      } else {
        game.GuestMove = Move{}
      }
    })

		so.On("disconnection", func() {
			log.Println("on disconnect")
			userSet[user] = false
			emitUserList(so, userSet)
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	mux.Handle("/socket.io/", server)

	// servers html and js files
	// you can also load everything in the repo
	// if you ask for it. Neat feature! 
	// Speeds up how quickly hackers can find interesting bugs.
	fs := http.FileServer(http.Dir(workingDir))
	mux.Handle("/", http.StripPrefix("/", fs))

	// provides a means for the client to get the server IP
	// not sure if this makes sense but it works lol
	mux.HandleFunc("/serverip/", func (w http.ResponseWriter, r *http.Request) { w.Write([]byte(serverAddress)) })

	// run mux and Fatal log any errors.
    log.Fatal(http.ListenAndServe(":8080", mux))
}
