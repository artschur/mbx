CREATE TYPE message_status AS ENUM('pending', 'sent', 'failed');

CREATE TABLE scheduled_messages (
  id UUID PRIMARY KEY,
  to_type VARCHAR(255) NOT NULL,
  send_at TIMESTAMP NOT NULL,
  content TEXT NOT NULL,
  provider_template_id VARCHAR(255) NOT NULL,
  message_type VARCHAR(255) NOT NULL,
  status message_status NOT NULL DEFAULT 'pending',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
