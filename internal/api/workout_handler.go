package api

import (
	"database/sql"
	"encoding/json"
	"fem/internal/store"
	"fem/internal/utils"
	"log"
	"net/http"
)

type WorkoutHandler struct {
	workoutStore store.WorkoutStore
	logger       *log.Logger
}

func NewWorkoutHandler(workoutStore store.WorkoutStore, logger *log.Logger) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore: workoutStore,
		logger:       logger,
	}
}

func (wh *WorkoutHandler) HandleGetWorkoutByID(w http.ResponseWriter, r *http.Request) {
	workoutID, err := utils.ReadIDParam(r)
	if err != nil {
		wh.logger.Printf("ERROR: ReadIDParam %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid workout id"})
		return
	}

	workout, err := wh.workoutStore.GetWorkoutByID(workoutID)
	if err != nil {
		wh.logger.Printf("ERROR: GetWorkoutByID %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workout": workout})
}

func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	var workout store.Workout

	err := json.NewDecoder(r.Body).Decode(&workout)
	if err != nil {
		wh.logger.Printf("ERROR: decoding create workout %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request"})
		return
	}

	createdWorkout, err := wh.workoutStore.CreateWorkout(&workout)
	if err != nil {
		wh.logger.Printf("ERROR: CreateWorkout %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "invalid request"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"workout": createdWorkout})
}

func (wh *WorkoutHandler) HandleUpdateWorkoutByID(w http.ResponseWriter, r *http.Request) {
	workoutID, err := utils.ReadIDParam(r)
	if err != nil {
		wh.logger.Printf("ERROR: ReadIDParam %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid update workout id"})
		return
	}

	existingWorkout, err := wh.workoutStore.GetWorkoutByID(workoutID)
	if err != nil {
		wh.logger.Printf("ERROR: GetWorkoutByID %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	if existingWorkout == nil {
		http.NotFound(w, r)
		return
	}

	var updateWorkoutDTO struct {
		Title           *string              `json:"title"`
		Description     *string              `json:"description"`
		DurationMinutes *int                 `json:"duration_minutes"`
		CaloriesBurned  *int                 `json:"calories_burned"`
		Entries         []store.WorkoutEntry `json:"entries"`
	}

	err = json.NewDecoder(r.Body).Decode(&updateWorkoutDTO)
	if err != nil {
		wh.logger.Printf("ERROR: decoding update request body %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request"})
		return
	}

	if updateWorkoutDTO.Title != nil {
		// 💡 using *pointers here to check for 0-values
		// iow, to ensure `updateWorkoutDTO` has an actual value in those fields
		existingWorkout.Title = *updateWorkoutDTO.Title
	}
	if updateWorkoutDTO.Description != nil {
		existingWorkout.Description = *updateWorkoutDTO.Description
	}
	if updateWorkoutDTO.DurationMinutes != nil {
		existingWorkout.DurationMinutes = *updateWorkoutDTO.DurationMinutes
	}
	if updateWorkoutDTO.CaloriesBurned != nil {
		existingWorkout.CaloriesBurned = *updateWorkoutDTO.CaloriesBurned
	}
	if updateWorkoutDTO.Entries != nil {
		existingWorkout.Entries = updateWorkoutDTO.Entries
	}

	err = wh.workoutStore.UpdateWorkout(existingWorkout)
	if err != nil {
		wh.logger.Printf("ERROR: UpdateWorkout %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workout": existingWorkout})
}

func (wh *WorkoutHandler) HandleDeleteWorkoutByID(w http.ResponseWriter, r *http.Request) {
	workoutID, err := utils.ReadIDParam(r)
	if err != nil {
		wh.logger.Printf("ERROR: ReadIDParam %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid delete workout id"})
		return
	}

	err = wh.workoutStore.DeleteWorkoutByID(workoutID)
	if err == sql.ErrNoRows {
		wh.logger.Printf("ERROR: workout not found %v", err)
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "workout not found"})
		return
	}

	if err != nil {
		wh.logger.Printf("ERROR: DeleteWorkoutByID %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
