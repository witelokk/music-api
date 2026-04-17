ALTER TABLE songs RENAME COLUMN cover_url TO cover_media_id;
ALTER TABLE songs RENAME COLUMN stream_url TO stream_media_id;

ALTER TABLE artists RENAME COLUMN avatar_url TO avatar_media_id;
ALTER TABLE artists RENAME COLUMN cover_url TO cover_media_id;

ALTER TABLE releases RENAME COLUMN cover_url TO cover_media_id;
