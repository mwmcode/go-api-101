package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fem/internal/store"
	"fem/internal/tokens"
	"fem/internal/utils"
	"net/http"
	"strings"
)

type UserMiddleware struct {
	UserStore    store.UserStore
	WorkoutStore store.WorkoutStore
}

type contextKey string

const UserContextKey = contextKey("user")

func SetUser(r *http.Request, user *store.User) *http.Request {
	ctx := context.WithValue(r.Context(), UserContextKey, user)
	return r.WithContext(ctx)
}

func GetUser(r *http.Request) *store.User {
	user, ok := r.Context().Value(UserContextKey).(*store.User)
	// .(*store.User) ensure the result is a point to a User
	// otherwise `ok` will be `false`

	if !ok {
		panic("missing user in request")
	}

	return user
}

func (um *UserMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			r = SetUser(r, store.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authHeader, " ") // Bearer <TOKEN>
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
				"error": "invalid authorization header",
			})
			return
		}

		token := headerParts[1]
		user, err := um.UserStore.GetUserByToken(tokens.ScopeAuth, token)

		if err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
				"error": "invalid token",
			})
			return
		}

		if user == nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
				"error": "token expired or invalid",
			})
			return
		}

		r = SetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

// @note: not used
func (um *UserMiddleware) RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)

		if user.IsAnonymous() {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
				"error": "unauthorized",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (um *UserMiddleware) IsAuthed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)

		if user.IsAnonymous() {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
				"error": "unauthorized",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (um *UserMiddleware) CanModifyWorkouts(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionUserID := GetUser(r).ID
		workoutID, _ := utils.ReadIDParam(r)
		workoutOwnerID, err := um.WorkoutStore.GetWorkoutOwnerID(workoutID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "workout not found"})
				return
			}
			utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
			return
		}
		if sessionUserID != workoutOwnerID {
			utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "unauthorized"})
			return
		}

		next.ServeHTTP(w, r)
	})
}
