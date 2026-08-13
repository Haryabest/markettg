package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markettg/markettg/packages/go-shared/pkg/telegram"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	TelegramID   int64     `json:"telegram_id"`
	Username     *string   `json:"username"`
	FirstName    *string   `json:"first_name"`
	LastName     *string   `json:"last_name"`
	LanguageCode *string   `json:"language_code"`
	PhotoURL     *string   `json:"photo_url"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
}

type AdminUser struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	TOTPEnabled  bool      `json:"totp_enabled"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) UpsertUser(ctx context.Context, tgUser *telegram.WebAppUser) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users.users (telegram_id, username, first_name, last_name, language_code, photo_url, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (telegram_id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			language_code = EXCLUDED.language_code,
			photo_url = EXCLUDED.photo_url,
			updated_at = NOW()
		RETURNING id, telegram_id, username, first_name, last_name, language_code, photo_url, is_admin, created_at`,
		tgUser.ID, nullStr(tgUser.Username), nullStr(tgUser.FirstName),
		nullStr(tgUser.LastName), nullStr(tgUser.LanguageCode), nullStr(tgUser.PhotoURL),
	).Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName,
		&u.LanguageCode, &u.PhotoURL, &u.IsAdmin, &u.CreatedAt)
	return &u, err
}

func (r *Repository) GetByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, language_code, photo_url, is_admin, created_at
		FROM users.users WHERE telegram_id = $1`, telegramID,
	).Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName,
		&u.LanguageCode, &u.PhotoURL, &u.IsAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, language_code, photo_url, is_admin, created_at
		FROM users.users WHERE id = $1`, id,
	).Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName,
		&u.LanguageCode, &u.PhotoURL, &u.IsAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) ListFavorites(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `SELECT product_id FROM users.favorites WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Repository) AddFavorite(ctx context.Context, userID, productID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users.favorites (user_id, product_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, userID, productID)
	return err
}

func (r *Repository) RemoveFavorite(ctx context.Context, userID, productID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users.favorites WHERE user_id = $1 AND product_id = $2`, userID, productID)
	return err
}

func (r *Repository) GetAdminByEmail(ctx context.Context, email string) (*AdminUser, error) {
	var a AdminUser
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, role::text, totp_enabled FROM users.admin_users WHERE email = $1`, email,
	).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.Role, &a.TOTPEnabled)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) UpdateAdminLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users.admin_users SET last_login = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) ListUsers(ctx context.Context, limit, offset int) ([]User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, language_code, photo_url, is_admin, created_at
		FROM users.users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName,
			&u.LanguageCode, &u.PhotoURL, &u.IsAdmin, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *Repository) AuditLog(ctx context.Context, actorID uuid.UUID, action, entityType, entityID string, payload any, ip string) error {
	data, _ := json.Marshal(payload)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users.admin_audit_log (actor_id, action, entity_type, entity_id, payload, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6::inet)`, actorID, action, entityType, entityID, data, ip)
	return err
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
