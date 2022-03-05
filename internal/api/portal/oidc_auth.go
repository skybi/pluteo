package portal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/skybi/data-server/internal/random"
	"net/http"
	"time"
)

var (
	cookieNameToken = "session_token"

	stateLength         = 16
	nonceLength         = 16
	cookieNameState     = "login_state"
	cookieLifetimeState = int(time.Hour.Seconds())
)

type oidcLoginFlowState struct {
	ID         string `json:"id"`
	Nonce      string `json:"nonce"`
	Afterwards string `json:"afterwards"`
}

// EndpointOIDCLoginFlow handles the 'GET /v1/auth/oidc/login_flow' endpoint
func (service *Service) EndpointOIDCLoginFlow(writer http.ResponseWriter, request *http.Request) {
	afterwards := request.URL.Query().Get("afterwards")

	// Create and set the login flow state cookie
	state := oidcLoginFlowState{
		ID:         random.String(stateLength, random.CharsetAlphanumeric),
		Nonce:      random.String(nonceLength, random.CharsetAlphanumeric),
		Afterwards: afterwards,
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		service.internalError(writer, err)
		return
	}
	http.SetCookie(writer, &http.Cookie{
		Name:     cookieNameState,
		Value:    base64.StdEncoding.EncodeToString(stateJSON),
		MaxAge:   cookieLifetimeState,
		Secure:   service.Config.IsPortalAPISecure(),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect the user to the authentication endpoint of the OIDC provider
	http.Redirect(writer, request, service.oidcOAuth2Config.AuthCodeURL(state.ID, oidc.Nonce(state.Nonce)), http.StatusFound)
}

// EndpointOIDCLoginCallback handles the 'GET /v1/auth/oidc/callback' endpoint
func (service *Service) EndpointOIDCLoginCallback(writer http.ResponseWriter, request *http.Request) {
	// Extract the state cookie
	stateCookie, err := request.Cookie(cookieNameState)
	if err != nil {
		service.error(writer, http.StatusBadRequest, "no login flow initiated")
		return
	}
	stateJSON, err := base64.StdEncoding.DecodeString(stateCookie.Value)
	if err != nil {
		service.error(writer, http.StatusBadRequest, "invalid state cookie")
		return
	}
	state := new(oidcLoginFlowState)
	if err := json.Unmarshal(stateJSON, state); err != nil {
		service.error(writer, http.StatusBadRequest, "invalid state cookie")
		return
	}

	// Validate the state ID
	if request.URL.Query().Get("state") != state.ID {
		service.error(writer, http.StatusBadRequest, "states do not match")
		return
	}

	// Unset the state cookie
	http.SetCookie(writer, &http.Cookie{
		Name:     cookieNameState,
		Value:    "",
		Expires:  time.Now().Add(-time.Second),
		HttpOnly: true,
	})

	// Retrieve the OAuth2 access token and extract and verify the ID token + nonce
	oauth2Token, err := service.oidcOAuth2Config.Exchange(request.Context(), request.URL.Query().Get("code"))
	if err != nil {
		service.error(writer, http.StatusForbidden, "invalid login code (expired?)")
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		service.internalError(writer, errors.New("no 'id_token' field in OAuth2 access token; most likely an OIDC provider error"))
		return
	}
	idToken, err := service.oidcIDTokenVerifier.Verify(request.Context(), rawIDToken)
	if err != nil {
		service.internalError(writer, errors.New("received invalid ID token; most likely an OIDC provider error"))
		return
	}
	if idToken.Nonce != state.Nonce {
		service.error(writer, http.StatusForbidden, "nonces do not match")
		return
	}

	// Set the ID token cookie
	http.SetCookie(writer, &http.Cookie{
		Name:     cookieNameToken,
		Value:    rawIDToken,
		Secure:   service.Config.IsPortalAPISecure(),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Redirect the user to the URL specified on login flow initiating
	http.Redirect(writer, request, state.Afterwards, http.StatusFound)
}

// EndpointOIDCBackchannelLogout handles the 'POST /v1/auth/oidc/backchannel_logout' endpoint
func (service *Service) EndpointOIDCBackchannelLogout(writer http.ResponseWriter, request *http.Request) {
	// TODO: implement backchannel logout logic
	service.error(writer, http.StatusNotImplemented, "not implemented yet")
}
