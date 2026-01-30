package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtTTL = 24 * time.Hour

type jwtClaims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

// LoginRequest тело запроса на вход
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginResponse ответ с токеном
type LoginResponse struct {
	Token string `json:"token"`
}

// checkAuth читает заголовок Authorization (Bearer <token>), проверяет JWT. При ошибке пишет в w и возвращает false.
func (h *Handlers) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "требуется авторизация"})
		return false
	}
	const prefix = "Bearer "
	if len(auth) < len(prefix) || auth[:len(prefix)] != prefix {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "неверный формат Authorization"})
		return false
	}
	tokenStr := auth[len(prefix):]
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "недействительный токен"})
		return false
	}
	return true
}

// Login godoc
// @Summary      Вход в API
// @Description  Проверяет логин и пароль, при успехе возвращает JWT для заголовка Authorization.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  LoginRequest  true  "Логин и пароль"
// @Success      200   {object}  LoginResponse
// @Failure      400   {object}  object  "Неверный запрос"
// @Failure      401   {object}  object  "Неверные логин или пароль"
// @Router       /api/login [post]
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	expectedLogin := os.Getenv("API_LOGIN")
	expectedPassword := os.Getenv("API_PASSWORD")
	if expectedLogin == "" {
		expectedLogin = os.Getenv("LOGIN")
	}
	if expectedPassword == "" {
		expectedPassword = os.Getenv("PASSWORD")
	}
	if expectedLogin == "" || expectedPassword == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "авторизация не настроена"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "неверный JSON"})
		return
	}
	if req.Login != expectedLogin || req.Password != expectedPassword {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "неверный логин или пароль"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	claims := jwtClaims{
		Login: req.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка выдачи токена"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Token: tokenStr})
}
