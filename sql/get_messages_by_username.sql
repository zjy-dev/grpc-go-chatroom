SELECT
    id,
    user_id,
    username,
    message,
    created_at
FROM chat_messages
WHERE
    username = ?;