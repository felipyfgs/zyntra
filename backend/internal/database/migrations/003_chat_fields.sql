-- Add favorite and archived fields to chats
ALTER TABLE chats ADD COLUMN IF NOT EXISTS is_favorite BOOLEAN DEFAULT FALSE;
ALTER TABLE chats ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT FALSE;

-- Add index for filtering
CREATE INDEX IF NOT EXISTS idx_chats_is_favorite ON chats(is_favorite) WHERE is_favorite = true;
CREATE INDEX IF NOT EXISTS idx_chats_is_archived ON chats(is_archived) WHERE is_archived = true;
