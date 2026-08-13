-- Default admin: admin@markettg.local / admin123 (change in production!)
INSERT INTO users.admin_users (email, password_hash, role)
VALUES (
  'admin@markettg.local',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
  'superadmin'
) ON CONFLICT (email) DO NOTHING;
