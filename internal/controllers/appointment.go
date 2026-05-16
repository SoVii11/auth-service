package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/SoVii11/auth-service/internal/usecases"
	sharedJWT "github.com/SoVii11/shared/pkg/jwt"
	"github.com/SoVii11/shared/pkg/response"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type AppointmentController struct {
	usecase   *usecases.AppointmentUsecase
	log       *zap.Logger
	jwtSecret string
}

func NewAppointmentController(usecase *usecases.AppointmentUsecase, log *zap.Logger, jwtSecret string) *AppointmentController {
	return &AppointmentController{usecase: usecase, log: log, jwtSecret: jwtSecret}
}

type createAppointmentRequest struct {
	PsychologistID int64  `json:"psychologist_id" example:"1"`
	Comment        string `json:"comment" example:"Хочу записаться на консультацию"`
}

func (c *AppointmentController) getClaimsFromRequest(r *http.Request) (*sharedJWT.Claims, error) {
	tokenStr := r.Header.Get("Authorization")
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}
	return sharedJWT.ParseToken(tokenStr, c.jwtSecret)
}

// CreateAppointment godoc
// @Summary      Записаться к психологу
// @Description  Создаёт запись к психологу для авторизованного пользователя
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                    true  "Bearer токен"
// @Param        input          body      createAppointmentRequest  true  "Данные записи"
// @Success      201            {object}  map[string]any
// @Failure      400            {object}  map[string]string
// @Failure      401            {object}  map[string]string
// @Router       /appointments [post]
func (c *AppointmentController) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	claims, err := c.getClaimsFromRequest(r)
	if err != nil {
		response.Unauthorized(w, "unauthorized")
		return
	}

	var req createAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	a, err := c.usecase.CreateAppointment(claims.UserID, req.PsychologistID, req.Comment)
	if err != nil {
		c.log.Warn("create appointment failed", zap.Error(err))
		response.BadRequest(w, err.Error())
		return
	}

	c.log.Info("appointment created", zap.Int64("user_id", claims.UserID))
	response.Created(w, map[string]any{
		"id":              a.ID,
		"psychologist_id": a.PsychologistID,
		"status":          a.Status,
		"comment":         a.Comment,
	})
}

// GetMyAppointments godoc
// @Summary      Мои записи
// @Description  Возвращает список записей авторизованного пользователя
// @Tags         appointments
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer токен"
// @Success      200            {object}  map[string]any
// @Failure      401            {object}  map[string]string
// @Router       /appointments/my [get]
func (c *AppointmentController) GetMyAppointments(w http.ResponseWriter, r *http.Request) {
	claims, err := c.getClaimsFromRequest(r)
	if err != nil {
		response.Unauthorized(w, "unauthorized")
		return
	}

	appointments, err := c.usecase.GetMyAppointments(claims.UserID)
	if err != nil {
		response.Internal(w, "failed to get appointments")
		return
	}

	response.Success(w, appointments)
}

// GetAllAppointments godoc
// @Summary      Все записи (админ)
// @Description  Возвращает все записи — только для администратора
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer токен"
// @Success      200            {object}  map[string]any
// @Failure      401            {object}  map[string]string
// @Failure      403            {object}  map[string]string
// @Router       /admin/appointments [get]
func (c *AppointmentController) GetAllAppointments(w http.ResponseWriter, r *http.Request) {
	claims, err := c.getClaimsFromRequest(r)
	if err != nil {
		response.Unauthorized(w, "unauthorized")
		return
	}
	if claims.Role != "admin" {
		response.Forbidden(w, "admins only")
		return
	}

	appointments, err := c.usecase.GetAllAppointments()
	if err != nil {
		response.Internal(w, "failed to get appointments")
		return
	}

	response.Success(w, appointments)
}

// ApproveAppointment godoc
// @Summary      Принять запись (админ)
// @Description  Меняет статус записи на approved
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer токен"
// @Param        id             path      int     true  "ID записи"
// @Success      200            {object}  map[string]string
// @Failure      400            {object}  map[string]string
// @Failure      403            {object}  map[string]string
// @Router       /admin/appointments/{id}/approve [patch]
func (c *AppointmentController) ApproveAppointment(w http.ResponseWriter, r *http.Request) {
	claims, err := c.getClaimsFromRequest(r)
	if err != nil {
		response.Unauthorized(w, "unauthorized")
		return
	}
	if claims.Role != "admin" {
		response.Forbidden(w, "admins only")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	if err := c.usecase.ApproveAppointment(id); err != nil {
		c.log.Warn("approve failed", zap.Error(err))
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "appointment approved"})
}

// RejectAppointment godoc
// @Summary      Отклонить запись (админ)
// @Description  Меняет статус записи на rejected
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer токен"
// @Param        id             path      int     true  "ID записи"
// @Success      200            {object}  map[string]string
// @Failure      400            {object}  map[string]string
// @Failure      403            {object}  map[string]string
// @Router       /admin/appointments/{id}/reject [patch]
func (c *AppointmentController) RejectAppointment(w http.ResponseWriter, r *http.Request) {
	claims, err := c.getClaimsFromRequest(r)
	if err != nil {
		response.Unauthorized(w, "unauthorized")
		return
	}
	if claims.Role != "admin" {
		response.Forbidden(w, "admins only")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	if err := c.usecase.RejectAppointment(id); err != nil {
		c.log.Warn("reject failed", zap.Error(err))
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "appointment rejected"})
}

// GetPsychologists godoc
// @Summary      Список психологов
// @Description  Возвращает список всех психологов
// @Tags         psychologists
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /psychologists [get]
func (c *AppointmentController) GetPsychologists(w http.ResponseWriter, r *http.Request) {
	psychologists, err := c.usecase.GetPsychologists()
	if err != nil {
		response.Internal(w, "failed to get psychologists")
		return
	}
	response.Success(w, psychologists)
}

// GetPsychologistByID godoc
// @Summary      Информация о психологе
// @Description  Возвращает информацию о конкретном психологе
// @Tags         psychologists
// @Produce      json
// @Param        id   path      int  true  "ID психолога"
// @Success      200  {object}  map[string]any
// @Failure      404  {object}  map[string]string
// @Router       /psychologists/{id} [get]
func (c *AppointmentController) GetPsychologistByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	p, err := c.usecase.GetPsychologistByID(id)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}

	response.Success(w, p)
}
