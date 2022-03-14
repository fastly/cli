package auth

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/pkg/browser"
)

// Opener represents something that can open a URL.
type Opener func(string) error

// Browser is the default implementation of an Opener.
var Browser Opener = browser.OpenURL

// ListenAndServe listens on a random TCP network port and then calls Serve with
// custom handler type to handle requests on incoming connections.
//
// NOTE: This function is expected to be called asynchronously, hence channels
// are provided for synchronised communication.
func ListenAndServe(port chan int, token chan string) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	port <- l.Addr().(*net.TCPAddr).Port
	http.Serve(l, Server{token})
}

// Server represents a HTTP Server.
type Server struct {
	token chan string
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth-callback":
		s.authCallbackHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No page found for: %s\n", r.URL)
	}
}

// NOTE: authCallbackHandler is expected to be provided with a token query.
func (s Server) authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query().Get("access_token")
	if t == "" {
		// TODO: Handle `auth_error` query param!
		// TODO: Check --token doesn't get persisted to disk!
		// TODO: Invalidate old token by checking its expiry is set to never expire (migration scenario for someone with long lived token in their config).
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "unable to parse an authentication token")
	}
	s.token <- t
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Authentication complete. You may close this browser tab and return to your terminal.")
}
