ALTER TABLE users.users
    ADD COLUMN IF NOT EXISTS referral_code VARCHAR(16);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_referral_code ON users.users(referral_code)
    WHERE referral_code IS NOT NULL;

CREATE TABLE IF NOT EXISTS users.referrals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    referrer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    referred_id UUID NOT NULL UNIQUE REFERENCES users.users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    referred_bonus_used BOOLEAN NOT NULL DEFAULT FALSE,
    referrer_rewarded BOOLEAN NOT NULL DEFAULT FALSE,
    first_order_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    qualified_at TIMESTAMPTZ,
    CHECK (referrer_id <> referred_id)
);

CREATE INDEX IF NOT EXISTS idx_referrals_referrer ON users.referrals(referrer_id);
CREATE INDEX IF NOT EXISTS idx_referrals_status ON users.referrals(status);

CREATE TABLE IF NOT EXISTS users.referral_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    referral_id UUID REFERENCES users.referrals(id) ON DELETE SET NULL,
    reward_type VARCHAR(32) NOT NULL,
    promo_code VARCHAR(32) NOT NULL UNIQUE,
    discount_type VARCHAR(16) NOT NULL,
    discount_value BIGINT NOT NULL,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_referral_rewards_user ON users.referral_rewards(user_id);
