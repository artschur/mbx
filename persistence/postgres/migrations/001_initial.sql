-- migrations/001_create_tables.up.sql
CREATE TABLE IF NOT EXISTS templates (
    content_id TEXT PRIMARY KEY,
    friendly_name TEXT NOT nil,
    language TEXT,
    body TEXT,
    variables JSONB,
    types JSONB,
    date_created TIMESTAMP WITH TIME ZONE,
    date_updated TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS webhook_requests (
    id SERIAL PRIMARY KEY,
    message_sid TEXT NOT NULL,
    status TEXT,
    raw_data JSONB,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scheduled_messages (
    id SERIAL PRIMARY KEY,
    to_phone TEXT NOT NULL,
    body TEXT,
    template_id TEXT,
    content JSONB,
    language TEXT,
    scheduled_for TIMESTAMP WITH TIME ZONE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE,
    status TEXT DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

