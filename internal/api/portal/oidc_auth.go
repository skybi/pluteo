package portal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/skybi/data-server/internal/api/portal/session"
	"github.com/skybi/data-server/internal/api/schema"
	"github.com/skybi/data-server/internal/random"
	"github.com/skybi/data-server/internal/user"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

var (
	sessionTokenCookieName = "session_token"

	loginStateStateLength    = 16
	loginStateNonceLength    = 16
	loginStateCookieName     = "login_state"
	loginStateCookieLifetime = int(time.Hour.Seconds())

	contextValueSession = "session"
	contextValueUser    = "user"

	displayNameClaimChain = []string{
		"preferred_username",
		"nickname",
		"given_name",
		"name",
	}
)

var (
	errAuthNoLoginFlowInitiated = &schema.Error{
		Type:    "portal.auth.noLoginFlowInitiated",
		Message: "No login flow initiated.",
		Details: map[string]any{},
	}
	errAuthInvalidStateCookie = &schema.Error{
		Type:    "portal.auth.invalidStateCookie",
		Message: "Invalid state cookie.",
		Details: map[string]any{},
	}
	errAuthStatesDoNotMatch = &schema.Error{
		Type:    "portal.auth.statesDoNotMatch",
		Message: "States do not match.",
		Details: map[string]any{},
	}
	errAuthInvalidLoginCode = &schema.Error{
		Type:    "portal.auth.invalidLoginCode",
		Message: "Invalid login code. It may be expired.",
		Details: map[string]any{},
	}
	errAuthInvalidNonce = &schema.Error{
		Type:    "portal.auth.invalidNonce",
		Message: "Invalid nonce.",
		Details: map[string]any{},
	}
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
		ID:         random.String(loginStateStateLength, random.CharsetAlphanumeric),
		Nonce:      random.String(loginStateNonceLength, random.CharsetAlphanumeric),
		Afterwards: afterwards,
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	http.SetCookie(writer, &http.Cookie{
		Name:     loginStateCookieName,
		Value:    base64.StdEncoding.EncodeToString(stateJSON),
		MaxAge:   loginStateCookieLifetime,
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
	stateCookie, err := request.Cookie(loginStateCookieName)
	if err != nil {
		service.writer.WriteErrors(writer, http.StatusBadRequest, errAuthNoLoginFlowInitiated)
		return
	}
	stateJSON, err := base64.StdEncoding.DecodeString(stateCookie.Value)
	if err != nil {
		service.writer.WriteErrors(writer, http.StatusBadRequest, errAuthInvalidStateCookie)
		return
	}
	state := new(oidcLoginFlowState)
	if err := json.Unmarshal(stateJSON, state); err != nil {
		service.writer.WriteErrors(writer, http.StatusBadRequest, errAuthInvalidStateCookie)
		return
	}

	// Validate the state ID
	if request.URL.Query().Get("state") != state.ID {
		service.writer.WriteErrors(writer, http.StatusBadRequest, errAuthStatesDoNotMatch)
		return
	}

	// Unset the state cookie
	unsetCookie(writer, loginStateCookieName)

	// Retrieve the OAuth2 access token and extract and verify the ID token + nonce
	oauth2Token, err := service.oidcOAuth2Config.Exchange(request.Context(), request.URL.Query().Get("code"))
	if err != nil {
		service.writer.WriteErrors(writer, http.StatusForbidden, errAuthInvalidLoginCode)
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		service.writer.WriteInternalError(writer, errors.New("no 'id_token' field in OAuth2 access token; most likely an OIDC provider error"))
		return
	}
	idToken, err := service.oidcIDTokenVerifier.Verify(request.Context(), rawIDToken)
	if err != nil {
		service.writer.WriteInternalError(writer, errors.New("received invalid ID token; most likely an OIDC provider error"))
		return
	}
	if idToken.Nonce != state.Nonce {
		service.writer.WriteErrors(writer, http.StatusForbidden, errAuthInvalidNonce)
		return
	}

	// Extract the token claims into a map
	claims := make(map[string]any)
	if err = idToken.Claims(&claims); err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	// Extract the session ID claim if the provider provided one
	sessionID, _ := extractFirstString(claims, "sid")

	// Extract the display name to use out of the ID token claims (checking each claim out of displayNameClaimChain in order).
	// If there is no claim in the ID token that provides a usable display name, call the userinfo endpoint and use the response.
	displayName, found := extractFirstString(claims, displayNameClaimChain...)
	if !found {
		userInfo, err := service.oidcProvider.UserInfo(request.Context(), oauth2.StaticTokenSource(oauth2Token))
		if err != nil {
			service.writer.WriteInternalError(writer, err)
			return
		}

		userInfoClaims := make(map[string]any)
		if err := userInfo.Claims(&userInfoClaims); err != nil {
			service.writer.WriteInternalError(writer, err)
			return
		}

		displayName, found = extractFirstString(userInfoClaims, displayNameClaimChain...)
	}
	if !found {
		displayName = sessionID
		if displayName == "" {
			displayName = "banana"
		}
	}

	// Initialize or update the user's database entry
	userObj, err := service.Storage.Users().GetByID(request.Context(), idToken.Subject)
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}
	if userObj == nil {
		userObj, err = service.Storage.Users().Create(request.Context(), &user.Create{
			ID:           idToken.Subject,
			DisplayName:  displayName,
			APIKeyPolicy: user.DefaultAPIKeyPolicy(),
			Admin:        false,
		})
		if err != nil {
			service.writer.WriteInternalError(writer, err)
			return
		}
	}
	if userObj.DisplayName != displayName {
		userObj, err = service.Storage.Users().Update(request.Context(), userObj.ID, &user.Update{
			DisplayName: &displayName,
		})
		if err != nil {
			service.writer.WriteInternalError(writer, err)
			return
		}
	}

	// Create a new session for the user
	sessionToken, err := service.sessionStorage.Create(request.Context(), idToken.Subject, sessionID, idToken.Expiry.Unix())
	if err != nil {
		service.writer.WriteInternalError(writer, err)
		return
	}

	// Set the session token cookie
	http.SetCookie(writer, &http.Cookie{
		Name:     sessionTokenCookieName,
		Value:    sessionToken,
		Path:     "/",
		Expires:  idToken.Expiry,
		Secure:   service.Config.IsPortalAPISecure(),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Redirect the user to the URL specified on login flow initiating
	http.Redirect(writer, request, state.Afterwards, http.StatusFound)
}

// MiddlewareVerifySession makes sure that the requesting client has provided a valid session token.
// Additionally, it injects the session object itself into the request context.
func (service *Service) MiddlewareVerifySession(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Retrieve the session token cookie
		cookie, err := request.Cookie(sessionTokenCookieName)
		if err != nil {
			service.writer.WriteErrors(writer, http.StatusUnauthorized, schema.ErrUnauthorized)
			return
		}

		// Retrieve the session itself and validate its expiration time
		ses, err := service.sessionStorage.GetByRawToken(request.Context(), cookie.Value)
		if err != nil {
			service.writer.WriteInternalError(writer, err)
			return
		}
		if ses == nil || ses.Expires <= time.Now().Unix() {
			unsetCookie(writer, sessionTokenCookieName)
			service.writer.WriteErrors(writer, http.StatusUnauthorized, schema.ErrUnauthorized)
			return
		}

		// Delegate to the next handler
		request = request.WithContext(context.WithValue(request.Context(), contextValueSession, ses))
		next(writer, request)
	}
}

// MiddlewareFetchUser fetches the authenticated user object and injects it into the request context
func (service *Service) MiddlewareFetchUser(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Extract the session
		ses, ok := request.Context().Value("session").(*session.Session)
		if !ok {
			service.writer.WriteInternalError(writer, errors.New("user initialization without session initialization"))
			return
		}

		// Fetch the user object
		userObj, err := service.Storage.Users().GetByID(request.Context(), ses.UserID)
		if err != nil {
			service.writer.WriteInternalError(writer, err)
			return
		}
		if userObj == nil {
			service.writer.WriteInternalError(writer, errors.New("valid session without user object"))
			return
		}

		// Delegate to the next handler
		request = request.WithContext(context.WithValue(request.Context(), contextValueUser, userObj))
		next(writer, request)
	}
}

// MiddlewareCheckAdmin validates that the requesting client is an admin
func (service *Service) MiddlewareCheckAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userObj, ok := request.Context().Value(contextValueUser).(*user.User)
		if !ok {
			service.writer.WriteInternalError(writer, errors.New("admin check without user validation"))
			return
		}
		if !userObj.Admin {
			service.writer.WriteErrors(writer, http.StatusForbidden, schema.ErrForbidden)
			return
		}
		next(writer, request)
	}
}

func unsetCookie(writer http.ResponseWriter, name string) {
	http.SetCookie(writer, &http.Cookie{
		Name:   name,
		MaxAge: -1,
	})
}

func extractFirstString(vals map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if raw, ok := vals[key]; ok {
			if val, ok := raw.(string); ok {
				return val, true
			}
		}
	}
	return "", false
}
