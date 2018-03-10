package main

import (
    "fmt"
    "log"
    "net/http"
	"io/ioutil"
	"errors"
	//"golang.org/x/tools/playground/socket"
	//"gopkg.in/googollee/go-socket.io.v1"
	"github.com/googollee/go-socket.io"
)

func getJsonFromRequest(w http.ResponseWriter, r * http.Request) (string, error) {
	if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body",
					http.StatusInternalServerError)
			}
			return string(body), nil
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Println("Client sent bad request")
		return "", errors.New("Invalid request method from client")
	}
}

func lobbyHandler(w http.ResponseWriter, r *http.Request) {
	json, httpErr := getJsonFromRequest(w, r)
	if httpErr != nil {
		return
	}
	fmt.Println(json)
	w.Write([]byte("hello"))
}

func gameStateHandler(w http.ResponseWriter, r *http.Request) {
	json, httpErr := getJsonFromRequest(w, r)
	if httpErr != nil {
		return
	}
	fmt.Println(json)
	w.Write([]byte("hello"))
}

/*
 * Runs go server for Forts and Ruins
 * Game logic is handled and verified
 * so clients can't cheat
 *
 * Make sure to run this from the same dir 
 * as all the files. TODO allow for custom serve dir
 */

func main() {

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
		so.Join("chat")
		so.On("chat message", func(msg string) {
			log.Println("emit:", so.Emit("chat message", msg))
			so.BroadcastTo("chat", "chat message", msg)
		})
		so.On("disconnection", func() {
			log.Println("on disconnect")
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
