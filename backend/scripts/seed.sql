-- gAPI Platform Data Seeding Script
-- Run: docker exec gapi-postgres psql -U gapi -d gapi -f /seed.sql
-- 
-- Note: Admin credentials are in config.yaml, not database.
-- Default: admin / admin123

BEGIN;

-- Create test users
INSERT INTO users (id, username, email, password_hash, level, remain_quota, status, created_at) VALUES
(1, 'testuser1', 'test1@test.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.Y5RjlVRNQFMpBJ7qP', 'vip', 100000, 'active', NOW() - INTERVAL '30 days'),
(2, 'testuser2', 'test2@test.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.Y5RjlVRNQFMpBJ7P', 'vip', 50000, 'active', NOW() - INTERVAL '25 days'),
(3, 'testuser3', 'test3@test.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.Y5RjlVRNQFMpBJ7P', 'free', 5000, 'active', NOW() - INTERVAL '20 days'),
(4, 'testuser4', 'test4@test.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.Y5RjlVRNQFMpBJ7P', 'vip', 200000, 'active', NOW() - INTERVAL '15 days'),
(5, 'testuser5', 'test5@test.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.Y5RjlVRNQFMpBJ7P', 'free', 10000, 'active', NOW() - INTERVAL '10 days')
ON CONFLICT (id) DO NOTHING;

-- Create tokens for users
INSERT INTO tokens (id, user_id, name, token_key, token_hash, key_prefix, remain_quota, status, created_at) VALUES
(1, 1, 'Main API Key', 'sk-test-abc123def456', 'abc123def456', 'sk-test-', 100000, 'active', NOW() - INTERVAL '30 days'),
(2, 1, 'Secondary Key', 'sk-test-xyz789uvw012', 'xyz789uvw012', 'sk-test-', 50000, 'active', NOW() - INTERVAL '25 days'),
(3, 2, 'Production Key', 'sk-test-prod001', 'prod001', 'sk-test-', 30000, 'active', NOW() - INTERVAL '20 days'),
(4, 3, 'Free Tier Key', 'sk-test-free001', 'free001', 'sk-test-', 5000, 'active', NOW() - INTERVAL '15 days'),
(5, 4, 'Enterprise Key', 'sk-test-ent001', 'ent001', 'sk-test-', 200000, 'active', NOW() - INTERVAL '10 days'),
(6, 5, 'Starter Key', 'sk-test-start001', 'start001', 'sk-test-', 10000, 'active', NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

COMMIT;

-- Generate usage_logs for past 30 days
DO $$
DECLARE
    i INTEGER;
    user_id BIGINT;
    token_id BIGINT;
    model_name TEXT;
    prompt_t INTEGER;
    completion_t INTEGER;
    req_time TIMESTAMP;
    status_code INTEGER;
    cost_val NUMERIC(10,4);
    response_time_val BIGINT;
BEGIN
    -- Generate 500 usage log entries
    FOR i IN 1..500 LOOP
        user_id := 1 + floor(random() * 5)::BIGINT;
        token_id := 1 + floor(random() * 6)::BIGINT;
        
        -- Random model selection
        CASE floor(random() * 8)::INT
            WHEN 0 THEN model_name := 'gpt-4o';
            WHEN 1 THEN model_name := 'gpt-4o-mini';
            WHEN 2 THEN model_name := 'gpt-4-turbo';
            WHEN 3 THEN model_name := 'claude-3-5-sonnet';
            WHEN 4 THEN model_name := 'claude-3-haiku';
            WHEN 5 THEN model_name := 'gemini-1.5-pro';
            WHEN 6 THEN model_name := 'gemini-1.5-flash';
            ELSE model_name := 'deepseek-chat';
        END CASE;
        
        prompt_t := (50 + floor(random() * 500))::BIGINT;
        completion_t := (20 + floor(random() * 300))::BIGINT;
        cost_val := (prompt_t * 0.00001 + completion_t * 0.00003)::NUMERIC(10,4);
        response_time_val := (100 + floor(random() * 2000))::BIGINT;
        status_code := CASE WHEN random() < 0.95 THEN 200 ELSE 500 END;
        req_time := NOW() - (floor(random() * 30) || ' days')::INTERVAL - (floor(random() * 24) || ' hours')::INTERVAL;
        
        INSERT INTO usage_logs (tenant_id, user_id, token_id, model, prompt_tokens, completion_tokens, total_tokens, cost, status_code, response_time_ms, created_at)
        VALUES (1, user_id, token_id, model_name, prompt_t, completion_t, prompt_t + completion_t, cost_val, status_code, response_time_val, req_time);
    END LOOP;
END $$;

-- Generate api_access_logs for past 30 days
DO $$
DECLARE
    i INTEGER;
    user_id BIGINT;
    endpoint_name TEXT;
    method_name TEXT;
    status_c INTEGER;
    response_t INTEGER;
    ip_addr TEXT;
    req_time TIMESTAMP;
BEGIN
    -- Generate 1000 API access log entries
    FOR i IN 1..1000 LOOP
        user_id := 1 + floor(random() * 5)::BIGINT;
        
        -- Random endpoint selection
        CASE floor(random() * 6)::INT
            WHEN 0 THEN endpoint_name := '/v1/chat/completions';
            WHEN 1 THEN endpoint_name := '/v1/completions';
            WHEN 2 THEN endpoint_name := '/v1/embeddings';
            WHEN 3 THEN endpoint_name := '/v1/models';
            WHEN 4 THEN endpoint_name := '/v1/images/generations';
            ELSE endpoint_name := '/v1/audio/transcriptions';
        END CASE;
        
        method_name := CASE WHEN random() < 0.9 THEN 'POST' ELSE 'GET' END;
        status_c := CASE 
            WHEN random() < 0.90 THEN 200 
            WHEN random() < 0.97 THEN 400 
            WHEN random() < 0.99 THEN 401 
            ELSE 500 
        END;
        response_t := (50 + floor(random() * 500))::INTEGER;
        
        -- Random IP addresses
        ip_addr := '192.168.' || (1 + floor(random() * 255))::TEXT || '.' || (1 + floor(random() * 255))::TEXT;
        
        req_time := NOW() - (floor(random() * 30) || ' days')::INTERVAL - (floor(random() * 24) || ' hours')::INTERVAL;
        
        INSERT INTO api_access_logs (user_id, endpoint, method, status_code, response_time, request_ip, created_at)
        VALUES (user_id, endpoint_name, method_name, status_c, response_t, ip_addr, req_time);
    END LOOP;
END $$;

-- Verify data
SELECT 'usage_logs count:' AS label, COUNT(*) AS count FROM usage_logs
UNION ALL
SELECT 'api_access_logs count:', COUNT(*) FROM api_access_logs
UNION ALL
SELECT 'users count:', COUNT(*) FROM users
UNION ALL
SELECT 'tokens count:', COUNT(*) FROM tokens;
