package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/services/user-service/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) getTelegramID(c *fiber.Ctx) (int64, error) {
	tid := c.Get("X-Telegram-Id")
	if tid == "" {
		// fallback from auth body for direct calls
		return 0, fiber.ErrUnauthorized
	}
	return strconv.ParseInt(tid, 10, 64)
}

func (h *Handler) AuthTelegram(c *fiber.Ctx) error {
	var req struct {
		InitData string `json:"init_data"`
	}
	if err := c.BodyParser(&req); err != nil {
		req.InitData = c.Get("Authorization")
		if len(req.InitData) > 4 {
			req.InitData = req.InitData[4:] // strip "tma "
		}
	}
	resp, err := h.svc.AuthTelegram(c.Context(), req.InitData)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, resp)
}

func (h *Handler) GetMe(c *fiber.Ctx) error {
	tid, err := h.getTelegramID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	user, err := h.svc.GetMe(c.Context(), tid)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, user)
}

func (h *Handler) ListFavorites(c *fiber.Ctx) error {
	tid, err := h.getTelegramID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	ids, err := h.svc.ListFavorites(c.Context(), tid)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"product_ids": ids})
}

func (h *Handler) AddFavorite(c *fiber.Ctx) error {
	tid, err := h.getTelegramID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	pid, err := uuid.Parse(c.Params("productId"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.AddFavorite(c.Context(), tid, pid); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusCreated, fiber.Map{"ok": true})
}

func (h *Handler) RemoveFavorite(c *fiber.Ctx) error {
	tid, err := h.getTelegramID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	pid, err := uuid.Parse(c.Params("productId"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.RemoveFavorite(c.Context(), tid, pid); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ResolveByTelegramID(c *fiber.Ctx) error {
	tid, err := strconv.ParseInt(c.Params("telegramId"), 10, 64)
	if err != nil {
		return fiber.ErrBadRequest
	}
	user, err := h.svc.GetMe(c.Context(), tid)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, user)
}

func (h *Handler) ResolveByUserID(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	user, err := h.svc.GetUserByID(c.Context(), userID)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, user)
}

func (h *Handler) AdminLogin(c *fiber.Ctx) error {
	var req service.AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	resp, err := h.svc.AdminLogin(c.Context(), req)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, resp)
}

func (h *Handler) AdminRefresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	resp, err := h.svc.AdminRefresh(c.Context(), req.RefreshToken)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, resp)
}

func (h *Handler) AdminListUsers(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	users, err := h.svc.ListUsers(c.Context(), limit, offset)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"users": users})
}

func (h *Handler) AdminAuditLog(c *fiber.Ctx) error {
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"logs": []any{}})
}

func (h *Handler) GetReferrals(c *fiber.Ctx) error {
	tid, err := h.getTelegramID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	profile, err := h.svc.GetReferralProfile(c.Context(), tid)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, profile)
}

func (h *Handler) InternalValidateReferralPromo(c *fiber.Ctx) error {
	var req struct {
		UserID    string `json:"user_id"`
		PromoCode string `json:"promo_code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fiber.ErrBadRequest
	}
	promo, err := h.svc.ValidateReferralPromo(c.Context(), userID, req.PromoCode)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, promo)
}

func (h *Handler) InternalMarkReferralPromoUsed(c *fiber.Ctx) error {
	var req struct {
		RewardID string `json:"reward_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	rewardID, err := uuid.Parse(req.RewardID)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.MarkReferralPromoUsed(c.Context(), rewardID); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *Handler) InternalCompleteReferralOrder(c *fiber.Ctx) error {
	var req struct {
		UserID  string `json:"user_id"`
		OrderID string `json:"order_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fiber.ErrBadRequest
	}
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.CompleteReferralOrder(c.Context(), userID, orderID); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}
