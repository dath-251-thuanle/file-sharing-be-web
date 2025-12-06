-- Rollback: Recreate shared_with table (for migration rollback purposes only)
-- Note: This is for migration rollback only. The shared_with table is deprecated.

-- Recreate the shared_with table
CREATE TABLE IF NOT EXISTS shared_with (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL,
    user_id UUID NOT NULL,
    email VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(file_id, user_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_shared_with_file_id ON shared_with(file_id);
CREATE INDEX IF NOT EXISTS idx_shared_with_user_id ON shared_with(user_id);
CREATE INDEX IF NOT EXISTS idx_shared_with_email ON shared_with(email);
CREATE INDEX IF NOT EXISTS idx_file_user ON shared_with(file_id, user_id);

-- Add foreign key constraints
ALTER TABLE shared_with
ADD CONSTRAINT shared_with_file_id_fkey 
FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE;

ALTER TABLE shared_with
ADD CONSTRAINT shared_with_user_id_fkey 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Recreate unique constraint on users(id, email) if needed
ALTER TABLE users
ADD CONSTRAINT unique_id_email UNIQUE (id, email);

