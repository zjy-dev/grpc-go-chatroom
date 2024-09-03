SELECT
    id,
    user_id,
    username,
    message,
    created_at
FROM messages
WHERE
    username = ?;