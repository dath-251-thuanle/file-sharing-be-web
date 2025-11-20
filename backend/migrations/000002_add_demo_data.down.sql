-- Remove all demo data
-- Execute in reverse order of creation to avoid foreign key constraints

-- Delete download history
DELETE FROM download_history 
WHERE file_id IN (
    SELECT id FROM files WHERE share_token LIKE 'demo_%'
);

-- Delete file statistics
DELETE FROM file_statistics 
WHERE file_id IN (
    SELECT id FROM files WHERE share_token LIKE 'demo_%'
);

-- Delete shared_with relationships
DELETE FROM shared_with 
WHERE file_id IN (
    SELECT id FROM files WHERE share_token LIKE 'demo_%'
);

-- Delete demo files
DELETE FROM files WHERE share_token LIKE 'demo_%';

-- Delete demo users
DELETE FROM users WHERE username IN ('admin', 'nguyenvana', 'tranthib');

