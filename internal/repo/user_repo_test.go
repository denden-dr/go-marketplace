package repo

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/denden-dr/go-shop-yourself/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	uID := uuid.New()
	now := time.Now()

	u := &domain.User{ID: uID, CreatedAt: now}
	auth := &domain.UserAuth{UserID: uID, Email: "test@example.com", PasswordHash: "hashed_pass", CreatedAt: now}
	profile := &domain.UserProfile{UserID: uID, Username: "testuser", UpdatedAt: now}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").WithArgs(u.ID, u.CreatedAt).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO users_auth").WithArgs(auth.UserID, auth.Email, auth.PasswordHash, auth.CreatedAt).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO users_profile").WithArgs(profile.UserID, profile.Username, profile.UpdatedAt).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.CreateUser(ctx, u, auth, profile)
	assert.NoError(t, err)
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	uID := uuid.New()
	now := time.Now()
	email := "test@example.com"

	rows := sqlmock.NewRows([]string{"user_id", "email", "password_hash", "created_at"}).
		AddRow(uID, email, "hashed_pass", now)

	mock.ExpectQuery("SELECT user_id, email, password_hash, created_at FROM users_auth").
		WithArgs(email).
		WillReturnRows(rows)

	auth, err := repo.GetUserByEmail(ctx, email)
	assert.NoError(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, uID, auth.UserID)
	assert.Equal(t, email, auth.Email)
}
