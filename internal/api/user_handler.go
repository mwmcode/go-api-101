package api

import (
	"encoding/json"
	"errors"
	"fem/internal/dto"
	"fem/internal/store"
	"fem/internal/utils"
	"log"
	"net/http"
	"regexp"
)

type UserHandler struct {
	userStore store.UserStore
	logger    *log.Logger
}

func NewUserHandler(userStore store.UserStore, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userStore: userStore,
		logger:    logger,
	}
}

func (h *UserHandler) validateRegisterRequest(userReq *dto.RegisterUserDTO) error {
	if userReq.Username == "" {
		return errors.New("username is required")
	}

	if len(userReq.Username) > 50 {
		return errors.New("username cannot be greater than 50 characters")
	}

	if userReq.Email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(userReq.Email) {
		return errors.New("invalid email format")
	}

	if userReq.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

func (h *UserHandler) HandleRegisterUser(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserDTO

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Printf("error decoding request %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "invalid request payload",
		})
		return
	}
	err = h.validateRegisterRequest(&req)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": err.Error(),
		})
		return
	}

	user := &store.User{
		Username: req.Username,
		Email:    req.Email,
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}

	err = user.PasswordHash.Set(req.Password)
	if err != nil {
		h.logger.Printf("error hasing password %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	err = h.userStore.CreateUser(user)

	if err != nil {
		h.logger.Printf("error registering user %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"user": user})
}
