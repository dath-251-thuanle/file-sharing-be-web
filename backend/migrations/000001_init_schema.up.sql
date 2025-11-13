-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

-- Create user_role enum
CREATE TYPE user_role AS ENUM ('user', 'admin');

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username CITEXT UNIQUE NOT NULL,
    email CITEXT UNIQUE NOT NULL,
    role user_role DEFAULT 'user',
    password_hash VARCHAR(255) NOT NULL,
    totp_secret VARCHAR(32),
    totp_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- Create files table
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Share token để tạo link public
    share_token VARCHAR(32) UNIQUE NOT NULL,
    
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(512) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    
    -- Owner có thể NULL (upload ẩn danh)
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Quyền truy cập
    is_public BOOLEAN DEFAULT TRUE,
    password_hash VARCHAR(255),
    
    -- Thời gian hiệu lực
    available_from TIMESTAMP WITH TIME ZONE,
    available_to TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_available_dates CHECK (
        available_from IS NULL OR 
        available_to IS NULL OR 
        available_from < available_to
    ),
    CONSTRAINT chk_at_least_one_date CHECK (
        available_from IS NOT NULL OR available_to IS NOT NULL
    )
);

-- Create indexes for files
CREATE INDEX idx_files_owner ON files(owner_id);
CREATE INDEX idx_files_share_token ON files(share_token);
CREATE INDEX idx_files_created_at ON files(created_at);
CREATE INDEX idx_files_availability ON files(available_from, available_to);

-- Create shared_with table
CREATE TABLE shared_with (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    UNIQUE(file_id, user_id)
);

-- Create indexes for shared_with
CREATE INDEX idx_shared_with_file ON shared_with(file_id);
CREATE INDEX idx_shared_with_user ON shared_with(user_id);

-- Create system_policy table
CREATE TABLE system_policy (
    id SERIAL PRIMARY KEY,
    max_file_size_mb INTEGER DEFAULT 50,
    min_validity_hours INTEGER DEFAULT 1,
    max_validity_days INTEGER DEFAULT 30,
    default_validity_days INTEGER DEFAULT 7,
    require_password_min_length INTEGER DEFAULT 6
);

-- Insert default policy
INSERT INTO system_policy (id, min_validity_hours, max_validity_days, default_validity_days) 
VALUES (1, 1, 30, 7);
