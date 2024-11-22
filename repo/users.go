package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/btschwartz12/forum/repo/db"
)

var (
	ErrUserNotFound      = fmt.Errorf("user not found")
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
)

func (r *Repo) InsertUser(ctx context.Context, user *User) (int64, error) {
	q := db.New(r.db)
	row, err := q.UsernameExists(ctx, user.Username)
	if err != nil {
		return 0, fmt.Errorf("error checking if username exists: %w", err)
	}
	usernameExists := row == 1
	if usernameExists {
		return 0, ErrUserAlreadyExists
	}
	id, err := q.InsertUser(ctx, user.toDb())
	if err != nil {
		return 0, fmt.Errorf("error inserting user: %w", err)
	}
	return id, nil
}

func (r *Repo) UpdateUser(ctx context.Context, user *User) error {
	q := db.New(r.db)
	err := q.UpdateUser(ctx, db.UpdateUserParams{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		IsAdmin:  boolToInt(user.IsAdmin),
	})
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}

func (r *Repo) GetAllUsers(ctx context.Context) ([]User, error) {
	q := db.New(r.db)
	rows, err := q.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	users := make([]User, 0, len(rows))
	for _, row := range rows {
		user := User{}
		user.fromDb(&row)
		users = append(users, user)
	}
	return users, nil
}

func (r *Repo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	q := db.New(r.db)
	row, err := q.GetUserByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	user := User{}
	user.fromDb(&row)
	return &user, nil
}

func (r *Repo) DeleteUser(ctx context.Context, username string) error {
	q := db.New(r.db)
	user, err := q.GetUserByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return fmt.Errorf("error getting user: %w", err)
	}
	err = q.DeletePostsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("error deleting posts for user: %w", err)
	}
	err = q.DeleteUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	return nil
}
