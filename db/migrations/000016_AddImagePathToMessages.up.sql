-- Add image support to messages
ALTER TABLE messages ADD COLUMN image_path TEXT;

-- Index for filtering messages with images
CREATE INDEX idx_messages_image_path ON messages(image_path);
