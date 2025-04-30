package store

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable")
	if err != nil {
		t.Fatalf("error connecting to test db: %v", err)
	}
	err = Migrate(db, "../../migrations")
	if err != nil {
		t.Fatalf("error migrating test db: %v", err)
	}

	_, err = db.Exec(`TRUNCATE workouts, workout_entries CASCADE`)
	if err != nil {
		t.Fatalf("error migrating test db: %v", err)
	}

	return db
}

func TestCreateWorkout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewPostgresWorkoutStore(db)

	tests := []struct {
		name    string
		workout *Workout
		wantErr bool
	}{
		{
			name:    "valid workout",
			wantErr: false,
			workout: &Workout{
				Title:           "push",
				Description:     "Upper body day",
				DurationMinutes: 60,
				CaloriesBurned:  200,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Bench press",
						Sets:         3,
						Reps:         IntPtr(10),
						Weight:       FlotPtr(130),
						Notes:        "do it ✔️",
						OrderIndex:   1,
					},
				},
			},
		},
		{
			name:    "invalid workout",
			wantErr: true,
			workout: &Workout{
				Title:           "full",
				Description:     "complete workout",
				DurationMinutes: 60,
				CaloriesBurned:  200,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Plank",
						Sets:         3,
						Reps:         IntPtr(3),
						Notes:        "do it ✔️",
						OrderIndex:   1,
					},
					{
						ExerciseName:    "Squats",
						Sets:            4,
						Reps:            IntPtr(8),
						DurationSeconds: IntPtr(90),
						Notes:           "do it ✔️",
						OrderIndex:      2,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdWorkout, err := store.CreateWorkout(tt.workout)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.workout.Title, createdWorkout.Title)
			assert.Equal(t, tt.workout.Description, createdWorkout.Description)
			assert.Equal(t, tt.workout.DurationMinutes, createdWorkout.DurationMinutes)

			dbWorkout, err := store.GetWorkoutByID(int64(createdWorkout.ID))
			require.NoError(t, err)

			assert.Equal(t, dbWorkout.ID, createdWorkout.ID)
			assert.Equal(t, len(dbWorkout.Entries), len(createdWorkout.Entries))

			for i, entry := range dbWorkout.Entries {
				assert.Equal(t,
					entry.ExerciseName,
					tt.workout.Entries[i].ExerciseName,
				)
				assert.Equal(t,
					entry.Sets,
					tt.workout.Entries[i].Sets,
				)
				assert.Equal(t,
					entry.OrderIndex,
					tt.workout.Entries[i].OrderIndex,
				)
			}
		})
	}
}

func IntPtr(i int) *int {
	return &i
}

func FlotPtr(f float64) *float64 {
	return &f
}
