package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Referral struct {
	ID                  uuid.UUID  `json:"id"`
	ReferrerID          uuid.UUID  `json:"referrer_id"`
	ReferredID          uuid.UUID  `json:"referred_id"`
	Status              string     `json:"status"`
	ReferredBonusUsed   bool       `json:"referred_bonus_used"`
	ReferrerRewarded    bool       `json:"referrer_rewarded"`
	FirstOrderID        *uuid.UUID `json:"first_order_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	QualifiedAt         *time.Time `json:"qualified_at,omitempty"`
}

type ReferralReward struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	ReferralID    *uuid.UUID `json:"referral_id,omitempty"`
	RewardType    string     `json:"reward_type"`
	PromoCode     string     `json:"promo_code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue int64      `json:"discount_value"`
	IsUsed        bool       `json:"is_used"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ReferralStats struct {
	Code            string `json:"code"`
	InvitedCount    int    `json:"invited_count"`
	QualifiedCount  int    `json:"qualified_count"`
	PendingCount    int    `json:"pending_count"`
	TotalBonusKopecks int64 `json:"total_bonus_kopecks"`
	BotUsername     string `json:"bot_username,omitempty"`
}

func referralCodeForUser(userID uuid.UUID) string {
	sum := sha256.Sum256([]byte(userID.String()))
	return strings.ToUpper(hex.EncodeToString(sum[:4]))
}

func (r *Repository) EnsureReferralCode(ctx context.Context, userID uuid.UUID) (string, error) {
	var code string
	err := r.pool.QueryRow(ctx, `SELECT referral_code FROM users.users WHERE id = $1`, userID).Scan(&code)
	if err != nil {
		return "", err
	}
	if code != "" {
		return code, nil
	}
	code = referralCodeForUser(userID)
	_, err = r.pool.Exec(ctx, `
		UPDATE users.users SET referral_code = $2, updated_at = NOW()
		WHERE id = $1 AND referral_code IS NULL`, userID, code)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (r *Repository) GetByReferralCode(ctx context.Context, code string) (*User, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, language_code, photo_url, is_admin, created_at
		FROM users.users WHERE referral_code = $1`, code,
	).Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName,
		&u.LanguageCode, &u.PhotoURL, &u.IsAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateReferral(ctx context.Context, referrerID, referredID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users.referrals (referrer_id, referred_id)
		VALUES ($1, $2)
		ON CONFLICT (referred_id) DO NOTHING`, referrerID, referredID)
	return err
}

func (r *Repository) GetReferralStats(ctx context.Context, userID uuid.UUID) (*ReferralStats, error) {
	code, err := r.EnsureReferralCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats := &ReferralStats{Code: code}
	err = r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE status = 'qualified')::int,
			COUNT(*) FILTER (WHERE status = 'pending')::int
		FROM users.referrals WHERE referrer_id = $1`, userID,
	).Scan(&stats.InvitedCount, &stats.QualifiedCount, &stats.PendingCount)
	if err != nil {
		return nil, err
	}
	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(discount_value), 0)
		FROM users.referral_rewards
		WHERE user_id = $1 AND reward_type = 'referrer_reward'`, userID,
	).Scan(&stats.TotalBonusKopecks)
	return stats, nil
}

func (r *Repository) ListReferralRewards(ctx context.Context, userID uuid.UUID) ([]ReferralReward, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, referral_id, reward_type, promo_code, discount_type, discount_value, is_used, expires_at, created_at
		FROM users.referral_rewards
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rewards []ReferralReward
	for rows.Next() {
		var rw ReferralReward
		if err := rows.Scan(&rw.ID, &rw.UserID, &rw.ReferralID, &rw.RewardType, &rw.PromoCode,
			&rw.DiscountType, &rw.DiscountValue, &rw.IsUsed, &rw.ExpiresAt, &rw.CreatedAt); err != nil {
			return nil, err
		}
		rewards = append(rewards, rw)
	}
	return rewards, nil
}

func (r *Repository) GetActiveReferredBonus(ctx context.Context, userID uuid.UUID) (*ReferralReward, error) {
	var rw ReferralReward
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, referral_id, reward_type, promo_code, discount_type, discount_value, is_used, expires_at, created_at
		FROM users.referral_rewards
		WHERE user_id = $1 AND reward_type = 'referred_welcome' AND is_used = FALSE
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC LIMIT 1`, userID,
	).Scan(&rw.ID, &rw.UserID, &rw.ReferralID, &rw.RewardType, &rw.PromoCode,
		&rw.DiscountType, &rw.DiscountValue, &rw.IsUsed, &rw.ExpiresAt, &rw.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rw, nil
}

func (r *Repository) GetReferralRewardByCode(ctx context.Context, userID uuid.UUID, code string) (*ReferralReward, error) {
	var rw ReferralReward
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, referral_id, reward_type, promo_code, discount_type, discount_value, is_used, expires_at, created_at
		FROM users.referral_rewards
		WHERE user_id = $1 AND promo_code = $2 AND is_used = FALSE
		  AND (expires_at IS NULL OR expires_at > NOW())`, userID, strings.ToUpper(code),
	).Scan(&rw.ID, &rw.UserID, &rw.ReferralID, &rw.RewardType, &rw.PromoCode,
		&rw.DiscountType, &rw.DiscountValue, &rw.IsUsed, &rw.ExpiresAt, &rw.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rw, nil
}

func (r *Repository) MarkReferralRewardUsed(ctx context.Context, rewardID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users.referral_rewards SET is_used = TRUE WHERE id = $1`, rewardID)
	return err
}

func (r *Repository) CreateReferralReward(ctx context.Context, rw ReferralReward) (*ReferralReward, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users.referral_rewards (user_id, referral_id, reward_type, promo_code, discount_type, discount_value, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, referral_id, reward_type, promo_code, discount_type, discount_value, is_used, expires_at, created_at`,
		rw.UserID, rw.ReferralID, rw.RewardType, rw.PromoCode, rw.DiscountType, rw.DiscountValue, rw.ExpiresAt,
	).Scan(&rw.ID, &rw.UserID, &rw.ReferralID, &rw.RewardType, &rw.PromoCode,
		&rw.DiscountType, &rw.DiscountValue, &rw.IsUsed, &rw.ExpiresAt, &rw.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rw, nil
}

func (r *Repository) QualifyReferralOnOrder(ctx context.Context, referredUserID, orderID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var ref Referral
	err = tx.QueryRow(ctx, `
		SELECT id, referrer_id, referred_id, status, referred_bonus_used, referrer_rewarded
		FROM users.referrals
		WHERE referred_id = $1
		FOR UPDATE`, referredUserID,
	).Scan(&ref.ID, &ref.ReferrerID, &ref.ReferredID, &ref.Status, &ref.ReferredBonusUsed, &ref.ReferrerRewarded)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	if ref.Status != "qualified" {
		_, err = tx.Exec(ctx, `
			UPDATE users.referrals
			SET status = 'qualified', qualified_at = NOW(), first_order_id = $2, referred_bonus_used = TRUE
			WHERE id = $1`, ref.ID, orderID)
		if err != nil {
			return err
		}
		_, _ = tx.Exec(ctx, `
			UPDATE users.referral_rewards SET is_used = TRUE
			WHERE user_id = $1 AND reward_type = 'referred_welcome' AND is_used = FALSE`, referredUserID)
	}

	if !ref.ReferrerRewarded {
		promo := fmt.Sprintf("REF%s", strings.ToUpper(hex.EncodeToString(orderID[:4])))
		expires := time.Now().Add(90 * 24 * time.Hour)
		_, err = tx.Exec(ctx, `
			INSERT INTO users.referral_rewards (user_id, referral_id, reward_type, promo_code, discount_type, discount_value, expires_at)
			VALUES ($1, $2, 'referrer_reward', $3, 'FIXED', 10000, $4)`,
			ref.ReferrerID, ref.ID, promo, expires)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE users.referrals SET referrer_rewarded = TRUE WHERE id = $1`, ref.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetReferralByReferredID(ctx context.Context, referredID uuid.UUID) (*Referral, error) {
	var ref Referral
	err := r.pool.QueryRow(ctx, `
		SELECT id, referrer_id, referred_id, status, referred_bonus_used, referrer_rewarded, first_order_id, created_at, qualified_at
		FROM users.referrals WHERE referred_id = $1`, referredID,
	).Scan(&ref.ID, &ref.ReferrerID, &ref.ReferredID, &ref.Status, &ref.ReferredBonusUsed,
		&ref.ReferrerRewarded, &ref.FirstOrderID, &ref.CreatedAt, &ref.QualifiedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

func (r *Repository) UserExistsBeforeUpsert(ctx context.Context, telegramID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users.users WHERE telegram_id = $1)`, telegramID).Scan(&exists)
	return exists, err
}
