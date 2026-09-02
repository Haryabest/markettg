package service

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/markettg/markettg/services/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo         *repository.Repository
	jwtSecret    string
	botToken     string
	botUsername  string
}

func New(repo *repository.Repository, jwtSecret string) *Service {
	return &Service{
		repo:        repo,
		jwtSecret:   jwtSecret,
		botToken:    config.GetEnv("TELEGRAM_BOT_TOKEN", ""),
		botUsername: strings.TrimPrefix(config.GetEnv("TELEGRAM_BOT_USERNAME", ""), "@"),
	}
}

type AuthResponse struct {
	User  *repository.User `json:"user"`
	Token string           `json:"token,omitempty"`
}

func (s *Service) AuthTelegram(ctx context.Context, initData string) (*AuthResponse, error) {
	return s.authTelegramWithReferral(ctx, initData)
}

func (s *Service) GetMe(ctx context.Context, telegramID int64) (*repository.User, error) {
	return s.repo.GetByTelegramID(ctx, telegramID)
}

func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*repository.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *Service) ListFavorites(ctx context.Context, telegramID int64) ([]uuid.UUID, error) {
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, apperrors.ErrNotFound
	}
	return s.repo.ListFavorites(ctx, user.ID)
}

func (s *Service) AddFavorite(ctx context.Context, telegramID int64, productID uuid.UUID) error {
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return apperrors.ErrNotFound
	}
	return s.repo.AddFavorite(ctx, user.ID, productID)
}

func (s *Service) RemoveFavorite(ctx context.Context, telegramID int64, productID uuid.UUID) error {
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return apperrors.ErrNotFound
	}
	return s.repo.RemoveFavorite(ctx, user.ID, productID)
}

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code"`
}

type AdminTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (s *Service) AdminLogin(ctx context.Context, req AdminLoginRequest) (*AdminTokenResponse, error) {
	admin, err := s.repo.GetAdminByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	_ = s.repo.UpdateAdminLastLogin(ctx, admin.ID)
	return s.generateAdminTokens(admin)
}

func (s *Service) generateAdminTokens(admin *repository.AdminUser) (*AdminTokenResponse, error) {
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  admin.ID.String(),
		"role": admin.Role,
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	})
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  admin.ID.String(),
		"type": "refresh",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	accessStr, err := access.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	refreshStr, err := refresh.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	return &AdminTokenResponse{AccessToken: accessStr, RefreshToken: refreshStr, ExpiresIn: 900}, nil
}

func (s *Service) AdminRefresh(ctx context.Context, refreshToken string) (*AdminTokenResponse, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, apperrors.ErrUnauthorized
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims["type"] != "refresh" {
		return nil, apperrors.ErrUnauthorized
	}
	adminID, _ := uuid.Parse(claims["sub"].(string))
	admin, err := s.repo.GetAdminByEmail(ctx, "")
	_ = adminID
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	return s.generateAdminTokens(admin)
}

func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]repository.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListUsers(ctx, limit, offset)
}
