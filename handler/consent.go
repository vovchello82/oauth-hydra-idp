package handler

import (
	"context"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

func (h *Handler) consentGet(w http.ResponseWriter, r *http.Request) {
	log.Println("consentGet")
	consent_challenge := r.URL.Query().Get("consent_challenge")

	tmpl := template.Must(template.ParseFiles("view/consent.html"))

	if consent_challenge == "" {
		log.Println("consent_challenge missed")

		err := tmpl.Execute(w, nil)

		if err != nil {
			log.Println("error during templating: ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte("An expected error occured")); err != nil {
				panic("unexpected error:" + err.Error())
			}
		}

		return
	}

	consentGETparams := admin.NewGetConsentRequestParamsWithContext(r.Context()).WithHTTPClient(h.httpClient)
	consentGETparams.SetConsentChallenge(consent_challenge)

	consentGETResp, err := h.HydraClient.Admin.GetConsentRequest(consentGETparams)

	if err != nil {
		log.Println("GetConsentRequest failed", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("GetConsentRequest failed")); err != nil {
			panic("unexpected error:" + err.Error())
		}
		return
	}

	session, err := h.createSessionWithCustomClaims(consentGETResp)

	if err != nil {
		log.Println("Error on setting custom claims", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("Error on setting custom claims")); err != nil {
			panic("unexpected error:" + err.Error())
		}
		return
	}

	if consentGETResp.GetPayload().Skip {
		//grant the consent request.
		log.Println("skip consent")
		consentAcceptResp, err := h.acceptConsentRequest(r.Context(), consent_challenge, consentGETResp, session)
		if err != nil {
			log.Println("AcceptConsentRequest failed", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("AcceptConsentRequest failed")); err != nil {
				panic("unexpected error:" + err.Error())
			}
			return
		}

		redirectUrl := *consentAcceptResp.GetPayload().RedirectTo
		if len(strings.TrimSpace(h.alt_redirect_hydra_url)) > 0 {
			log.Println("use alt redirect url ", h.alt_redirect_hydra_url)
			matchUrl := h.hydra_public_url
			if len(strings.TrimSpace(h.issuerUri)) > 0 {
				matchUrl = h.issuerUri
			}
			redirectUrl = strings.Replace(redirectUrl, matchUrl, h.alt_redirect_hydra_url, 1)
		}
		log.Println("redirect to after consent: ", redirectUrl)

		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"RequestedScopes":  consentGETResp.GetPayload().RequestedScope,
		"ConsentApp":       consentGETResp.GetPayload().Client.ClientName,
		"ConsentChallenge": consent_challenge,
	})

	if err != nil {
		log.Println("error during templating: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("An expected error occured")); err != nil {
			panic("unexpected error:" + err.Error())
		}
	}
}

func (h *Handler) consentPOST(w http.ResponseWriter, r *http.Request) {
	formData := struct {
		ConsentChallenge string `validate:"required"`
	}{
		ConsentChallenge: r.FormValue("consent_challenge"),
	}

	//TODO if access denied
	//h.HydraClient.Admin.RejectConsentRequest()

	consentGETParams := admin.NewGetConsentRequestParamsWithContext(r.Context()).WithHTTPClient(h.httpClient)
	consentGETParams.SetConsentChallenge(formData.ConsentChallenge)
	consentGETResp, err := h.HydraClient.Admin.GetConsentRequest(consentGETParams)
	if err != nil {
		log.Println("error GetConsentRequest", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("GetConsentRequest failed")); err != nil {
			panic("unexpected error:" + err.Error())
		}
		return
	}

	session, err := h.createSessionWithCustomClaims(consentGETResp)

	if err != nil {
		log.Println("Error on setting custom claims", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("Error on setting custom claims")); err != nil {
			panic("unexpected error:" + err.Error())
		}
		return
	}
	//TODO  read and provide granted scopes from the form
	consentAcceptResp, err := h.acceptConsentRequest(r.Context(), formData.ConsentChallenge, consentGETResp, session)
	if err != nil {
		log.Println("AcceptConsentRequest failed", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("AcceptConsentRequest failed")); err != nil {
			panic("unexpected error:" + err.Error())
		}
		return
	}

	redirectUrl := *consentAcceptResp.GetPayload().RedirectTo
	if len(strings.TrimSpace(h.alt_redirect_hydra_url)) > 0 {
		log.Println("use alt redirect url ", h.alt_redirect_hydra_url)
		matchUrl := h.hydra_public_url
		if len(strings.TrimSpace(h.issuerUri)) > 0 {
			matchUrl = h.issuerUri
		}
		redirectUrl = strings.Replace(redirectUrl, matchUrl, h.alt_redirect_hydra_url, 1)
	}

	log.Println("after consent redirect to: ", redirectUrl)
	http.Redirect(w, r, redirectUrl, http.StatusFound)
}

func (h *Handler) acceptConsentRequest(ctx context.Context,
	consentChallenge string,
	consentGETResp *admin.GetConsentRequestOK,
	session *models.ConsentRequestSession) (acceptConsentResp *admin.AcceptConsentRequestOK, err error) {

	consentAcceptParams := admin.NewAcceptConsentRequestParamsWithContext(ctx).WithHTTPClient(h.httpClient)
	consentAcceptParams.SetConsentChallenge(consentChallenge)
	consentAcceptParams.SetBody(&models.AcceptConsentRequest{
		GrantAccessTokenAudience: consentGETResp.GetPayload().RequestedAccessTokenAudience,
		GrantScope:               consentGETResp.GetPayload().RequestedScope,
		Session:                  session,
		Remember:                 true,
	})

	consentAcceptResp, err := h.HydraClient.Admin.AcceptConsentRequest(consentAcceptParams)

	if err != nil {
		return nil, err
	}

	return consentAcceptResp, nil
}
func (h *Handler) createSessionWithCustomClaims(consentGETResp *admin.GetConsentRequestOK) (session *models.ConsentRequestSession, err error) {
	user, err := h.UserRepo.GetUserByEmail(consentGETResp.GetPayload().Subject)

	if err != nil {
		return nil, err
	}

	roles := make([]string, 0)
	if user.Roles != nil {
		roles = user.Roles
	}

	return &models.ConsentRequestSession{
		AccessToken: map[string]interface{}{
			"groups": roles,
		},
		IDToken: map[string]interface{}{
			"groups":         roles,
			"email":          user.Email,
			"email_verified": true,
		},
	}, nil
}
