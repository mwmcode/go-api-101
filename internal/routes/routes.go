package routes

import (
	"fem/internal/app"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(app.Middleware.Authenticate)
		r.Use(app.Middleware.IsAuthed)

		r.Post("/workouts", app.WorkoutHandler.HandleCreateWorkout)
		r.Put("/workouts/{id}", app.Middleware.CanModifyWorkouts(app.WorkoutHandler.HandleUpdateWorkoutByID))
		r.Delete("/workouts/{id}", app.Middleware.CanModifyWorkouts(app.WorkoutHandler.HandleDeleteWorkoutByID))
	})

	r.Get("/health", app.HealthCheck)
	r.Get("/workouts/{id}", app.WorkoutHandler.HandleGetWorkoutByID)
	r.Post("/users", app.UserHandler.HandleRegisterUser)
	r.Post("/tokens/auth", app.TokenHandler.HandleCreateToken)

	return r
}
