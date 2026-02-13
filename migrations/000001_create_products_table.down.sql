-- Drop trigger
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_products_deleted_at;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_status;
DROP INDEX IF EXISTS idx_products_api_key;

-- Drop table
DROP TABLE IF EXISTS products;
