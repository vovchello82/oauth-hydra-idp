package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"text/template"

	"golang.org/x/oauth2"
)

// Endpoint is OAuth 2.0 endpoint.
var Endpoint = oauth2.Endpoint{

	AuthURL:   os.Getenv("AUTH_URL"),  //"https://localhost:4444/oauth2/auth",
	TokenURL:  os.Getenv("TOKEN_URL"), //"https://localhost:4444/oauth2/token",
	AuthStyle: oauth2.AuthStyleInHeader,
}

//http://localhost:1234/callbacks
// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var OAuthConf = &oauth2.Config{
	RedirectURL:  os.Getenv("REDIRECT_URL"), //"http://localhost:3000/callback", //
	ClientID:     "myclient",                //os.Getenv("CLIENT_ID"),     // TODO from hydra
	ClientSecret: "secret",                  // os.Getenv("CLIENT_SECRET"), // TODO from hydra

	// https://github.com/coreos/go-oidc/blob/v3/oidc/oidc.go#L23-L36
	// offline scope for requesting Refresh Token
	// openid for Open ID Connect
	//Scopes:   []string{"users.write", "users.read", "users.edit", "users.delete", "offline", "openid"},
	Scopes:   []string{"openid"},
	Endpoint: Endpoint,
}

var stateStore = map[string]bool{}

func main() {
	mux := http.NewServeMux()

	_, present := os.LookupEnv("AUTH_URL")
	if !present {
		log.Fatalln("AUTH_URL is not set")
	}
	_, present = os.LookupEnv("TOKEN_URL")
	if !present {
		log.Fatalln("TOKEN_URL is not set")
	}
	_, present = os.LookupEnv("REDIRECT_URL")
	if !present {
		log.Fatalln("REDIRECT_URL is not set")
	}

	mux.HandleFunc("/", homepage)
	mux.HandleFunc("/callback", callback)
	log.Println("starte app")
	log.Fatal(http.ListenAndServe(":3000", mux))
}

func homepage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("client/views/index.html"))

	// Generate random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("error on state generating", err.Error())
	}

	state := base64.StdEncoding.EncodeToString(b)

	stateStore[state] = true

	// Will return loginURL,
	// for example: http://localhost:4444/oauth2/auth?client_id=myclient&prompt=consent&redirect_uri=http%3A%2F%2Fexample.com&response_type=code&scope=users.write+users.read+users.edit&state=XfFcFf7KL7ajzA2nBY%2F8%2FX3lVzZ6VZ0q7a8rM3kOfMM%3D
	loginURL := OAuthConf.AuthCodeURL(state)
	err = tmpl.Execute(w, map[string]interface{}{
		"LoginURL": loginURL,
	})

	if err != nil {
		log.Println("error during templating: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("An expected error occured")); err != nil {
			panic("unexpected error:" + err.Error())
		}
	}
}

func callback(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("client/views/welcome.html"))

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if exist := stateStore[state]; !exist {
		log.Println("state is unknown")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("state must be set")); err != nil {
			panic("unexpected error:" + err.Error())
		}
		return
	}

	clientHTTP := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	ctx := context.WithValue(r.Context(), oauth2.HTTPClient, clientHTTP)
	accessToken, err := OAuthConf.Exchange(ctx, code)

	if err != nil {
		log.Println("error on exchange", err.Error())
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("error on getting accessToken due to " + err.Error())); err != nil {
			panic("unexpected error:" + err.Error())
		}
		return
	}

	idToken, found := accessToken.Extra("id_token").(string)
	if !found {
		log.Println("idToken not found")
		idToken = "n.a."
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"AccessToken": accessToken,
		"IdToken":     idToken,
	})

	if err != nil {
		log.Println("error during templating: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("An expected error occured")); err != nil {
			panic("unexpected error:" + err.Error())
		}
	}
}
