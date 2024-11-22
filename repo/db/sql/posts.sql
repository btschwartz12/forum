-- name: InsertPost :one
INSERT INTO
    posts (author_id, content, timestamp, ip, image_filename)
VALUES
    (?, ?, ?, ?, ?)
RETURNING
    id;

-- name: GetImageFilenameForPost :one
SELECT
    image_filename
FROM
    posts
WHERE
    id = ?;

-- name: GetAllPosts :many
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
    posts.timestamp DESC;

-- name: DeletePost :exec
DELETE FROM
    posts
WHERE
    id = ?;

-- name: UpdatePostContent :exec
UPDATE
    posts
SET
    content = ?,
    ip = ?
WHERE
    id = ?;

-- name: GetPostAuthor :one
SELECT
    user.*
FROM
    posts
JOIN
    users AS user ON posts.author_id = user.id
WHERE
    posts.id = ?;

-- name: DeletePostsForUser :exec
DELETE FROM
    posts
WHERE
    author_id = ?;
