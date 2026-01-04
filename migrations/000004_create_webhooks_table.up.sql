-- Create webhooks table
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('slack', 'discord', 'custom')),
    url TEXT NOT NULL,
    events TEXT[] NOT NULL, -- Array de eventos habilitados
    is_active BOOLEAN NOT NULL DEFAULT true,
    secret VARCHAR(255), -- Para validação de webhooks custom
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_trigger TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_webhooks_user_id ON webhooks(user_id);
CREATE INDEX idx_webhooks_is_active ON webhooks(is_active);
CREATE INDEX idx_webhooks_type ON webhooks(type);

-- Add comments
COMMENT ON TABLE webhooks IS 'Stores webhook configurations for users';
COMMENT ON COLUMN webhooks.id IS 'Unique identifier for the webhook (UUID)';
COMMENT ON COLUMN webhooks.user_id IS 'References the user who owns this webhook';
COMMENT ON COLUMN webhooks.name IS 'Friendly name for the webhook';
COMMENT ON COLUMN webhooks.type IS 'Type of webhook: slack, discord, or custom';
COMMENT ON COLUMN webhooks.url IS 'Webhook URL endpoint';
COMMENT ON COLUMN webhooks.events IS 'Array of event types that trigger this webhook';
COMMENT ON COLUMN webhooks.is_active IS 'Whether the webhook is currently active';
COMMENT ON COLUMN webhooks.secret IS 'Optional secret for custom webhook validation';
COMMENT ON COLUMN webhooks.last_trigger IS 'Timestamp of the last webhook trigger';
