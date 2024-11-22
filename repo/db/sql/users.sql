-- name: InsertUser :one
INSERT INTO
    users (username, password, is_admin)
VALUES
    (?, ?, ?)
RETURNING
    id;

-- name: UpdateUser :exec
UPDATE
    users
SET
    username = ?,
    password = ?,
    is_admin = ?
WHERE
    id = ?;

-- name: GetAllUsers :many
SELECT
    *
FROM
    users;

-- name: GetUserByUsername :one
SELECT
    *
FROM
    users
WHERE
    username = ?;

-- name: UsernameExists :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            users
        WHERE
            username = ?
    );

-- name: DeleteUser :exec
DELETE FROM
    users
WHERE
    id = ?;
