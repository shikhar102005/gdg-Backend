package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/auth"
)

type AuthHandler struct {
	authSvc *auth.AuthService
}

func NewAuthHandler(authSvc *auth.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string        `json:"token"`
	ExpiresAt time.Time     `json:"expires_at"`
	Role      auth.UserRole `json:"role"`
	Email     string        `json:"email"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, exp, role, err := h.authSvc.AuthenticateOrganizer(req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	writeJSON(w, http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresAt: exp,
		Role:      role,
		Email:     req.Email,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := GetUserClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
	})
}
