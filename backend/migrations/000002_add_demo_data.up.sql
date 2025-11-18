-- ===========================================
-- Demo Data for Development and Testing
-- ===========================================

-- Insert demo users
-- Login credentials for testing:
--   admin@example.com / password123
--   john@example.com / password123  
--   jane@example.com / password123
INSERT INTO users (id, username, email, role, password_hash, totp_enabled) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'admin', 'admin@example.com', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', false),
    ('550e8400-e29b-41d4-a716-446655440002', 'john_doe', 'john@example.com', 'user', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', false),
    ('550e8400-e29b-41d4-a716-446655440003', 'jane_smith', 'jane@example.com', 'user', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', false);

-- Insert demo files
-- Mix of public/private files with various file types and availability windows
-- API testing: GET /files/{shareToken}
INSERT INTO files (id, share_token, file_name, file_path, file_size, mime_type, owner_id, is_public, available_from, available_to) VALUES
    ('650e8400-e29b-41d4-a716-446655440001', 'abc123token', 'document.pdf', '/uploads/document.pdf', 1048576, 'application/pdf', '550e8400-e29b-41d4-a716-446655440002', true, NOW(), NOW() + INTERVAL '7 days'),
    ('650e8400-e29b-41d4-a716-446655440002', 'xyz456token', 'image.jpg', '/uploads/image.jpg', 2097152, 'image/jpeg', '550e8400-e29b-41d4-a716-446655440002', true, NOW(), NOW() + INTERVAL '14 days'),
    ('650e8400-e29b-41d4-a716-446655440003', 'def789token', 'presentation.pptx', '/uploads/presentation.pptx', 5242880, 'application/vnd.openxmlformats-officedocument.presentationml.presentation', '550e8400-e29b-41d4-a716-446655440003', false, NOW(), NOW() + INTERVAL '30 days'),
    ('650e8400-e29b-41d4-a716-446655440004', 'ghi012token', 'report.xlsx', '/uploads/report.xlsx', 3145728, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', '550e8400-e29b-41d4-a716-446655440001', true, NOW() - INTERVAL '5 days', NOW() + INTERVAL '10 days'),
    ('650e8400-e29b-41d4-a716-446655440005', 'jkl345token', 'video.mp4', '/uploads/video.mp4', 52428800, 'video/mp4', '550e8400-e29b-41d4-a716-446655440002', false, NOW() - INTERVAL '2 days', NOW() + INTERVAL '5 days'),
    ('650e8400-e29b-41d4-a716-446655440006', 'mno678token', 'archive.zip', '/uploads/archive.zip', 10485760, 'application/zip', '550e8400-e29b-41d4-a716-446655440003', true, NOW() - INTERVAL '10 days', NOW() + INTERVAL '20 days');

-- Insert shared_with relationships
-- Demonstrates private file sharing feature
-- API testing: GET /files/shared-with-me
INSERT INTO shared_with (file_id, user_id) VALUES
    ('650e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002'), -- Jane's private presentation shared with John
    ('650e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001'), -- John's private video shared with Admin
    ('650e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440003'); -- John's private video shared with Jane

-- Insert file statistics
-- Aggregated stats showing which files are most popular
-- API testing: GET /files/:id/statistics
INSERT INTO file_statistics (id, file_id, download_count, unique_downloaders, last_downloaded_at) VALUES
    ('750e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440001', 15, 8, NOW() - INTERVAL '2 hours'),
    ('750e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440002', 23, 12, NOW() - INTERVAL '5 hours'),
    ('750e8400-e29b-41d4-a716-446655440003', '650e8400-e29b-41d4-a716-446655440003', 5, 2, NOW() - INTERVAL '1 day'),
    ('750e8400-e29b-41d4-a716-446655440004', '650e8400-e29b-41d4-a716-446655440004', 87, 34, NOW() - INTERVAL '1 hour'),     -- Most popular
    ('750e8400-e29b-41d4-a716-446655440005', '650e8400-e29b-41d4-a716-446655440005', 12, 5, NOW() - INTERVAL '8 hours'),
    ('750e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440006', 156, 67, NOW() - INTERVAL '3 hours');  -- Viral file!

-- Insert download history
-- Detailed log of each download (mix of authenticated and anonymous downloads)
-- API testing: GET /files/:id/download-history
INSERT INTO download_history (id, file_id, downloader_id, downloaded_at, download_completed) VALUES
    -- File 1: document.pdf (John's public PDF - 1MB)
    ('850e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440003', NOW() - INTERVAL '2 days', true),  -- Jane downloaded
    ('850e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440001', NULL, NOW() - INTERVAL '1 day', true),                                      -- Anonymous download
    ('850e8400-e29b-41d4-a716-446655440003', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '5 hours', true), -- Admin downloaded
    ('850e8400-e29b-41d4-a716-446655440004', '650e8400-e29b-41d4-a716-446655440001', NULL, NOW() - INTERVAL '2 hours', true),                                    -- Anonymous download
    
    -- File 2: image.jpg (John's public image - 2MB)
    ('850e8400-e29b-41d4-a716-446655440005', '650e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003', NOW() - INTERVAL '3 days', true),  -- Jane downloaded
    ('850e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440002', NULL, NOW() - INTERVAL '1 day', true),                                      -- Anonymous
    ('850e8400-e29b-41d4-a716-446655440007', '650e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '12 hours', true),-- Admin
    ('850e8400-e29b-41d4-a716-446655440008', '650e8400-e29b-41d4-a716-446655440002', NULL, NOW() - INTERVAL '5 hours', false),                                   -- Failed download
    
    -- File 3: presentation.pptx (Jane's private file - 5MB, shared with John)
    ('850e8400-e29b-41d4-a716-446655440009', '650e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '1 day', true),   -- John (has access)
    ('850e8400-e29b-41d4-a716-446655440010', '650e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440003', NOW() - INTERVAL '6 hours', true), -- Jane (owner)
    
    -- File 4: report.xlsx (Admin's trending file - 3MB, most popular!)
    ('850e8400-e29b-41d4-a716-446655440011', '650e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '5 days', true),
    ('850e8400-e29b-41d4-a716-446655440012', '650e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440003', NOW() - INTERVAL '4 days', true),
    ('850e8400-e29b-41d4-a716-446655440013', '650e8400-e29b-41d4-a716-446655440004', NULL, NOW() - INTERVAL '3 days', true),
    ('850e8400-e29b-41d4-a716-446655440014', '650e8400-e29b-41d4-a716-446655440004', NULL, NOW() - INTERVAL '2 days', true),
    ('850e8400-e29b-41d4-a716-446655440015', '650e8400-e29b-41d4-a716-446655440004', NULL, NOW() - INTERVAL '1 day', true),
    ('850e8400-e29b-41d4-a716-446655440016', '650e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '12 hours', true),
    ('850e8400-e29b-41d4-a716-446655440017', '650e8400-e29b-41d4-a716-446655440004', NULL, NOW() - INTERVAL '6 hours', true),
    ('850e8400-e29b-41d4-a716-446655440018', '650e8400-e29b-41d4-a716-446655440004', NULL, NOW() - INTERVAL '3 hours', true),
    ('850e8400-e29b-41d4-a716-446655440019', '650e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440003', NOW() - INTERVAL '1 hour', true),
    
    -- File 5: video.mp4 (John's private video - 50MB, shared with Admin and Jane)
    ('850e8400-e29b-41d4-a716-446655440020', '650e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '2 days', true),   -- Admin (has access)
    ('850e8400-e29b-41d4-a716-446655440021', '650e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440003', NOW() - INTERVAL '1 day', true),    -- Jane (has access)
    ('850e8400-e29b-41d4-a716-446655440022', '650e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440003', NOW() - INTERVAL '18 hours', false), -- Jane's failed download
    ('850e8400-e29b-41d4-a716-446655440023', '650e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '8 hours', true),   -- Admin again
    
    -- File 6: archive.zip (Jane's viral file - 10MB, most downloaded!)
    ('850e8400-e29b-41d4-a716-446655440024', '650e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '10 days', true),
    ('850e8400-e29b-41d4-a716-446655440025', '650e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '9 days', true),
    ('850e8400-e29b-41d4-a716-446655440026', '650e8400-e29b-41d4-a716-446655440006', NULL, NOW() - INTERVAL '8 days', true),
    ('850e8400-e29b-41d4-a716-446655440027', '650e8400-e29b-41d4-a716-446655440006', NULL, NOW() - INTERVAL '7 days', true),
    ('850e8400-e29b-41d4-a716-446655440028', '650e8400-e29b-41d4-a716-446655440006', NULL, NOW() - INTERVAL '6 days', true),
    ('850e8400-e29b-41d4-a716-446655440029', '650e8400-e29b-41d4-a716-446655440006', NULL, NOW() - INTERVAL '5 days', true),
    ('850e8400-e29b-41d4-a716-446655440030', '650e8400-e29b-41d4-a716-446655440006', NULL, NOW() - INTERVAL '4 days', false),
    ('850e8400-e29b-41d4-a716-446655440031', '650e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '3 days', true),
    ('850e8400-e29b-41d4-a716-446655440032', '650e8400-e29b-41d4-a716-446655440006', NULL, NOW() - INTERVAL '2 days', true),
    ('850e8400-e29b-41d4-a716-446655440033', '650e8400-e29b-41d4-a716-446655440006', NULL, NOW() - INTERVAL '1 day', true),
    ('850e8400-e29b-41d4-a716-446655440034', '650e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '18 hours', true),
    ('850e8400-e29b-41d4-a716-446655440035', '650e8400-e29b-41d4-a716-446655440006', NULL, NOW() - INTERVAL '12 hours', true),
    ('850e8400-e29b-41d4-a716-446655440036', '650e8400-e29b-41d4-a716-446655440006', NULL, NOW() - INTERVAL '6 hours', true),
    ('850e8400-e29b-41d4-a716-446655440037', '650e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440003', NOW() - INTERVAL '3 hours', true);