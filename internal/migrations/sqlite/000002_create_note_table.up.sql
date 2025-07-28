-- Create notes table
CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('text', 'markdown', 'code', 'link', 'image')),
    tags TEXT, -- JSON array of tags
    metadata TEXT, -- JSON object for additional properties
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create unique index on title for uniqueness constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_notes_title_unique ON notes(title);

-- Create index on type for filtering
CREATE INDEX IF NOT EXISTS idx_notes_type ON notes(type);

-- Create index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at DESC);

-- Create index on updated_at for sorting
CREATE INDEX IF NOT EXISTS idx_notes_updated_at ON notes(updated_at DESC);

-- Create full-text search virtual table for notes
CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts5(
    title,
    content,
    content='notes',
    content_rowid='id'
);

-- Create triggers to keep FTS table in sync with notes table
CREATE TRIGGER IF NOT EXISTS notes_fts_insert AFTER INSERT ON notes BEGIN
    INSERT INTO notes_fts(rowid, title, content) VALUES (NEW.id, NEW.title, NEW.content);
END;

CREATE TRIGGER IF NOT EXISTS notes_fts_update AFTER UPDATE ON notes BEGIN
    UPDATE notes_fts SET 
        title = NEW.title,
        content = NEW.content
    WHERE rowid = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS notes_fts_delete AFTER DELETE ON notes BEGIN
    DELETE FROM notes_fts WHERE rowid = OLD.id;
END;

-- Create trigger to update updated_at on row update
CREATE TRIGGER IF NOT EXISTS update_notes_updated_at 
AFTER UPDATE ON notes
FOR EACH ROW
BEGIN
    UPDATE notes SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;