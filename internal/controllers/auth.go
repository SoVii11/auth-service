package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/SoVii11/auth-service/internal/usecases"
	"github.com/SoVii11/shared/pkg/response"
	"go.uber.org/zap"
)

type AuthController struct {
	usecase *usecases.AuthUsecase
	log     *zap.Logger
}

func NewAuthController(usecase *usecases.AuthUsecase, log *zap.Logger) *AuthController {
	return &AuthController{usecase: usecase, log: log}
}

type registerRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

type loginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

type resetCodeRequest struct {
	Email string `json:"email" example:"user@example.com"`
}

type resetPasswordRequest struct {
	Email       string `json:"email" example:"user@example.com"`
	Code        string `json:"code" example:"123456"`
	NewPassword string `json:"new_password" example:"newpassword123"`
}

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Создаёт нового пользователя по email и паролю
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      registerRequest  true  "Данные для регистрации"
// @Success      201    {object}  map[string]any
// @Failure      400    {object}  map[string]string
// @Router       /auth/register [post]
func (a *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	user, err := a.usecase.Register(req.Email, req.Password)
	if err != nil {
		a.log.Warn("register failed", zap.Error(err))
		response.BadRequest(w, err.Error())
		return
	}

	a.log.Info("user registered", zap.String("email", user.Email))
	response.Created(w, map[string]any{
		"id":    user.ID,
		"email": user.Email,
	})
}

// Login godoc
// @Summary      Вход в аккаунт
// @Description  Возвращает JWT токен при успешной авторизации
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      loginRequest  true  "Данные для входа"
// @Success      200    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Router       /auth/login [post]
func (a *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	token, err := a.usecase.Login(req.Email, req.Password)
	if err != nil {
		a.log.Warn("login failed", zap.String("email", req.Email))
		response.Unauthorized(w, err.Error())
		return
	}

	a.log.Info("user logged in", zap.String("email", req.Email))
	response.Success(w, map[string]string{"token": token})
}

// SendResetCode godoc
// @Summary      Отправка кода восстановления
// @Description  Отправляет 6-значный код на email для сброса пароля
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      resetCodeRequest  true  "Email пользователя"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Router       /auth/send-reset-code [post]
func (a *AuthController) SendResetCode(w http.ResponseWriter, r *http.Request) {
	var req resetCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := a.usecase.SendResetCode(req.Email); err != nil {
		a.log.Warn("send reset code failed", zap.Error(err))
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "code sent to email"})
}

// ResetPassword godoc
// @Summary      Сброс пароля
// @Description  Сбрасывает пароль пользователя по коду из email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      resetPasswordRequest  true  "Данные для сброса пароля"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Router       /auth/reset-password [post]
func (a *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := a.usecase.ResetPassword(req.Email, req.Code, req.NewPassword); err != nil {
		a.log.Warn("reset password failed", zap.Error(err))
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "password reset successful"})
}
