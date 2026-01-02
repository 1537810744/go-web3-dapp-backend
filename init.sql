CREATE DATABASE IF NOT EXISTS web3_dapp;
USE web3_dapp;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password VARCHAR(128) NOT NULL,
    eth_address VARCHAR(42) NOT NULL UNIQUE,
    eth_private_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    tx_hash VARCHAR(66) NOT NULL,
    event_type VARCHAR(16) NOT NULL,
    user_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42),
    amount DECIMAL(65,0) NOT NULL,
    block_number BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sync_status (
    id INT PRIMARY KEY,
    last_block BIGINT
);
INSERT INTO sync_status (id, last_block) VALUES (1, 0)
    ON DUPLICATE KEY UPDATE id=id;
