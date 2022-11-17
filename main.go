package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"simple-login-endpoint/handler"
	"simple-login-endpoint/user"

	hydra "github.com/ory/hydra-client-go/client"
)

const ClientsJSONFile = "import/clients.json"
const UsersJSONFile = "import/users.json"
const ServerPort = ":3000"

func importUsers() (users map[string]*user.User) {
	jsonContent, err := os.ReadFile(UsersJSONFile)
	if err != nil {
		log.Println("Error on importing json file", err.Error())
		return make(map[string]*user.User, 0)
	}
	log.Println("importing users...")
	var imports []*user.User
	err = json.Unmarshal(jsonContent, &imports)

	if err != nil {
		log.Println("error on parsinf json content", err.Error())
		return make(map[string]*user.User, 0)
	}

	log.Printf("imported %d users", len(imports))
	userMap := make(map[string]*user.User, len(imports))

	for _, u := range imports {
		userMap[u.Email] = u
	}

	log.Println("users...", imports)
	return userMap
}

func registerClients(h *handler.Handler) {
	if _, err := os.Stat(ClientsJSONFile); errors.Is(err, os.ErrNotExist) {
		log.Println("no clients to import")
	} else {
		jsonContent, err := os.ReadFile(ClientsJSONFile)
		if err != nil {
			log.Println("Error on importing json file", err.Error())
		} else {
			ctx := context.Background()
			go h.RegisterClients(ctx, jsonContent)
		}
	}
}
func main() {
	log.Println("starting identity provider on port...", ServerPort)

	hydraAdminURL, found := os.LookupEnv("HYDRA_ADMIN_URL")
	if !found {
		log.Fatal("HYDRA_ADMIN_URL is not set")
	}
	adminURL, _ := url.Parse(hydraAdminURL)
	hydraClient := hydra.NewHTTPClientWithConfig(nil,
		&hydra.TransportConfig{
			Schemes:  []string{adminURL.Scheme},
			Host:     adminURL.Host,
			BasePath: adminURL.Path,
		},
	)

	repo := user.NewUserInMemoryRepo(importUsers())
	handler := handler.NewHandler(hydraClient, repo)
	registerClients(handler)

	mux := http.NewServeMux()

	mux.Handle("/idp/static/", http.StripPrefix("/idp/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/idp/health", handler.HandleHealth)
	mux.HandleFunc("/idp/login", handler.HandleLogin)
	mux.HandleFunc("/idp/consent", handler.HandleConsent)
	mux.HandleFunc("/idp/error", handler.HandleError)
	log.Fatal(http.ListenAndServe(ServerPort, mux))
}
