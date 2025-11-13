-- ============================================
-- DEMO QUERIES FOR FILE SHARING SYSTEM
-- ============================================

-- ============================================
-- 1. USERS - Quản lý người dùng
-- ============================================

-- Tạo user thường
INSERT INTO users (username, email, password_hash) 
VALUES 
    ('nguyenvana', 'nguyenvana@example.com', '$2b$12$hash1'),
    ('tranthib', 'tranthib@example.com', '$2b$12$hash2');

-- Tạo admin
INSERT INTO users (username, email, role, password_hash) 
VALUES ('admin', 'admin@example.com', 'admin', '$2b$12$hash3');

-- Tạo user với 2FA enabled
INSERT INTO users (username, email, password_hash, totp_secret, totp_enabled) 
VALUES ('secureuser', 'secure@example.com', '$2b$12$hash4', 'JBSWY3DPEHPK3PXP', TRUE);

-- Xem tất cả users
SELECT id, username, email, role, totp_enabled, created_at 
FROM users;

-- Tìm user theo email (case-insensitive)
SELECT * FROM users WHERE email = 'NGUYENVANA@EXAMPLE.COM';

-- Đếm số user theo role
SELECT role, COUNT(*) as total 
FROM users 
GROUP BY role;


-- ============================================
-- 2. FILES - Quản lý files
-- ============================================

-- Upload file công khai, không password, có owner
INSERT INTO files (share_token, file_name, file_path, file_size, mime_type, owner_id, is_public, available_from, available_to)
VALUES (
    'abc123def456',
    'presentation.pdf',
    '/uploads/2025/11/presentation.pdf',
    2048576,
    'application/pdf',
    (SELECT id FROM users WHERE username = 'nguyenvana'),
    TRUE,
    NOW(),
    NOW() + INTERVAL '7 days'
);

-- Upload file ẩn danh với password
INSERT INTO files (share_token, file_name, file_path, file_size, mime_type, is_public, password_hash, available_to)
VALUES (
    'xyz789ghi012',
    'secret_document.docx',
    '/uploads/2025/11/secret_document.docx',
    512000,
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    FALSE,
    '$2b$12$passwordhash',
    NOW() + INTERVAL '3 days'
);

-- Upload file với thời gian bắt đầu trong tương lai
INSERT INTO files (share_token, file_name, file_path, file_size, mime_type, owner_id, available_from, available_to)
VALUES (
    'future999',
    'meeting_agenda.pdf',
    '/uploads/2025/11/meeting_agenda.pdf',
    102400,
    'application/pdf',
    (SELECT id FROM users WHERE username = 'admin'),
    NOW() + INTERVAL '2 days',
    NOW() + INTERVAL '9 days'
);

-- Xem tất cả files
SELECT 
    f.id,
    f.share_token,
    f.file_name,
    f.file_size / 1024.0 / 1024.0 as size_mb,
    u.username as owner,
    f.is_public,
    f.password_hash IS NOT NULL as has_password,
    f.available_from,
    f.available_to,
    f.created_at
FROM files f
LEFT JOIN users u ON f.owner_id = u.id;

-- Tìm file theo share token
SELECT * FROM files WHERE share_token = 'abc123def456';

-- Lấy các file đang available (trong khoảng thời gian hợp lệ)
SELECT 
    f.file_name,
    f.share_token,
    f.available_from,
    f.available_to
FROM files f
WHERE 
    (f.available_from IS NULL OR f.available_from <= NOW())
    AND (f.available_to IS NULL OR f.available_to >= NOW());

-- Lấy các file sắp hết hạn (trong 24h tới)
SELECT 
    f.file_name,
    f.share_token,
    f.available_to,
    f.available_to - NOW() as time_remaining
FROM files f
WHERE f.available_to BETWEEN NOW() AND NOW() + INTERVAL '24 hours'
ORDER BY f.available_to ASC;

-- Lấy file của một user cụ thể
SELECT 
    f.file_name,
    f.share_token,
    f.file_size,
    f.created_at
FROM files f
WHERE f.owner_id = (SELECT id FROM users WHERE username = 'nguyenvana');

-- Tìm file công khai không cần password
SELECT 
    f.file_name,
    f.share_token,
    f.file_size / 1024.0 as size_kb
FROM files f
WHERE f.is_public = TRUE 
    AND f.password_hash IS NULL
    AND (f.available_from IS NULL OR f.available_from <= NOW())
    AND (f.available_to IS NULL OR f.available_to >= NOW());


-- ============================================
-- 3. SHARED_WITH - Chia sẻ file với user cụ thể
-- ============================================

-- Chia sẻ file với một user
INSERT INTO shared_with (file_id, user_id)
VALUES (
    (SELECT id FROM files WHERE share_token = 'abc123def456'),
    (SELECT id FROM users WHERE username = 'tranthib')
);

-- Chia sẻ file với nhiều users
INSERT INTO shared_with (file_id, user_id)
SELECT 
    (SELECT id FROM files WHERE share_token = 'future999'),
    u.id
FROM users u
WHERE u.username IN ('nguyenvana', 'tranthib');

-- Xem ai được share file nào
SELECT 
    f.file_name,
    u.username,
    u.email
FROM shared_with sw
JOIN files f ON sw.file_id = f.id
JOIN users u ON sw.user_id = u.id
ORDER BY f.file_name, u.username;

-- Xem tất cả files mà một user được share
SELECT 
    f.file_name,
    f.share_token,
    owner.username as file_owner,
    f.available_to
FROM shared_with sw
JOIN files f ON sw.file_id = f.id
LEFT JOIN users owner ON f.owner_id = owner.id
WHERE sw.user_id = (SELECT id FROM users WHERE username = 'tranthib');

-- Đếm số user được share cho mỗi file
SELECT 
    f.file_name,
    COUNT(sw.user_id) as shared_with_count
FROM files f
LEFT JOIN shared_with sw ON f.id = sw.file_id
GROUP BY f.id, f.file_name
ORDER BY shared_with_count DESC;


-- ============================================
-- 4. SYSTEM_POLICY - Cấu hình hệ thống
-- ============================================

-- Xem policy hiện tại
SELECT * FROM system_policy WHERE id = 1;

-- Update max file size
UPDATE system_policy 
SET max_file_size_mb = 100 
WHERE id = 1;

-- Update validity defaults
UPDATE system_policy 
SET 
    min_validity_hours = 2,
    max_validity_days = 60,
    default_validity_days = 14
WHERE id = 1;


-- ============================================
-- 5. TRUY VẤN NÂNG CAO
-- ============================================

-- Files của user kèm thống kê sharing
SELECT 
    u.username,
    COUNT(DISTINCT f.id) as total_files,
    SUM(f.file_size) / 1024.0 / 1024.0 as total_size_mb,
    COUNT(DISTINCT sw.user_id) as total_shared_users
FROM users u
LEFT JOIN files f ON u.id = f.owner_id
LEFT JOIN shared_with sw ON f.id = sw.file_id
GROUP BY u.id, u.username;

-- Top 5 files lớn nhất
SELECT 
    f.file_name,
    f.file_size / 1024.0 / 1024.0 as size_mb,
    u.username as owner,
    f.created_at
FROM files f
LEFT JOIN users u ON f.owner_id = u.id
ORDER BY f.file_size DESC
LIMIT 5;

-- Files theo loại mime type
SELECT 
    f.mime_type,
    COUNT(*) as file_count,
    SUM(f.file_size) / 1024.0 / 1024.0 as total_size_mb
FROM files f
GROUP BY f.mime_type
ORDER BY file_count DESC;

-- Tìm file mà user có thể truy cập (owner hoặc được share)
CREATE OR REPLACE FUNCTION get_accessible_files(user_email CITEXT)
RETURNS TABLE (
    file_id UUID,
    file_name VARCHAR,
    share_token VARCHAR,
    access_type TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Files owned by user
    SELECT 
        f.id,
        f.file_name,
        f.share_token,
        'owner'::TEXT
    FROM files f
    JOIN users u ON f.owner_id = u.id
    WHERE u.email = user_email
    
    UNION
    
    -- Files shared with user
    SELECT 
        f.id,
        f.file_name,
        f.share_token,
        'shared'::TEXT
    FROM files f
    JOIN shared_with sw ON f.id = sw.file_id
    JOIN users u ON sw.user_id = u.id
    WHERE u.email = user_email;
END;
$$ LANGUAGE plpgsql;

-- Sử dụng function
SELECT * FROM get_accessible_files('nguyenvana@example.com');

-- Tìm file hết hạn và cần cleanup
SELECT 
    f.id,
    f.file_name,
    f.file_path,
    f.available_to,
    NOW() - f.available_to as expired_duration
FROM files f
WHERE f.available_to < NOW()
ORDER BY f.available_to DESC;

-- Thống kê theo ngày
SELECT 
    DATE(created_at) as date,
    COUNT(*) as files_uploaded,
    SUM(file_size) / 1024.0 / 1024.0 as total_mb
FROM files
GROUP BY DATE(created_at)
ORDER BY date DESC;


-- ============================================
-- 6. XÓA DỮ LIỆU
-- ============================================

-- Xóa file (tự động xóa shared_with do CASCADE)
DELETE FROM files WHERE share_token = 'abc123def456';

-- Xóa user (files của user sẽ có owner_id = NULL)
DELETE FROM users WHERE username = 'tranthib';

-- Xóa tất cả file hết hạn
DELETE FROM files 
WHERE available_to < NOW() - INTERVAL '30 days';

-- Revoke sharing
DELETE FROM shared_with 
WHERE file_id = (SELECT id FROM files WHERE share_token = 'future999')
    AND user_id = (SELECT id FROM users WHERE username = 'nguyenvana');
