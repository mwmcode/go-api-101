package api

import (
	"encoding/json"
	"fem/internal/dto"
	"fem/internal/store"
	"fem/internal/tokens"
	"fem/internal/utils"
	"log"
	"net/http"
	"time"
)

type TokenHandler struct {
	tokenStore store.TokenStore
	userStore  store.UserStore
	logger     *log.Logger
}

func NewTokenHandler(tokenStore store.TokenStore, userStore store.UserStore, logger *log.Logger) *TokenHandler {
	return &TokenHandler{
		tokenStore: tokenStore,
		userStore:  userStore,
		logger:     logger,
	}
}

func (h *TokenHandler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTokenDTO

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Printf("error creating token %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "invalid request payload",
		})
		return
	}

	user, err := h.userStore.GetUserByUsername(req.Username)
	if err != nil || user == nil {
		h.logger.Printf("error GetUserByUsername %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	passwordsMatch, err := user.PasswordHash.Matches(req.Password)
	if err != nil {
		h.logger.Printf("error PasswordHash.Matches %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	if !passwordsMatch {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
			"error": "invalid login attempt",
		})
		return
	}

	token, err := h.tokenStore.CreateNewToken(user.ID, 24*time.Hour, tokens.ScopeAuth)
	if err != nil {
		h.logger.Printf("error CreateNewToken %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"auth_token": token})
}
