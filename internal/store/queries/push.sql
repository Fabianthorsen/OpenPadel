-- name: CreatePushSubscription :exec
INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh, auth, created_at) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetPushSubscriptionsByUserID :many
SELECT id, endpoint, p256dh, auth FROM push_subscriptions WHERE user_id = ?;

-- name: DeletePushSubscription :exec
DELETE FROM push_subscriptions WHERE endpoint = ? AND p256dh = ? AND auth = ?;
