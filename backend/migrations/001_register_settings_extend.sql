-- Migration: Register settings extension
-- Date: 2026-05-05

-- 1. Add IP tracking to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS ip_registered_from VARCHAR(50);

-- 2. Extend signup_configs table with new fields
ALTER TABLE signup_configs ADD COLUMN IF NOT EXISTS allowed_domains VARCHAR(500);
ALTER TABLE signup_configs ADD COLUMN IF NOT EXISTS min_password_length INTEGER DEFAULT 8;
ALTER TABLE signup_configs ADD COLUMN IF NOT EXISTS signup_reward_type VARCHAR(20) DEFAULT 'quota';
ALTER TABLE signup_configs ADD COLUMN IF NOT EXISTS signup_reward_amount BIGINT DEFAULT 50000;

-- 3. Update existing signup config with defaults
UPDATE signup_configs 
SET allowed_domains = COALESCE(allowed_domains, ''),
    min_password_length = COALESCE(min_password_length, 8),
    signup_reward_type = COALESCE(signup_reward_type, 'quota'),
    signup_reward_amount = COALESCE(signup_reward_amount, 50000)
WHERE id = 1;