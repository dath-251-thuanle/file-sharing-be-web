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

INSERT INTO users (username, email, role, password_hash) 
VALUES ('admin', 'admin@example.com', 'admin', '$2b$12$hash3');

INSERT INTO users (username, email, password_hash, totp_secret, totp_enabled) 
VALUES ('secureuser', 'secure@example.com', '$2b$12$hash4', 'JBSWY3DPEHPK3PXP', TRUE);

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
-- 6. FILE STATISTICS - Thống kê file
-- ============================================

-- Khởi tạo statistics cho file (chỉ cho files có owner)
INSERT INTO file_statistics (file_id, download_count, unique_downloaders, view_count)
SELECT 
    f.id,
    0,
    0,
    0
FROM files f
WHERE f.owner_id IS NOT NULL
ON CONFLICT (file_id) DO NOTHING;

-- Xem statistics của tất cả files
SELECT 
    f.file_name,
    u.username as owner,
    COALESCE(fs.download_count, 0) as downloads,
    COALESCE(fs.unique_downloaders, 0) as unique_users,
    COALESCE(fs.view_count, 0) as views,
    fs.last_downloaded_at,
    fs.last_viewed_at
FROM files f
LEFT JOIN users u ON f.owner_id = u.id
LEFT JOIN file_statistics fs ON f.id = fs.file_id
ORDER BY fs.download_count DESC NULLS LAST;

-- Top 10 files có nhiều downloads nhất
SELECT 
    f.file_name,
    f.share_token,
    u.username as owner,
    fs.download_count,
    fs.unique_downloaders,
    fs.view_count,
    fs.last_downloaded_at
FROM file_statistics fs
JOIN files f ON fs.file_id = f.id
LEFT JOIN users u ON f.owner_id = u.id
ORDER BY fs.download_count DESC
LIMIT 10;

-- Statistics của files của một user cụ thể
SELECT 
    f.file_name,
    f.share_token,
    f.created_at,
    COALESCE(fs.download_count, 0) as downloads,
    COALESCE(fs.view_count, 0) as views,
    COALESCE(fs.unique_downloaders, 0) as unique_users,
    fs.last_downloaded_at
FROM files f
LEFT JOIN file_statistics fs ON f.id = fs.file_id
WHERE f.owner_id = (SELECT id FROM users WHERE username = 'nguyenvana')
ORDER BY fs.download_count DESC NULLS LAST;

-- Tổng hợp statistics theo user
SELECT 
    u.username,
    u.email,
    COUNT(DISTINCT f.id) as total_files,
    COALESCE(SUM(fs.download_count), 0) as total_downloads,
    COALESCE(SUM(fs.view_count), 0) as total_views,
    COALESCE(SUM(fs.unique_downloaders), 0) as total_unique_downloaders
FROM users u
LEFT JOIN files f ON u.id = f.owner_id
LEFT JOIN file_statistics fs ON f.id = fs.file_id
GROUP BY u.id, u.username, u.email
ORDER BY total_downloads DESC;


-- ============================================
-- 7. DOWNLOAD HISTORY - Simple download tracking (browser-style)
-- ============================================

-- Record download from authenticated user
INSERT INTO download_history (
    file_id, 
    downloader_id, 
    download_completed
)
VALUES (
    (SELECT id FROM files WHERE share_token = 'abc123def456'),
    (SELECT id FROM users WHERE username = 'tranthib'),
    TRUE
);

-- Record download from anonymous user
INSERT INTO download_history (
    file_id,
    downloader_id,
    download_completed
)
VALUES (
    (SELECT id FROM files WHERE share_token = 'abc123def456'),
    NULL,  -- Anonymous download
    TRUE
);

-- Record failed download (interrupted)
INSERT INTO download_history (
    file_id,
    downloader_id,
    download_completed
)
VALUES (
    (SELECT id FROM files WHERE share_token = 'xyz789ghi012'),
    (SELECT id FROM users WHERE username = 'secureuser'),
    FALSE  -- Download was interrupted
);

-- Record multiple downloads for testing
INSERT INTO download_history (file_id, downloader_id, download_completed)
SELECT 
    (SELECT id FROM files WHERE share_token = 'abc123def456'),
    u.id,
    TRUE
FROM users u
WHERE u.username IN ('nguyenvana', 'tranthib', 'secureuser');

-- View all download history
SELECT 
    f.file_name,
    f.share_token,
    COALESCE(u.username, 'Anonymous') as downloader,
    COALESCE(u.email, 'Anonymous') as downloader_email,
    dh.downloaded_at,
    dh.download_completed
FROM download_history dh
JOIN files f ON dh.file_id = f.id
LEFT JOIN users u ON dh.downloader_id = u.id
ORDER BY dh.downloaded_at DESC
LIMIT 50;

-- Download history for a specific file
-- API equivalent: GET /files/:id/download-history
SELECT 
    dh.downloaded_at,
    COALESCE(u.username, 'Anonymous') as downloader,
    u.email as downloader_email,
    dh.download_completed
FROM download_history dh
LEFT JOIN users u ON dh.downloader_id = u.id
WHERE dh.file_id = (SELECT id FROM files WHERE share_token = 'abc123def456')
ORDER BY dh.downloaded_at DESC;

-- Download history for a user (what files they downloaded)
SELECT 
    f.file_name,
    f.share_token,
    owner.username as file_owner,
    dh.downloaded_at,
    dh.download_completed
FROM download_history dh
JOIN files f ON dh.file_id = f.id
LEFT JOIN users owner ON f.owner_id = owner.id
WHERE dh.downloader_id = (SELECT id FROM users WHERE username = 'tranthib')
ORDER BY dh.downloaded_at DESC;

-- Count downloads by day (last 7 days)
SELECT 
    DATE(dh.downloaded_at) as date,
    COUNT(*) as total_downloads,
    COUNT(DISTINCT dh.downloader_id) as unique_users,
    COUNT(CASE WHEN dh.downloader_id IS NULL THEN 1 END) as anonymous_downloads,
    COUNT(CASE WHEN dh.download_completed = FALSE THEN 1 END) as failed_downloads
FROM download_history dh
WHERE dh.downloaded_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(dh.downloaded_at)
ORDER BY date DESC;

-- Top downloaders (users with most downloads)
SELECT 
    u.username,
    u.email,
    COUNT(*) as total_downloads,
    COUNT(DISTINCT dh.file_id) as unique_files_downloaded,
    COUNT(CASE WHEN dh.download_completed = FALSE THEN 1 END) as failed_downloads,
    MAX(dh.downloaded_at) as last_download
FROM download_history dh
JOIN users u ON dh.downloader_id = u.id
GROUP BY u.id, u.username, u.email
ORDER BY total_downloads DESC
LIMIT 10;

-- Most downloaded files
SELECT 
    f.file_name,
    f.share_token,
    COUNT(*) as total_downloads,
    COUNT(DISTINCT dh.downloader_id) FILTER (WHERE dh.downloader_id IS NOT NULL) as unique_users,
    COUNT(*) FILTER (WHERE dh.downloader_id IS NULL) as anonymous_downloads,
    COUNT(*) FILTER (WHERE dh.download_completed = FALSE) as failed_downloads
FROM download_history dh
JOIN files f ON dh.file_id = f.id
GROUP BY f.id, f.file_name, f.share_token
ORDER BY total_downloads DESC
LIMIT 10;

-- Download activity by hour of day
SELECT 
    EXTRACT(HOUR FROM dh.downloaded_at) as hour_of_day,
    COUNT(*) as downloads,
    COUNT(CASE WHEN dh.download_completed = FALSE THEN 1 END) as failed
FROM download_history dh
GROUP BY hour_of_day
ORDER BY hour_of_day;


-- ============================================
-- 8. ANALYTICS - Phân tích nâng cao
-- ============================================

-- Dashboard overview cho owner
SELECT 
    (SELECT COUNT(*) FROM files WHERE owner_id = (SELECT id FROM users WHERE username = 'nguyenvana')) as total_files,
    (SELECT COALESCE(SUM(download_count), 0) FROM file_statistics fs 
     JOIN files f ON fs.file_id = f.id 
     WHERE f.owner_id = (SELECT id FROM users WHERE username = 'nguyenvana')) as total_downloads,
    (SELECT COALESCE(SUM(view_count), 0) FROM file_statistics fs 
     JOIN files f ON fs.file_id = f.id 
     WHERE f.owner_id = (SELECT id FROM users WHERE username = 'nguyenvana')) as total_views,
    (SELECT COUNT(*) FROM download_history dh 
     JOIN files f ON dh.file_id = f.id 
     WHERE f.owner_id = (SELECT id FROM users WHERE username = 'nguyenvana')
     AND dh.downloaded_at >= NOW() - INTERVAL '24 hours') as downloads_last_24h;

-- File popularity score (downloads + views weighted)
SELECT 
    f.file_name,
    f.share_token,
    u.username as owner,
    COALESCE(fs.download_count, 0) as downloads,
    COALESCE(fs.view_count, 0) as views,
    (COALESCE(fs.download_count, 0) * 2 + COALESCE(fs.view_count, 0)) as popularity_score
FROM files f
LEFT JOIN users u ON f.owner_id = u.id
LEFT JOIN file_statistics fs ON f.id = fs.file_id
WHERE f.owner_id IS NOT NULL
ORDER BY popularity_score DESC
LIMIT 20;

-- Conversion rate (views to downloads)
SELECT 
    f.file_name,
    fs.view_count,
    fs.download_count,
    CASE 
        WHEN fs.view_count > 0 THEN ROUND(fs.download_count::NUMERIC / fs.view_count * 100, 2)
        ELSE 0 
    END as conversion_rate_percent
FROM files f
JOIN file_statistics fs ON f.id = fs.file_id
WHERE fs.view_count > 0
ORDER BY conversion_rate_percent DESC;

-- Recently active files (downloaded or viewed in last 7 days)
SELECT 
    f.file_name,
    f.share_token,
    u.username as owner,
    fs.download_count,
    fs.view_count,
    GREATEST(
        COALESCE(fs.last_downloaded_at, '1970-01-01'::timestamp), 
        COALESCE(fs.last_viewed_at, '1970-01-01'::timestamp)
    ) as last_activity,
    CASE 
        WHEN COALESCE(fs.last_downloaded_at, '1970-01-01'::timestamp) > COALESCE(fs.last_viewed_at, '1970-01-01'::timestamp) THEN 'download'
        WHEN COALESCE(fs.last_viewed_at, '1970-01-01'::timestamp) > COALESCE(fs.last_downloaded_at, '1970-01-01'::timestamp) THEN 'view'
        ELSE 'unknown'
    END as last_activity_type
FROM files f
LEFT JOIN users u ON f.owner_id = u.id
JOIN file_statistics fs ON f.id = fs.file_id
WHERE GREATEST(
    COALESCE(fs.last_downloaded_at, '1970-01-01'::timestamp), 
    COALESCE(fs.last_viewed_at, '1970-01-01'::timestamp)
) >= NOW() - INTERVAL '7 days'
ORDER BY last_activity DESC;

-- Success rate analysis (completed vs failed downloads)
SELECT 
    f.file_name,
    f.share_token,
    COUNT(*) as total_attempts,
    COUNT(*) FILTER (WHERE dh.download_completed = TRUE) as successful,
    COUNT(*) FILTER (WHERE dh.download_completed = FALSE) as failed,
    ROUND(
        COUNT(*) FILTER (WHERE dh.download_completed = TRUE)::NUMERIC / 
        NULLIF(COUNT(*), 0) * 100, 
        2
    ) as success_rate_percent
FROM download_history dh
JOIN files f ON dh.file_id = f.id
GROUP BY f.id, f.file_name, f.share_token
HAVING COUNT(*) > 0
ORDER BY total_attempts DESC;


-- ============================================
-- 9. MAINTENANCE - Cleanup và maintenance
-- ============================================

-- Xóa download history cũ (older than 90 days)
DELETE FROM download_history 
WHERE downloaded_at < NOW() - INTERVAL '90 days';

-- Xóa statistics của files đã bị xóa (orphaned records)
DELETE FROM file_statistics
WHERE file_id NOT IN (SELECT id FROM files);

-- Recalculate unique_downloaders cho một file
UPDATE file_statistics
SET unique_downloaders = (
    SELECT COUNT(DISTINCT downloader_id)
    FROM download_history
    WHERE file_id = file_statistics.file_id
        AND downloader_id IS NOT NULL
)
WHERE file_id = (SELECT id FROM files WHERE share_token = 'abc123def456');

-- Reset statistics (if needed)
UPDATE file_statistics
SET download_count = 0,
    view_count = 0,
    unique_downloaders = 0,
    last_downloaded_at = NULL,
    last_viewed_at = NULL,
    updated_at = NOW()
WHERE file_id = (SELECT id FROM files WHERE share_token = 'xyz789ghi012');


-- ============================================
-- 10. TESTING SCENARIOS
-- ============================================

-- Scenario 1: File với nhiều views nhưng ít downloads (people just browsing)
-- Tạo views
INSERT INTO file_statistics (file_id, view_count, last_viewed_at, updated_at)
VALUES (
    (SELECT id FROM files WHERE share_token = 'abc123def456'),
    50,
    NOW(),
    NOW()
)
ON CONFLICT (file_id) 
DO UPDATE SET 
    view_count = file_statistics.view_count + 50,
    last_viewed_at = NOW(),
    updated_at = NOW();

-- Scenario 2: Simulate viral file (many downloads from anonymous users)
INSERT INTO download_history (file_id, downloader_id, download_completed)
SELECT 
    (SELECT id FROM files WHERE share_token = 'abc123def456'),
    NULL,  -- Anonymous
    TRUE
FROM generate_series(1, 20);

-- Scenario 3: Check statistics accuracy
-- Compare download_history count with file_statistics.download_count
SELECT 
    f.file_name,
    COALESCE(fs.download_count, 0) as stats_count,
    (SELECT COUNT(*) FROM download_history WHERE file_id = f.id) as history_count,
    CASE 
        WHEN COALESCE(fs.download_count, 0) = (SELECT COUNT(*) FROM download_history WHERE file_id = f.id) 
        THEN '✓ Match' 
        ELSE '✗ Mismatch' 
    END as status
FROM files f
LEFT JOIN file_statistics fs ON f.id = fs.file_id
WHERE f.owner_id IS NOT NULL;


-- ============================================
-- 11. XÓA DỮ LIỆU
-- ============================================

-- Xóa file (tự động xóa shared_with, file_statistics, download_history do CASCADE)
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
