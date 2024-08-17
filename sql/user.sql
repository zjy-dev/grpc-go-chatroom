USE `grpc_go_chatroom`;

CREATE TABLE `user` (
    `id` int NOT NULL AUTO_INCREMENT COMMENT '用户唯一标识符',
    `username` varchar(255) NOT NULL UNIQUE COMMENT '用户名，必须唯一',
    `password_hash` varchar(255) NOT NULL COMMENT '用户密码的哈希值',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '用户创建时间',
    `last_login_at` timestamp NULL DEFAULT NULL COMMENT '用户最后登录时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

INSERT INTO `user` (`username`, `password_hash`) VALUES (?, ?);