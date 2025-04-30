package dto

import "fem/internal/store"

/**
* 💡 using *pointers below to catch 0-values
* to ensure fields for `var`s of this DTO type have no 0-values
* if a field is not set, its value would be `nil` instead of `""` or `0` (for example)
 */
type UpdateWorkoutDTO struct {
	Title           *string              `json:"title"`
	Description     *string              `json:"description"`
	DurationMinutes *int                 `json:"duration_minutes"`
	CaloriesBurned  *int                 `json:"calories_burned"`
	Entries         []store.WorkoutEntry `json:"entries"`
}
