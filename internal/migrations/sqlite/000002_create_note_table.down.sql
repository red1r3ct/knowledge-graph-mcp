-- Drop triggers
DROP TRIGGER IF EXISTS update_notes_updated_at;
DROP TRIGGER IF EXISTS notes_fts_delete;
DROP TRIGGER IF EXISTS notes_fts_update;
DROP TRIGGER IF EXISTS notes_fts_insert;

-- Drop full-text search virtual table
DROP TABLE IF EXISTS notes_fts;

-- Drop indexes
DROP INDEX IF EXISTS idx_notes_updated_at;
DROP INDEX IF EXISTS idx_notes_created_at;
DROP INDEX IF EXISTS idx_notes_type;
DROP INDEX IF EXISTS idx_notes_title_unique;

-- Drop notes table
DROP TABLE IF EXISTS notes;