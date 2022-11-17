package handler

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

func (h *Handler) showErrorPage(w http.ResponseWriter, errorTitle string, errorContent string) {
	tmpl := template.Must(template.ParseFiles("view/error.html"))
	w.WriteHeader(http.StatusBadRequest)
	err := tmpl.Execute(w, map[string]interface{}{
		"ErrorTitle":   errorTitle,
		"ErrorContent": errorContent,
	})

	if err != nil {
		log.Println("error during templating: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("An expected error occured")); err != nil {
			panic("unexpected error:" + err.Error())
		}
	}
}
func (h *Handler) loginGet(w http.ResponseWriter, r *http.Request) {
	login_chalenge := r.URL.Query().Get("login_challenge")
	log.Print("GET login ")

	if login_chalenge == "" {
		h.showErrorPage(w, "login_chalenge missed", "Login Chalenge muss als Query Parameter gesetzt werden")
		return
	}

	tmpl := template.Must(template.ParseFiles("view/login.html"))

	loginGetParam := admin.NewGetLoginRequestParamsWithHTTPClient(h.httpClient)
	loginGetParam.WithContext(r.Context())
	loginGetParam.SetLoginChallenge(login_chalenge)

	respLoginGet, err := h.HydraClient.Admin.GetLoginRequest(loginGetParam)
	if err != nil {
		log.Println("GetLoginRequest failed", err.Error())
		h.showErrorPage(w, "Fehler beim Starten von Code Flow ", "Bitte wiederholen Sie den Vorgang")
		return
	}

	skip := false
	if respLoginGet.GetPayload() != nil && respLoginGet.GetPayload().Skip != nil {
		skip = *respLoginGet.GetPayload().Skip
	}

	if skip {
		log.Print("skip login")

		respLoginAccept, err := h.acceptLoginRequest(r.Context(), respLoginGet.GetPayload().Subject, login_chalenge, true)

		if err != nil {
			log.Println("AcceptLoginRequest failed", err.Error())
			h.showErrorPage(w, "Fehler beim Starten von Code Flow ", "Bitte wiederholen Sie den Vorgang")
			return
		}

		redirectUrl := *respLoginAccept.GetPayload().RedirectTo
		if len(strings.TrimSpace(h.alt_redirect_hydra_url)) > 0 {
			redirectUrl = strings.Replace(redirectUrl, h.hydra_public_url, h.alt_redirect_hydra_url, 1)
		}
		log.Println("redirect to consent: ", redirectUrl)
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"LoginChallenge": login_chalenge,
	})
	if err != nil {
		log.Println("error during templating: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("An expected error occured")); err != nil {
			panic("unexpected error:" + err.Error())
		}
	}
}

func (h *Handler) loginPOST(w http.ResponseWriter, r *http.Request) {

	formData := struct {
		LoginChallenge string `validate:"required"`
		Email          string `validate:"required"`
		Password       string `validate:"required"`
		Remember       string `validate:"required"`
	}{
		LoginChallenge: r.FormValue("login_challenge"),
		Email:          r.FormValue("username"),
		Password:       r.FormValue("password"),
		Remember:       r.FormValue("remember"),
	}

	//TODO VZ implemenent loginReject Call
	if !h.isUserValid(formData.Email, formData.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		tmpl := template.Must(template.ParseFiles("view/login.html"))
		err := tmpl.Execute(w, map[string]interface{}{
			"LoginChallenge": formData.LoginChallenge,
			"ErrorTitle":     "Benutzername/Password falsch",
			"ErrorContent":   "Korrigieren Sie Ihre Angaben",
		})

		if err != nil {
			log.Println("error during templating: ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte("An expected error occured")); err != nil {
				panic("unexpected error:" + err.Error())
			}
		}

		return
	}

	loginParams := admin.NewGetLoginRequestParamsWithHTTPClient(h.httpClient).WithContext(r.Context())
	loginParams.SetLoginChallenge(formData.LoginChallenge)

	_, err := h.HydraClient.Admin.GetLoginRequest(loginParams)
	if err != nil {
		log.Println("error GetLoginRequest", err.Error())
		h.showErrorPage(w, "Fehler beim Starten von Code Flow ", "Bitte wiederholen Sie den Vorgang")
		return
	}

	respLoginAccept, err := h.acceptLoginRequest(r.Context(), &formData.Email, formData.LoginChallenge, formData.Remember == "on")
	if err != nil {
		// if error, redirects to ...
		log.Println("error AcceptLoginRequest", err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		if _, err := w.Write([]byte("AcceptLoginRequest failed")); err != nil {
			panic("unexpected error:" + err.Error())
		}
		return
	}
	redirectUrl := *respLoginAccept.GetPayload().RedirectTo
	if len(strings.TrimSpace(h.alt_redirect_hydra_url)) > 0 {
		log.Println("use alt redirect url ", h.alt_redirect_hydra_url)
		matchUrl := h.hydra_public_url
		if len(strings.TrimSpace(h.issuerUri)) > 0 {
			matchUrl = h.issuerUri
		}
		redirectUrl = strings.Replace(redirectUrl, matchUrl, h.alt_redirect_hydra_url, 1)
	}
	log.Println("after login redirect to consent: ", redirectUrl)
	// then show the consent form
	http.Redirect(w, r, redirectUrl, http.StatusFound)
}

func (h *Handler) acceptLoginRequest(ctx context.Context, subject *string, login_chalenge string, remember bool) (response *admin.AcceptLoginRequestOK, err error) {
	loginAcceptParams := admin.NewAcceptLoginRequestParamsWithContext(ctx).WithHTTPClient(h.httpClient)
	loginAcceptParams.SetLoginChallenge(login_chalenge)
	loginAcceptParams.SetBody(&models.AcceptLoginRequest{
		Subject:  subject,
		Remember: remember,
	})

	return h.HydraClient.Admin.AcceptLoginRequest(loginAcceptParams)
}

func (h *Handler) isUserValid(email string, password string) bool {
	if user, err := h.UserRepo.GetUserByEmail(email); err == nil {
		return user.Password == password
	}
	return false
}
