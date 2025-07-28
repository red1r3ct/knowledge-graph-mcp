-- Create connections table
CREATE TABLE IF NOT EXISTS connections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_note_id INTEGER NOT NULL,
    to_note_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN (
        'relates_to', 'references', 'supports', 'contradicts', 'influences',
        'depends_on', 'similar_to', 'part_of', 'cites', 'follows', 'precedes'
    )),
    description TEXT,
    strength INTEGER NOT NULL CHECK (strength >= 1 AND strength <= 10) DEFAULT 5,
    metadata TEXT, -- JSON object for additional properties
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (from_note_id) REFERENCES notes(id) ON DELETE CASCADE,
    FOREIGN KEY (to_note_id) REFERENCES notes(id) ON DELETE CASCADE
);

-- Create unique index to prevent duplicate connections
CREATE UNIQUE INDEX IF NOT EXISTS idx_connections_unique 
ON connections(from_note_id, to_note_id, type);

-- Create index on from_note_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_connections_from_note_id ON connections(from_note_id);

-- Create index on to_note_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_connections_to_note_id ON connections(to_note_id);

-- Create index on type for filtering
CREATE INDEX IF NOT EXISTS idx_connections_type ON connections(type);

-- Create index on strength for filtering
CREATE INDEX IF NOT EXISTS idx_connections_strength ON connections(strength);

-- Create index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_connections_created_at ON connections(created_at DESC);

-- Create index on updated_at for sorting
CREATE INDEX IF NOT EXISTS idx_connections_updated_at ON connections(updated_at DESC);

-- Create composite index for common queries
CREATE INDEX IF NOT EXISTS idx_connections_from_to ON connections(from_note_id, to_note_id);

-- Create trigger to update updated_at on row update
CREATE TRIGGER IF NOT EXISTS update_connections_updated_at 
AFTER UPDATE ON connections
FOR EACH ROW
BEGIN
    UPDATE connections SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create trigger to prevent self-connections
CREATE TRIGGER IF NOT EXISTS prevent_self_connection
BEFORE INSERT ON connections
FOR EACH ROW
WHEN NEW.from_note_id = NEW.to_note_id
BEGIN
    SELECT RAISE(ABORT, 'Self-connections are not allowed');
END;