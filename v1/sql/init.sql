-- CREATE DATABASE IF NOT EXISTS energy_recovery DEFAULT CHARSET utf8mb4;
-- USE energy_recovery;

-- CREATE TABLE users (
--     id INT AUTO_INCREMENT PRIMARY KEY,
--     device_id VARCHAR(64) NOT NULL UNIQUE,
--     height FLOAT,
--     weight FLOAT,
--     body_fat FLOAT,
--     age INT,
--     symptoms TEXT,
--     diagnosis_type VARCHAR(32),
--     diagnosis_label VARCHAR(64),
--     micro_death_index INT,
--     micro_death_level VARCHAR(16),
--     vitality_score INT DEFAULT 44,
--     current_stage INT DEFAULT 1,
--     stage_start_date DATE,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
-- ) ENGINE=InnoDB;

-- CREATE TABLE task_completions (
--     id INT AUTO_INCREMENT PRIMARY KEY,
--     device_id VARCHAR(64) NOT NULL,
--     task_date DATE NOT NULL,
--     task_id VARCHAR(32) NOT NULL,
--     completed TINYINT(1) DEFAULT 0,
--     UNIQUE KEY unique_task (device_id, task_date, task_id),
--     INDEX idx_device_date (device_id, task_date)
-- ) ENGINE=InnoDB;

-- CREATE TABLE checkins (
--     id INT AUTO_INCREMENT PRIMARY KEY,
--     device_id VARCHAR(64) NOT NULL,
--     checkin_date DATE NOT NULL,
--     UNIQUE KEY unique_checkin (device_id, checkin_date)
-- ) ENGINE=InnoDB;

-- CREATE TABLE stage_progress (
--     id INT AUTO_INCREMENT PRIMARY KEY,
--     device_id VARCHAR(64) NOT NULL,
--     stage_num INT NOT NULL,
--     days_completed INT DEFAULT 0,
--     UNIQUE KEY unique_stage (device_id, stage_num)
-- ) ENGINE=InnoDB;

-- CREATE TABLE visits (
--     id INT AUTO_INCREMENT PRIMARY KEY,
--     ip VARCHAR(45) NOT NULL,
--     device_id VARCHAR(64),
--     visit_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     user_agent TEXT
-- ) ENGINE=InnoDB;


CREATE DATABASE IF NOT EXISTS energy_recovery DEFAULT CHARSET utf8mb4;
USE energy_recovery;

-- 账号表
CREATE TABLE accounts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 用户健康档案（关联账号）
CREATE TABLE user_profile (
    id INT AUTO_INCREMENT PRIMARY KEY,
    account_id INT NOT NULL UNIQUE,
    height FLOAT,
    weight FLOAT,
    body_fat FLOAT,
    age INT,
    symptoms TEXT,
    diagnosis_type VARCHAR(32),
    diagnosis_label VARCHAR(64),
    micro_death_index INT,
    micro_death_level VARCHAR(16),
    vitality_score INT DEFAULT 44,
    current_stage INT DEFAULT 1,
    stage_start_date DATE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 任务完成记录
CREATE TABLE task_completions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    account_id INT NOT NULL,
    task_date DATE NOT NULL,
    task_id VARCHAR(32) NOT NULL,
    completed TINYINT(1) DEFAULT 0,
    UNIQUE KEY unique_task (account_id, task_date, task_id),
    INDEX idx_account_date (account_id, task_date),
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 打卡记录
CREATE TABLE checkins (
    id INT AUTO_INCREMENT PRIMARY KEY,
    account_id INT NOT NULL,
    checkin_date DATE NOT NULL,
    UNIQUE KEY unique_checkin (account_id, checkin_date),
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 阶段进度
CREATE TABLE stage_progress (
    id INT AUTO_INCREMENT PRIMARY KEY,
    account_id INT NOT NULL,
    stage_num INT NOT NULL,
    days_completed INT DEFAULT 0,
    UNIQUE KEY unique_stage (account_id, stage_num),
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 访问记录（保留IP，但不强制账号）
CREATE TABLE visits (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ip VARCHAR(45) NOT NULL,
    account_id INT DEFAULT NULL,
    visit_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_agent TEXT,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE SET NULL
) ENGINE=InnoDB;