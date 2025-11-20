-- Demo data for testing and development
-- Creates sample users, files, and relationships

-- Create demo users
-- Password for all users: "password123" (bcrypt hash)
INSERT INTO users (username, email, password_hash, role, totp_enabled, totp_secret) VALUES
    ('admin', 'admin@example.com', '$2a$10$rZ5p.O8YgKjH8YHVXq5xFO6P5vL0qjKZxvYXqYqYqYqYqYqYqYqYq', 'admin', FALSE, NULL),
    ('nguyenvana', 'nguyenvana@example.com', '$2a$10$rZ5p.O8YgKjH8YHVXq5xFO6P5vL0qjKZxvYXqYqYqYqYqYqYqYqYq', 'user', FALSE, NULL),
    ('tranthib', 'tranthib@example.com', '$2a$10$rZ5p.O8YgKjH8YHVXq5xFO6P5vL0qjKZxvYXqYqYqYqYqYqYqYqYq', 'user', FALSE, NULL);

-- Create demo files
-- File 1: Public file without password (by nguyenvana)
INSERT INTO files (
    share_token, file_name, file_path, file_size, mime_type, owner_id,
    is_public, password_hash,
    available_from, available_to
) VALUES (
    'demo_public_file_001',
    'public_document.pdf',
    '/uploads/2025/11/public_document.pdf',
    2048576,
    'application/pdf',
    (SELECT id FROM users WHERE username = 'nguyenvana'),
    TRUE,
    NULL,
    NOW(),
    NOW() + INTERVAL '30 days'
);

-- File 2: Password-protected file (by tranthib)
-- Password: "secret123"
INSERT INTO files (
    share_token, file_name, file_path, file_size, mime_type, owner_id,
    is_public, password_hash,
    available_from, available_to
) VALUES (
    'demo_password_file_002',
    'confidential_report.docx',
    '/uploads/2025/11/confidential_report.docx',
    512000,
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    (SELECT id FROM users WHERE username = 'tranthib'),
    FALSE,
    '$2a$10$rZ5p.O8YgKjH8YHVXq5xFO6P5vL0qjKZxvYXqYqYqYqYqYqYqYqYq',
    NOW(),
    NOW() + INTERVAL '7 days'
);

-- File 3: Future file (not yet available)
INSERT INTO files (
    share_token, file_name, file_path, file_size, mime_type, owner_id,
    is_public, password_hash,
    available_from, available_to
) VALUES (
    'demo_future_file_003',
    'meeting_agenda.pdf',
    '/uploads/2025/11/meeting_agenda.pdf',
    102400,
    'application/pdf',
    (SELECT id FROM users WHERE username = 'admin'),
    TRUE,
    NULL,
    NOW() + INTERVAL '2 days',
    NOW() + INTERVAL '9 days'
);

-- File 4: Expired file
INSERT INTO files (
    share_token, file_name, file_path, file_size, mime_type, owner_id,
    is_public, password_hash,
    available_from, available_to
) VALUES (
    'demo_expired_file_004',
    'old_presentation.pptx',
    '/uploads/2025/11/old_presentation.pptx',
    3145728,
    'application/vnd.openxmlformats-officedocument.presentationml.presentation',
    (SELECT id FROM users WHERE username = 'nguyenvana'),
    TRUE,
    NULL,
    NOW() - INTERVAL '10 days',
    NOW() - INTERVAL '3 days'
);

-- File 5: Anonymous upload (no owner)
INSERT INTO files (
    share_token, file_name, file_path, file_size, mime_type, owner_id,
    is_public, password_hash,
    available_from, available_to
) VALUES (
    'demo_anon_file_005',
    'anonymous_upload.jpg',
    '/uploads/2025/11/anonymous_upload.jpg',
    1048576,
    'image/jpeg',
    NULL,
    TRUE,
    NULL,
    NOW(),
    NOW() + INTERVAL '7 days'
);

-- Create file sharing relationships
-- Share file 2 (confidential_report.docx) with nguyenvana
INSERT INTO shared_with (file_id, user_id) VALUES (
    (SELECT id FROM files WHERE share_token = 'demo_password_file_002'),
    (SELECT id FROM users WHERE username = 'nguyenvana')
);

-- Create file statistics for files with owners
INSERT INTO file_statistics (file_id, download_count, unique_downloaders, last_downloaded_at)
SELECT 
    f.id,
    FLOOR(RANDOM() * 50)::INTEGER,
    FLOOR(RANDOM() * 10)::INTEGER,
    NOW() - (INTERVAL '1 hour' * FLOOR(RANDOM() * 48))
FROM files f
WHERE f.owner_id IS NOT NULL;

-- Create download history entries
-- nguyenvana downloads public_document.pdf
INSERT INTO download_history (file_id, downloader_id, download_completed) VALUES (
    (SELECT id FROM files WHERE share_token = 'demo_public_file_001'),
    (SELECT id FROM users WHERE username = 'nguyenvana'),
    TRUE
);

-- tranthib downloads public_document.pdf
INSERT INTO download_history (file_id, downloader_id, download_completed) VALUES (
    (SELECT id FROM files WHERE share_token = 'demo_public_file_001'),
    (SELECT id FROM users WHERE username = 'tranthib'),
    TRUE
);

-- Anonymous download of public_document.pdf
INSERT INTO download_history (file_id, downloader_id, download_completed) VALUES (
    (SELECT id FROM files WHERE share_token = 'demo_public_file_001'),
    NULL,
    TRUE
);

-- Failed download (interrupted)
INSERT INTO download_history (file_id, downloader_id, download_completed) VALUES (
    (SELECT id FROM files WHERE share_token = 'demo_password_file_002'),
    (SELECT id FROM users WHERE username = 'tranthib'),
    FALSE
);

