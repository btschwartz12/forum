// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: posts.sql

package db

import (
	"context"
	"database/sql"
)

const deletePost = `-- name: DeletePost :exec
DELETE FROM
    posts
WHERE
    id = ?
`

func (q *Queries) DeletePost(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deletePost, id)
	return err
}

const deletePostsForUser = `-- name: DeletePostsForUser :exec
DELETE FROM
    posts
WHERE
    author_id = ?
`

func (q *Queries) DeletePostsForUser(ctx context.Context, authorID int64) error {
	_, err := q.db.ExecContext(ctx, deletePostsForUser, authorID)
	return err
}

const getAllPosts = `-- name: GetAllPosts :many
SELECT
    posts.id AS post_id,
    posts.content AS post_content,
    posts.timestamp AS post_timestamp,
    posts.ip AS post_ip,
    posts.image_filename AS post_image_filename,
    users.id AS user_id,
    users.username AS user_username,
    users.password AS user_password,
    users.is_admin AS user_is_admin
FROM
    posts
JOIN
    users ON posts.author_id = users.id
ORDER BY
    posts.timestamp DESC
`

type GetAllPostsRow struct {
	PostID            int64
	PostContent       string
	PostTimestamp     string
	PostIp            string
	PostImageFilename sql.NullString
	UserID            int64
	UserUsername      string
	UserPassword      string
	UserIsAdmin       int64
}

func (q *Queries) GetAllPosts(ctx context.Context) ([]GetAllPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllPosts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllPostsRow
	for rows.Next() {
		var i GetAllPostsRow
		if err := rows.Scan(
			&i.PostID,
			&i.PostContent,
			&i.PostTimestamp,
			&i.PostIp,
			&i.PostImageFilename,
			&i.UserID,
			&i.UserUsername,
			&i.UserPassword,
			&i.UserIsAdmin,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getImageFilenameForPost = `-- name: GetImageFilenameForPost :one
SELECT
    image_filename
FROM
    posts
WHERE
    id = ?
`

func (q *Queries) GetImageFilenameForPost(ctx context.Context, id int64) (sql.NullString, error) {
	row := q.db.QueryRowContext(ctx, getImageFilenameForPost, id)
	var image_filename sql.NullString
	err := row.Scan(&image_filename)
	return image_filename, err
}

const getPostAuthor = `-- name: GetPostAuthor :one
SELECT
    user.id, user.username, user.password, user.is_admin
FROM
    posts
JOIN
    users AS user ON posts.author_id = user.id
WHERE
    posts.id = ?
`

func (q *Queries) GetPostAuthor(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRowContext(ctx, getPostAuthor, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.IsAdmin,
	)
	return i, err
}

const insertPost = `-- name: InsertPost :one
INSERT INTO
    posts (author_id, content, timestamp, ip, image_filename)
VALUES
    (?, ?, ?, ?, ?)
RETURNING
    id
`

type InsertPostParams struct {
	AuthorID      int64
	Content       string
	Timestamp     string
	Ip            string
	ImageFilename sql.NullString
}

func (q *Queries) InsertPost(ctx context.Context, arg InsertPostParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, insertPost,
		arg.AuthorID,
		arg.Content,
		arg.Timestamp,
		arg.Ip,
		arg.ImageFilename,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const updatePostContent = `-- name: UpdatePostContent :exec
UPDATE
    posts
SET
    content = ?,
    ip = ?
WHERE
    id = ?
`

type UpdatePostContentParams struct {
	Content string
	Ip      string
	ID      int64
}

func (q *Queries) UpdatePostContent(ctx context.Context, arg UpdatePostContentParams) error {
	_, err := q.db.ExecContext(ctx, updatePostContent, arg.Content, arg.Ip, arg.ID)
	return err
}
