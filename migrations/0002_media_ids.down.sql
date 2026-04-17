ALTER TABLE releases RENAME COLUMN cover_media_id TO cover_url;

ALTER TABLE artists RENAME COLUMN cover_media_id TO cover_url;
ALTER TABLE artists RENAME COLUMN avatar_media_id TO avatar_url;

ALTER TABLE songs RENAME COLUMN stream_media_id TO stream_url;
ALTER TABLE songs RENAME COLUMN cover_media_id TO cover_url;
