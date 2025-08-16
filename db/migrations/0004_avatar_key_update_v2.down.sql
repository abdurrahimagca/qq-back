ALTER TABLE users ADD COLUMN avatar_key_small TEXT;
ALTER TABLE users ADD COLUMN avatar_key_medium TEXT;
ALTER TABLE users ADD COLUMN avatar_key_large TEXT;

ALTER TABLE users DROP COLUMN avatar_key;