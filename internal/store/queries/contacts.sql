-- name: AddContact :exec
INSERT INTO contacts (user_id, contact_user_id, created_at) VALUES (?, ?, ?);

-- name: RemoveContact :exec
DELETE FROM contacts WHERE user_id = ? AND contact_user_id = ?;

-- name: UserExists :one
SELECT COUNT(*) FROM users WHERE id = ?;

-- name: GetContacts :many
SELECT
    c.contact_user_id,
    u.display_name,
    u.avatar_icon,
    u.avatar_color,
    c.created_at
FROM contacts c
JOIN users u ON u.id = c.contact_user_id
WHERE c.user_id = ?
ORDER BY c.created_at DESC;

-- name: SearchUsers :many
SELECT
    u.id,
    u.email,
    u.display_name,
    u.avatar_icon,
    u.avatar_color,
    CASE WHEN c.contact_user_id IS NOT NULL THEN 1 ELSE 0 END AS is_contact
FROM users u
LEFT JOIN contacts c ON c.contact_user_id = u.id AND c.user_id = ?
WHERE u.display_name LIKE ? AND u.id != ?
ORDER BY u.display_name LIMIT ?;
