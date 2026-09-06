package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/packages/go-shared/pkg/telegram"
	"github.com/markettg/markettg/services/user-service/internal/repository"
)

const (
	referredWelcomePercent = int64(5)
	referrerRewardKopecks  = int64(10000) // 100 ₽
)

type ReferralProfile struct {
	Code             string                    `json:"code"`
	Link             string                    `json:"link"`
	InvitedCount     int                       `json:"invited_count"`
	QualifiedCount   int                       `json:"qualified_count"`
	PendingCount     int                       `json:"pending_count"`
	TotalBonusKopecks int64                    `json:"total_bonus_kopecks"`
	Rewards          []repository.ReferralReward `json:"rewards"`
	BonusInfo        ReferralBonusInfo         `json:"bonus_info"`
}

type ReferralBonusInfo struct {
	ReferredWelcome string `json:"referred_welcome"`
	ReferrerReward  string `json:"referrer_reward"`
}

type PromoValidation struct {
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	RewardID      string `json:"reward_id"`
}

func (s *Service) authTelegramWithReferral(ctx context.Context, initData string) (*AuthResponse, error) {
	data, err := telegram.ValidateInitData(initData, s.botToken, 24*time.Hour)
	if err != nil || data.User == nil {
		return nil, apperrors.ErrUnauthorized
	}

	existed, err := s.repo.UserExistsBeforeUpsert(ctx, data.User.ID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	user, err := s.repo.UpsertUser(ctx, data.User)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	if _, err := s.repo.EnsureReferralCode(ctx, user.ID); err != nil {
		return nil, apperrors.ErrInternal
	}

	if !existed {
		s.attachReferral(ctx, user.ID, data.StartParam)
	} else if data.StartParam != "" {
		ref, _ := s.repo.GetReferralByReferredID(ctx, user.ID)
		if ref == nil {
			s.attachReferral(ctx, user.ID, data.StartParam)
		}
	}

	return &AuthResponse{User: user}, nil
}

func (s *Service) attachReferral(ctx context.Context, referredID uuid.UUID, startParam string) {
	code := parseReferralCode(startParam)
	if code == "" {
		return
	}
	referrer, err := s.repo.GetByReferralCode(ctx, code)
	if err != nil || referrer.ID == referredID {
		return
	}
	if err := s.repo.CreateReferral(ctx, referrer.ID, referredID); err != nil {
		return
	}
	ref, err := s.repo.GetReferralByReferredID(ctx, referredID)
	if err != nil || ref == nil {
		return
	}
	expires := time.Now().Add(30 * 24 * time.Hour)
	promo := "FRIEND" + strings.ToUpper(code)
	_, _ = s.repo.CreateReferralReward(ctx, repository.ReferralReward{
		UserID:        referredID,
		ReferralID:    &ref.ID,
		RewardType:    "referred_welcome",
		PromoCode:     promo,
		DiscountType:  "PERCENT",
		DiscountValue: referredWelcomePercent,
		ExpiresAt:     &expires,
	})
}

func parseReferralCode(startParam string) string {
	startParam = strings.TrimSpace(startParam)
	if startParam == "" {
		return ""
	}
	if strings.HasPrefix(startParam, "ref_") {
		return strings.TrimPrefix(startParam, "ref_")
	}
	if len(startParam) >= 6 && len(startParam) <= 16 {
		return startParam
	}
	return ""
}

func (s *Service) GetReferralProfile(ctx context.Context, telegramID int64) (*ReferralProfile, error) {
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		user, err = s.repo.UpsertUser(ctx, &telegram.WebAppUser{ID: telegramID, FirstName: "User"})
		if err != nil {
			return nil, apperrors.ErrInternal
		}
		if _, err := s.repo.EnsureReferralCode(ctx, user.ID); err != nil {
			return nil, apperrors.ErrInternal
		}
	}
	stats, err := s.repo.GetReferralStats(ctx, user.ID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	rewards, err := s.repo.ListReferralRewards(ctx, user.ID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	botUser := s.botUsername
	link := ""
	if botUser != "" {
		link = "https://t.me/" + botUser + "?start=ref_" + stats.Code
	}
	return &ReferralProfile{
		Code:              stats.Code,
		Link:              link,
		InvitedCount:      stats.InvitedCount,
		QualifiedCount:    stats.QualifiedCount,
		PendingCount:      stats.PendingCount,
		TotalBonusKopecks: stats.TotalBonusKopecks,
		Rewards:           rewards,
		BonusInfo: ReferralBonusInfo{
			ReferredWelcome: "5% на первый заказ друга",
			ReferrerReward:  "100 ₽ вам после первой покупки друга",
		},
	}, nil
}

func (s *Service) ValidateReferralPromo(ctx context.Context, userID uuid.UUID, code string) (*PromoValidation, error) {
	rw, err := s.repo.GetReferralRewardByCode(ctx, userID, code)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if rw == nil {
		return nil, apperrors.ErrInvalidPromoCode
	}
	return &PromoValidation{
		DiscountType:  rw.DiscountType,
		DiscountValue: rw.DiscountValue,
		RewardID:      rw.ID.String(),
	}, nil
}

func (s *Service) MarkReferralPromoUsed(ctx context.Context, rewardID uuid.UUID) error {
	return s.repo.MarkReferralRewardUsed(ctx, rewardID)
}

func (s *Service) CompleteReferralOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	return s.repo.QualifyReferralOnOrder(ctx, userID, orderID)
}
