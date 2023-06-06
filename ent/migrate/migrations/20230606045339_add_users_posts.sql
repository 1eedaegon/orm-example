-- Modify "users" table
ALTER TABLE `users` MODIFY COLUMN `name` varchar(255) NOT NULL, ADD UNIQUE INDEX `name` (`name`);
