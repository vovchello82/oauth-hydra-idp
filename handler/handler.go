package handler

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"simple-login-endpoint/user"
	"strconv"
	"text/template"
	"time"

	hydra "github.com/ory/hydra-client-go/client"
)

type Handler struct {
	HydraClient            *hydra.OryHydra
	UserRepo               user.UserRepository
	httpClient             *http.Client
	hydra_public_url       string
	issuerUri              string
	alt_redirect_hydra_url string
}

func NewHandler(hydraClient *hydra.OryHydra, userRepo user.UserRepository) (handler *Handler) {
	insecureSkipVerify, err := strconv.ParseBool(os.Getenv("SKIP_TLS_VERIFY"))

	if err != nil {
		insecureSkipVerify = false
	}

	client := &http.Client{
		Timeout: time.Second * 10,

		Transport: &http.Transport{
			IdleConnTimeout:       time.Second * 5,
			ResponseHeaderTimeout: time.Second * 3,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: insecureSkipVerify},
		}}

	return &Handler{
		httpClient:             client,
		HydraClient:            hydraClient,
		UserRepo:               userRepo,
		issuerUri:              os.Getenv("ISSUER_URI"),
		alt_redirect_hydra_url: os.Getenv("ALTERNATIVE_REDIRECT_HYDRA_URL"),
	}
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.healthGet(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleError(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.errorGet(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) healthGet(w http.ResponseWriter, r *http.Request) {
	log.Print("GET health")
	w.WriteHeader(http.StatusOK)
	resp := make(map[string]string)
	resp["status"] = "OK"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		panic("unexpected error:" + err.Error())
	}
	if _, err := w.Write(jsonResp); err != nil {
		panic("unexpected error:" + err.Error())
	}

}

func (h *Handler) errorGet(w http.ResponseWriter, r *http.Request) {
	log.Print("GET error")
	error_title := r.URL.Query().Get("error")
	error_description := r.URL.Query().Get("error_description")

	tmpl := template.Must(template.ParseFiles("view/error.html"))

	err := tmpl.Execute(w, map[string]interface{}{
		"ErrorTitle":   error_title,
		"ErrorContent": error_description,
	})
	if err != nil {
		log.Println("error during templating: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("An expected error occured")); err != nil {
			panic("unexpected error:" + err.Error())
		}
	}
}

func (h *Handler) HandleConsent(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.consentGet(w, r)
	case http.MethodPost:
		h.consentPOST(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.loginGet(w, r)
	case http.MethodPost:
		h.loginPOST(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) RegisterClients(ctx context.Context, clientsJsonRaw []byte) {
	log.Println("importing clients...")
	timeout := time.Duration(60 * time.Second)
	ctxWithTimeOut, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ch := make(chan bool)

	go h.waitForHydraIsHealthy(ctxWithTimeOut, ch)

	select {
	case <-ctxWithTimeOut.Done():
		log.Printf("Hydra not responded after %s", timeout.String())
	case healthy := <-ch:
		log.Println("Hydra is healthy: ", healthy)

		if healthy {
			h.parseClientFile(clientsJsonRaw, h.postNewClient)
		}
	}
}

func (h *Handler) parseClientFile(jsonContent []byte, proceed func([]byte)) {
	x := bytes.TrimLeft(jsonContent, " \t\n\r")
	isArray := len(x) > 0 && x[0] == '['

	if !isArray {
		proceed(jsonContent)
	} else {
		var results []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &results); err != nil {
			log.Println("error on json unmarshaling", err.Error())
			return
		}

		for _, result := range results {
			jsonContent, err := json.Marshal(result)
			if err == nil {
				log.Println("post a new client to hydra")
				proceed(jsonContent)
			}
		}
	}
}
func (h *Handler) postNewClient(jsonContent []byte) {
	req, _ := http.NewRequest(http.MethodPost, os.Getenv("HYDRA_ADMIN_URL")+"/clients", bytes.NewBuffer(jsonContent))
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Println("Error on posting new client to hydra: ", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Println("client is not created, response code: " + resp.Status)
	}
}

func (h *Handler) waitForHydraIsHealthy(cont context.Context, ch chan bool) {
	hydra_public_url, found := os.LookupEnv("HYDRA_PUBLIC_URL")
	if !found {
		log.Fatal("HYDRA_PUBLIC_URL is not set")
	}

	h.hydra_public_url = hydra_public_url

	for i := 0; i < 100; i++ {
		select {
		case <-cont.Done():
			log.Println("TIME OUT")
			return
		default:
			_, err := h.httpClient.Get(hydra_public_url + "/health/ready")

			if err == nil {
				ch <- true
				return
			}
			log.Printf("#%d failed due to: %s ", i, err.Error())

			time.Sleep(1 * time.Second)
		}
	}

	ch <- false
}
