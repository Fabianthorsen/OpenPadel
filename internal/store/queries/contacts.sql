-- name: AddContact :exec
INSERT INTO contacts (user_id, contact_user_id, created_at) VALUES (?, ?, ?);

-- name: RemoveContact :exec
DELETE FROM contacts WHERE user_id = ? AND contact_user_id = ?;

-- name: GetContactsByUserID :many
SELECT contact_user_id FROM contacts WHERE user_id = ? ORDER BY created_at DESC;

-- name: SearchUsers :many
SELECT id, email, display_name, avatar_icon, avatar_color FROM users
WHERE display_name LIKE ? AND id != ?
ORDER BY display_name LIMIT ?;
