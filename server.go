package main

import (
    "fmt"
    "log"
    "net/http"
	"io/ioutil"
	"errors"
	"golang.org/x/tools/playground/socket"
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

	// use HundleFunc for dynamic requests, and 
	// FileServer for uh you know serving files.
	// Mux ties them together.

	mux := http.NewServeMux()

	// for receiving gamestate from users
    mux.HandleFunc("/gamestate", gameStateHandler)

	// for getting info when someone joins or leaves the page
    mux.HandleFunc("/lobby", lobbyHandler)

	// servers html and js files
	// you can also load everything in the repo
	// if you ask for it. Neat feature! 
	// Speeds up how quickly hackers can find interesting bugs.
	fs := http.FileServer(http.Dir("./"))
	mux.Handle("/", http.StripPrefix("/", fs))

	// run mux and Fatal log any errors.
    log.Fatal(http.ListenAndServe(":8080", mux))
}
