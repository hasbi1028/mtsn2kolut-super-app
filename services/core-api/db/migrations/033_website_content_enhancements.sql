-- Migration 033: website content enhancements
-- Adds featured flag and SEO metadata fields to website_contents.

ALTER TABLE website_contents
    ADD COLUMN is_featured       BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN meta_title        TEXT    NOT NULL DEFAULT '',
    ADD COLUMN meta_description  TEXT    NOT NULL DEFAULT '';

-- Index to quickly fetch featured published content for homepage curation.
CREATE INDEX idx_website_contents_featured
    ON website_contents (kind, is_featured, status, published_at DESC)
    WHERE is_featured = TRUE AND status = 'published';
