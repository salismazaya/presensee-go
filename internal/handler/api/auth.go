package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"presensee/internal/middleware"
	"presensee/internal/model"

	"gorm.io/gorm"
)

type AuthHandler struct {
	DB *gorm.DB
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Username/password salah"})
		return
	}

	normUsername := strings.ToLower(strings.TrimSpace(req.Username))
	var user model.User
	if err := h.DB.Where("LOWER(TRIM(username)) = ? AND is_active = true", normUsername).First(&user).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Username/password salah"})
		return
	}

	if !user.CheckPassword(req.Password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Username/password salah"})
		return
	}

	user.Token = model.GenerateToken(20)
	h.DB.Model(&user).Update("token", user.Token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{
			"token":    user.Token,
			"username": user.Username,
			"type":     user.Type,
		},
	})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Unauthorized"})
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Format request tidak valid"})
		return
	}

	if !user.CheckPassword(req.OldPassword) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Password salah!"})
		return
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Gagal menyimpan password baru"})
		return
	}

	h.DB.Model(user).Update("password", user.Password)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]bool{"success": true},
	})
}
