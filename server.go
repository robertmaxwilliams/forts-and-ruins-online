package main

import (
    "fmt"
    "log"
    "net/http"
	"strings"
//	"io/ioutil"
//	"errors"
	//"golang.org/x/tools/playground/socket"
	//"gopkg.in/googollee/go-socket.io.v1"
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

func main() {
	var users = make(map[string]bool)


	// use WebSockets for client interaction, and 
	// FileServer for uh you know serving files.
	// Mux ties them together. TODO is Mux even needed?

	mux := http.NewServeMux()

	// copied from https://github.com/googollee/go-socket.io
	// TODO mess with once everything works
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
		room := "default"
		user := "default"
		so.Join(room)

		so.On("room", func(msg string) {
			fmt.Println("user joined room ", msg)
			so.Join(msg)
			room = msg
		})

		so.On("setname", func(msg string) {
			fmt.Println("got: ", msg, ", in room: ", room)
			so.BroadcastTo(room, "message", "you suck")
			so.Emit("message", "I hate you")
			user = strings.Trim(msg, " ")
			users[user] = true
		})

		so.On("disconnection", func() {
			fmt.Println("on disconnect")
			users[user] = false
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
	fs := http.FileServer(http.Dir("./"))
	mux.Handle("/", http.StripPrefix("/", fs))

	// run mux and Fatal log any errors.
    log.Fatal(http.ListenAndServe(":8080", mux))
}
