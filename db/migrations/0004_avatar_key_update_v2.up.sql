ALTER TABLE users DROP COLUMN avatar_key_small;
ALTER TABLE users DROP COLUMN avatar_key_medium;
ALTER TABLE users DROP COLUMN avatar_key_large;

ALTER TABLE users ADD COLUMN avatar_key TEXT;