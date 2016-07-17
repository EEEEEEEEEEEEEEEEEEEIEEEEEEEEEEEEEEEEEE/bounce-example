package main

import (
	"crypto/rand"
	"github.com/s-rah/bounce"
	"html"
	"html/template"
	"io"
	"net/http"
	"time"
)

type BounceInfo struct {
	Address string
}

// This is a really simple service to demo how Bouce could work as an authentication mechanism.
func main() {

	// Generate a random token encryption key for each run
	// We could load a key from a file if we wanted a long lived session id
	// (and to allow clients to connect even after a server restart)
	var tokenKey [32]byte
	io.ReadFull(rand.Reader, tokenKey[:])

	// Setup Bounce Service
	bounceService := new(bounce.BounceService)
	bounceService.Init("./private_key")
	bounceService.InitTokenService("http://trc4t4bl7d65hdiq.onion", tokenKey, time.Hour*24)
	go bounceService.Listen(bounceService, 12345)

	var templates = template.Must(template.ParseFiles("index.html", "login.html", "header.html", "secret.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", &BounceInfo{})
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		address := html.EscapeString(r.FormValue("address"))
		go bounceService.SendToken(address)
		templates.ExecuteTemplate(w, "login.html", &BounceInfo{Address: address})
	})

	http.HandleFunc("/bounce", func(w http.ResponseWriter, r *http.Request) {
		address := html.EscapeString(r.FormValue("address"))
		token := html.EscapeString(r.FormValue("token"))

		if bounceService.ValidateToken(token, address) == true {
			// We could also set a cookie here, for a long lived session
			templates.ExecuteTemplate(w, "secret.html", &BounceInfo{})
		} else {
			http.Redirect(w, r, "/", http.StatusForbidden)
		}
	})
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))
	http.ListenAndServe("127.0.0.1:8080", nil)
}
