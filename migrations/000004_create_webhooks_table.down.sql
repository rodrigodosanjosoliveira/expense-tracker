-- Drop indexes
DROP INDEX IF EXISTS idx_webhooks_type;
DROP INDEX IF EXISTS idx_webhooks_is_active;
DROP INDEX IF EXISTS idx_webhooks_user_id;

-- Drop webhooks table
DROP TABLE IF EXISTS webhooks;
