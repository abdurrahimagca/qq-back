ALTER TABLE users DROP COLUMN avatar_url;
ALTER TABLE users ADD COLUMN avatar_key_small VARCHAR(512);
ALTER TABLE users ADD COLUMN avatar_key_medium VARCHAR(512);
ALTER TABLE users ADD COLUMN avatar_key_large VARCHAR(512);