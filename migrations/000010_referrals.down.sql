DROP TABLE IF EXISTS users.referral_rewards;
DROP TABLE IF EXISTS users.referrals;
DROP INDEX IF EXISTS idx_users_referral_code;
ALTER TABLE users.users DROP COLUMN IF EXISTS referral_code;
