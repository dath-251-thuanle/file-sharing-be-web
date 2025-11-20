-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

-- Create user_role enum
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('user', 'admin');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

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
-- Core table for uploaded files with sharing and security features
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Sharing
    share_token VARCHAR(32) UNIQUE NOT NULL,  -- Unique token for public share link (e.g., /files/{shareToken})
    
    -- File information
    file_name VARCHAR(255) NOT NULL,           -- Original filename
    file_path VARCHAR(512) NOT NULL,           -- Server storage path
    file_size BIGINT NOT NULL,                 -- Size in bytes
    mime_type VARCHAR(100),                    -- Content type (e.g., application/pdf)
    
    -- Ownership
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,  -- NULL for anonymous uploads
    
    -- Access control
    is_public BOOLEAN DEFAULT TRUE,            -- Public files: anyone with link can access
                                               -- Private files: only owner and shared_with users
    password_hash VARCHAR(255),                -- Optional password protection (bcrypt hash)
    
    -- Availability window
    available_from TIMESTAMP WITH TIME ZONE,   -- File becomes available at this time
    available_to TIMESTAMP WITH TIME ZONE,     -- File expires at this time
    
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
-- Many-to-many relationship: which users have access to which private files
-- API endpoint: POST /files/:id/share (owner shares file with specific users)
CREATE TABLE shared_with (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    UNIQUE(file_id, user_id)  -- Prevent duplicate sharing
);

-- Create indexes for shared_with
CREATE INDEX idx_shared_with_file ON shared_with(file_id);  -- Find all users for a file
CREATE INDEX idx_shared_with_user ON shared_with(user_id);  -- Find all files shared with user

-- Create system_policy table
-- System-wide configuration (admin can update via API)
-- Single row table (id=1) for global settings
CREATE TABLE system_policy (
    id SERIAL PRIMARY KEY,
    max_file_size_mb INTEGER DEFAULT 50,         -- Maximum upload size (MB)
    min_validity_hours INTEGER DEFAULT 1,        -- Minimum file availability duration
    max_validity_days INTEGER DEFAULT 30,        -- Maximum file availability duration
    default_validity_days INTEGER DEFAULT 7,     -- Default validity if not specified
    require_password_min_length INTEGER DEFAULT 6  -- Minimum password length requirement
);

-- Insert default policy
-- API endpoints: GET /admin/policy, PATCH /admin/policy
INSERT INTO system_policy (id, min_validity_hours, max_validity_days, default_validity_days) 
VALUES (1, 1, 30, 7);

-- Create file_statistics table
-- Aggregated statistics for each file (only for files with owners, not anonymous uploads)
-- Updated automatically by application when downloads/views occur
-- API endpoint: GET /files/:id/statistics (owner only)
CREATE TABLE file_statistics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL UNIQUE REFERENCES files(id) ON DELETE CASCADE,
    download_count INTEGER DEFAULT 0,              -- Total number of downloads
    unique_downloaders INTEGER DEFAULT 0,          -- Number of unique users who downloaded
    last_downloaded_at TIMESTAMP WITH TIME ZONE,   -- Most recent download timestamp           -- Total page views (file info page)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_file_statistics_file_id ON file_statistics(file_id);
CREATE INDEX idx_file_statistics_download_count ON file_statistics(download_count DESC);  -- Sort by popularity
CREATE INDEX idx_file_statistics_last_downloaded ON file_statistics(last_downloaded_at DESC);  -- Sort by recent activity

-- Create download_history table
-- Simple download history
-- Records each download attempt for audit and analytics
-- API endpoint: GET /files/:id/download-history
CREATE TABLE download_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    downloader_id UUID REFERENCES users(id) ON DELETE SET NULL,  -- NULL for anonymous downloads
    downloaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    download_completed BOOLEAN DEFAULT TRUE  -- FALSE if download was interrupted
);

CREATE INDEX idx_download_history_file_id ON download_history(file_id);  -- Get download history for a file
CREATE INDEX idx_download_history_downloader_id ON download_history(downloader_id);  -- Get user's download history
CREATE INDEX idx_download_history_downloaded_at ON download_history(downloaded_at DESC);  -- Sort by recency