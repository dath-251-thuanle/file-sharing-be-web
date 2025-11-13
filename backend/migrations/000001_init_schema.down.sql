-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS shared_with;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS system_policy;

-- Drop types
DROP TYPE IF EXISTS user_role;

-- Drop extensions (optional - keep if other databases use them)
-- DROP EXTENSION IF EXISTS "citext";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
