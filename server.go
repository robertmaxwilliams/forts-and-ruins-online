package main

import (
    "fmt"
    "log"
    "net/http"
	"strings"
	"encoding/json"
	"flag"

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

func findString(arr []string, s string) int {
	for i := range arr{
		if arr[i] == s {
			return i
		}
	}
	return -1
}
func removeByValue(arr []string, s string) []string {
	index := findString(arr, s)
	arr = append(arr[:index], arr[index+1:]...)
	return arr
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

func main() {

	// "--testing" flag to run in local dir
	isTestingPtr := flag.Bool("testing", false, "set this flag to run on localhost under ./")
    flag.Parse()
	isTesting := *isTestingPtr
	workingDir := "/home/public/forts-and-ruins-online/"
	serverAddress := "http://fortsandruins.nfshost.com"
	if isTesting {
		workingDir = "./"
		serverAddress = "http://localhost:8080"
	}


	var userSet = make(map[string]bool)
	var matches []Match


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
			fmt.Println(user + " joined room ", msg)
			so.Join(msg)
			room = msg
		})

		so.On("setname", func(msg string) {
			fmt.Println("setname: ", msg, ", in room: ", room)
			user = strings.Trim(msg, " ")
			userSet[user] = true
			emitUserList(so, userSet)
			if len(matches) > 0 {
				matchJson, _ := json.Marshal(matches)
				so.Emit("updatedmatches", string(matchJson))
			}
		})

		so.On("unsetname", func(msg string) {
			fmt.Println("unsetname: ", msg, ", in room: ", room)
			so.BroadcastTo(room, "message", msg + " has logged out\n")
			user = strings.Trim(msg, " ")
			userSet[user] = false
			emitUserList(so, userSet)
		})

		so.On("message", func(msg string) {
			fmt.Println("message: ", msg)
			so.BroadcastTo(room, "message", msg)
		})

		so.On("makematch", func(msg string) {
			fmt.Println(msg)
			var newMatch Match
			_ = json.Unmarshal([]byte(msg), &newMatch)
			matches = append(matches, newMatch)
			fmt.Println("made new match: ", newMatch.Host, newMatch.Boardsize, " from: ", msg)
			matchJson, _ := json.Marshal(matches)
			so.BroadcastTo("lobby", "updatedmatches", string(matchJson))
			fmt.Println("sending callback to ", user)
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
			fmt.Println("deleted a match, here's what's left: ", matches)
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
			fmt.Println("making game room for", host, "and", guest)

			// TODO Check if both users are on lobby

			so.Emit("initiategame", gameRoom)
			so.BroadcastTo("lobby", "initiategame", gameRoom)
		})

		so.On("disconnection", func() {
			fmt.Println("on disconnect")
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
