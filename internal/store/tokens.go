package store

import (
	"database/sql"
	"time"

	"fem/internal/tokens"
)

type PostgresTokenStore struct {
	db *sql.DB
}

func NewPostgresTokenStore(db *sql.DB) *PostgresTokenStore {
	return &PostgresTokenStore{
		db: db,
	}
}

type TokenStore interface {
	CreateNewToken(userID int, ttl time.Duration, scope string) (*tokens.Token, error)
	DeleteUserTokens(userID int, scope string) error
}

func (pgStore *PostgresTokenStore) CreateNewToken(userID int, ttl time.Duration, scope string) (*tokens.Token, error) {
	token, err := tokens.GenerateToken(userID, ttl, scope)

	if err != nil {
		return nil, err
	}
	_, err = pgStore.db.Exec(
		`INSERT INTO tokens (hash, user_id, expiry, scope) VALUES ($1, $2, $3, $4)`,
		token.Hash,
		token.UserID,
		token.Expiry,
		token.Scope,
	)

	return token, err
}

func (pgStore *PostgresTokenStore) DeleteUserTokens(userID int, scope string) error {
	_, err := pgStore.db.Exec(`DELETE FROM tokens WHERE scope = $1 AND user_id = $2`)

	return err
}
