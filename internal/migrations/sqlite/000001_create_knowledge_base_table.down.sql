-- Drop trigger
DROP TRIGGER IF EXISTS update_knowledge_base_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_knowledge_base_created_at;
DROP INDEX IF EXISTS idx_knowledge_base_name;

-- Drop table
DROP TABLE IF EXISTS knowledge_base;