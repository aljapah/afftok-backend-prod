CREATE TABLE `afftok_users` (
	`id` varchar(36) NOT NULL,
	`username` varchar(50) NOT NULL,
	`email` varchar(255) NOT NULL,
	`full_name` varchar(100),
	`avatar_url` text,
	`role` enum('user','admin') NOT NULL DEFAULT 'user',
	`status` enum('active','suspended') NOT NULL DEFAULT 'active',
	`points` int NOT NULL DEFAULT 0,
	`level` int NOT NULL DEFAULT 1,
	`total_clicks` int NOT NULL DEFAULT 0,
	`total_conversions` int NOT NULL DEFAULT 0,
	`total_earnings` int NOT NULL DEFAULT 0,
	`created_at` timestamp NOT NULL DEFAULT (now()),
	`updated_at` timestamp NOT NULL DEFAULT (now()) ON UPDATE CURRENT_TIMESTAMP,
	CONSTRAINT `afftok_users_id` PRIMARY KEY(`id`),
	CONSTRAINT `afftok_users_username_unique` UNIQUE(`username`),
	CONSTRAINT `afftok_users_email_unique` UNIQUE(`email`)
);
--> statement-breakpoint
CREATE TABLE `badges` (
	`id` varchar(36) NOT NULL,
	`name` varchar(100) NOT NULL,
	`description` text,
	`icon_url` text,
	`criteria` text NOT NULL,
	`points_reward` int NOT NULL DEFAULT 0,
	`created_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `badges_id` PRIMARY KEY(`id`)
);
--> statement-breakpoint
CREATE TABLE `clicks` (
	`id` varchar(36) NOT NULL,
	`user_offer_id` varchar(36) NOT NULL,
	`ip_address` varchar(45),
	`user_agent` text,
	`device` varchar(50),
	`browser` varchar(50),
	`os` varchar(50),
	`country` varchar(2),
	`clicked_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `clicks_id` PRIMARY KEY(`id`)
);
--> statement-breakpoint
CREATE TABLE `conversions` (
	`id` varchar(36) NOT NULL,
	`user_offer_id` varchar(36) NOT NULL,
	`click_id` varchar(36),
	`amount` int NOT NULL DEFAULT 0,
	`commission` int NOT NULL DEFAULT 0,
	`status` enum('pending','approved','rejected') NOT NULL DEFAULT 'pending',
	`converted_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `conversions_id` PRIMARY KEY(`id`)
);
--> statement-breakpoint
CREATE TABLE `networks` (
	`id` varchar(36) NOT NULL,
	`name` varchar(100) NOT NULL,
	`api_url` text,
	`api_key` text,
	`postback_url` text,
	`hmac_secret` text,
	`status` enum('active','inactive') NOT NULL DEFAULT 'active',
	`created_at` timestamp NOT NULL DEFAULT (now()),
	`updated_at` timestamp NOT NULL DEFAULT (now()) ON UPDATE CURRENT_TIMESTAMP,
	CONSTRAINT `networks_id` PRIMARY KEY(`id`)
);
--> statement-breakpoint
CREATE TABLE `offers` (
	`id` varchar(36) NOT NULL,
	`network_id` varchar(36),
	`title` varchar(255) NOT NULL,
	`description` text,
	`image_url` text,
	`destination_url` text NOT NULL,
	`category` varchar(50),
	`payout` int NOT NULL DEFAULT 0,
	`commission` int NOT NULL DEFAULT 0,
	`status` enum('active','inactive','pending') NOT NULL DEFAULT 'pending',
	`total_clicks` int NOT NULL DEFAULT 0,
	`total_conversions` int NOT NULL DEFAULT 0,
	`created_at` timestamp NOT NULL DEFAULT (now()),
	`updated_at` timestamp NOT NULL DEFAULT (now()) ON UPDATE CURRENT_TIMESTAMP,
	CONSTRAINT `offers_id` PRIMARY KEY(`id`)
);
--> statement-breakpoint
CREATE TABLE `team_members` (
	`id` varchar(36) NOT NULL,
	`team_id` varchar(36) NOT NULL,
	`user_id` varchar(36) NOT NULL,
	`role` enum('leader','member') NOT NULL DEFAULT 'member',
	`points` int NOT NULL DEFAULT 0,
	`joined_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `team_members_id` PRIMARY KEY(`id`)
);
--> statement-breakpoint
CREATE TABLE `teams` (
	`id` varchar(36) NOT NULL,
	`name` varchar(100) NOT NULL,
	`description` text,
	`leader_id` varchar(36) NOT NULL,
	`total_points` int NOT NULL DEFAULT 0,
	`member_count` int NOT NULL DEFAULT 1,
	`created_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `teams_id` PRIMARY KEY(`id`)
);
--> statement-breakpoint
CREATE TABLE `user_badges` (
	`id` varchar(36) NOT NULL,
	`user_id` varchar(36) NOT NULL,
	`badge_id` varchar(36) NOT NULL,
	`earned_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `user_badges_id` PRIMARY KEY(`id`)
);
--> statement-breakpoint
CREATE TABLE `user_offers` (
	`id` varchar(36) NOT NULL,
	`user_id` varchar(36) NOT NULL,
	`offer_id` varchar(36) NOT NULL,
	`affiliate_link` text NOT NULL,
	`clicks` int NOT NULL DEFAULT 0,
	`conversions` int NOT NULL DEFAULT 0,
	`earnings` int NOT NULL DEFAULT 0,
	`joined_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `user_offers_id` PRIMARY KEY(`id`)
);
