-- name: SavePushSubscription :exec
INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh, auth, created_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(endpoint) DO UPDATE SET
    user_id = excluded.user_id,
    p256dh = excluded.p256dh,
    auth = excluded.auth,
    created_at = excluded.created_at;

-- name: GetPushSubscriptionsByUserID :many
SELECT id, user_id, endpoint, p256dh, auth, created_at FROM push_subscriptions WHERE user_id = ?;

-- name: DeletePushSubscription :exec
DELETE FROM push_subscriptions WHERE user_id = ? AND endpoint = ?;

-- name: DeleteStalePushSubscription :exec
DELETE FROM push_subscriptions WHERE endpoint = ?;

-- name: GetPushSubscriptionsForSession :many
SELECT
    ps.id,
    ps.user_id,
    ps.endpoint,
    ps.p256dh,
    ps.auth
FROM push_subscriptions ps
JOIN players p ON p.user_id = ps.user_id
WHERE p.session_id = ?
GROUP BY ps.id;
