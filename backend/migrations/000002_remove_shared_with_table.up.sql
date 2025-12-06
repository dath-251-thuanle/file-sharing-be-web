-- Remove shared_with table (no longer needed - using JSONB column shared_with_emails in files table)
-- This migration safely drops the shared_with table if it exists

-- Only proceed if the table exists
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'shared_with') THEN
        -- Drop foreign key constraints first (if they exist)
        ALTER TABLE shared_with DROP CONSTRAINT IF EXISTS fk_shared_with_user_email;
        ALTER TABLE shared_with DROP CONSTRAINT IF EXISTS shared_with_file_id_fkey;
        ALTER TABLE shared_with DROP CONSTRAINT IF EXISTS shared_with_user_id_fkey;

        -- Drop indexes
        DROP INDEX IF EXISTS idx_shared_with_email;
        DROP INDEX IF EXISTS idx_file_user;
        DROP INDEX IF EXISTS idx_shared_with_file_id;
        DROP INDEX IF EXISTS idx_shared_with_user_id;

        -- Drop the table
        DROP TABLE shared_with;
        
        RAISE NOTICE 'Dropped shared_with table';
    ELSE
        RAISE NOTICE 'shared_with table does not exist, skipping';
    END IF;
END $$;

-- Also remove the unique constraint on users(id, email) that was added for shared_with
ALTER TABLE users DROP CONSTRAINT IF EXISTS unique_id_email;

