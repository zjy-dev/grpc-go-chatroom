CREATE DATABASE IF NOT EXISTS `grpc_go_chatroom`;

use `grpc_go_chatroom`;

CREATE TABLE `users` (
    `id` int NOT NULL AUTO_INCREMENT COMMENT '用户唯一标识符',
    `username` varchar(255) NOT NULL UNIQUE COMMENT '用户名，必须唯一',
    `password_hash` varchar(255) NOT NULL COMMENT '用户密码的哈希值',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '用户创建时间',
    `last_login_at` timestamp NULL DEFAULT NULL COMMENT '用户最后登录时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `messages` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT NOT NULL,
    `username` VARCHAR(255) NOT NULL,
    `message` TEXT NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);