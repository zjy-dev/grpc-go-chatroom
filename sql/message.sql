use `grpc_go_chatroom`;

CREATE TABLE `chat_message` (
    `id` int NOT NULL AUTO_INCREMENT,
    `user_id` int NOT NULL,
    `username` varchar(255) NOT NULL,
    `message` text NOT NULL,
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB AUTO_INCREMENT = 340 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci
-- FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE