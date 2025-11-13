-- Insert demo users
-- Password for all demo users: "password123" (bcrypt hash)
INSERT INTO users (id, username, email, role, password_hash, totp_enabled) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'admin', 'admin@example.com', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', false),
    ('550e8400-e29b-41d4-a716-446655440002', 'john_doe', 'john@example.com', 'user', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', false),
    ('550e8400-e29b-41d4-a716-446655440003', 'jane_smith', 'jane@example.com', 'user', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', false);

-- Insert demo files
INSERT INTO files (id, share_token, file_name, file_path, file_size, mime_type, owner_id, is_public, available_from, available_to) VALUES
    ('650e8400-e29b-41d4-a716-446655440001', 'abc123token', 'document.pdf', '/uploads/document.pdf', 1048576, 'application/pdf', '550e8400-e29b-41d4-a716-446655440002', true, NOW(), NOW() + INTERVAL '7 days'),
    ('650e8400-e29b-41d4-a716-446655440002', 'xyz456token', 'image.jpg', '/uploads/image.jpg', 2097152, 'image/jpeg', '550e8400-e29b-41d4-a716-446655440002', true, NOW(), NOW() + INTERVAL '14 days'),
    ('650e8400-e29b-41d4-a716-446655440003', 'def789token', 'presentation.pptx', '/uploads/presentation.pptx', 5242880, 'application/vnd.openxmlformats-officedocument.presentationml.presentation', '550e8400-e29b-41d4-a716-446655440003', false, NOW(), NOW() + INTERVAL '30 days');

-- Insert shared_with relationships
INSERT INTO shared_with (file_id, user_id) VALUES
    ('650e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002'); -- jane's private file shared with john
