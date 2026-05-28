package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/SoVii11/auth-service/internal/infrastructure/postgres"
	"github.com/SoVii11/auth-service/services/appointment-service/internal/config"
	handler "github.com/SoVii11/auth-service/services/appointment-service/internal/delivery/http"
	wshandler "github.com/SoVii11/auth-service/services/appointment-service/internal/delivery/ws"
	"github.com/SoVii11/auth-service/services/appointment-service/internal/repository"
	usecases "github.com/SoVii11/auth-service/services/appointment-service/internal/usecase"
	"github.com/SoVii11/shared/pkg/logger"
)

// @title           Appointment Service API
// @version         1.0
// @description     Сервис записи к психологу
// @host            localhost:8081
// @BasePath        /
func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if err := logger.Init(cfg.AppMode); err != nil {
		panic(err)
	}
	defer logger.Sync()

	log := logger.Get()

	db, err := postgres.NewDB(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	log.Info("connected to database")

	appointmentRepo := repository.NewAppointmentRepository(db)
	psychologistRepo := repository.NewPsychologistRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)

	appointmentUsecase := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)

	appointmentHandler := handler.NewAppointmentController(appointmentUsecase, log, cfg.JWTSecret)
	chatHandler := wshandler.NewChatController(messageRepo, log, cfg.JWTSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/psychologists", appointmentHandler.GetPsychologists)
		r.Get("/psychologists/{id}", appointmentHandler.GetPsychologistByID)
		r.Get("/psychologists/{id}/vacations", appointmentHandler.GetVacations)
		r.Get("/psychologists/{id}/availability", appointmentHandler.CheckAvailability)

		r.Post("/appointments", appointmentHandler.CreateAppointment)
		r.Get("/appointments/my", appointmentHandler.GetMyAppointments)

		r.Get("/admin/appointments", appointmentHandler.GetAllAppointments)
		r.Patch("/admin/appointments/{id}/approve", appointmentHandler.ApproveAppointment)
		r.Patch("/admin/appointments/{id}/reject", appointmentHandler.RejectAppointment)
		r.Post("/admin/psychologists/{id}/vacation", appointmentHandler.AddVacation)
		r.Delete("/admin/psychologists/vacation/{id}", appointmentHandler.RemoveVacation)

		r.Get("/ws/chat", chatHandler.HandleUserChat)
		r.Get("/ws/admin/chat", chatHandler.HandleAdminChat)
		r.Get("/chat/history", chatHandler.GetMyChat)
		r.Get("/admin/chats", chatHandler.GetAllChats)
		r.Get("/admin/chat/{id}", chatHandler.GetChatHistory)
		r.Get("/admin/chat/{id}/export", chatHandler.ExportChatJSON)
	})

	r.Get("/psychologists", appointmentHandler.GetPsychologists)
	r.Get("/psychologists/{id}", appointmentHandler.GetPsychologistByID)
	r.Get("/psychologists/{id}/vacations", appointmentHandler.GetVacations)
	r.Get("/psychologists/{id}/availability", appointmentHandler.CheckAvailability)

	r.Post("/appointments", appointmentHandler.CreateAppointment)
	r.Get("/appointments/my", appointmentHandler.GetMyAppointments)

	r.Get("/admin/appointments", appointmentHandler.GetAllAppointments)
	r.Patch("/admin/appointments/{id}/approve", appointmentHandler.ApproveAppointment)
	r.Patch("/admin/appointments/{id}/reject", appointmentHandler.RejectAppointment)
	r.Post("/admin/psychologists/{id}/vacation", appointmentHandler.AddVacation)
	r.Delete("/admin/psychologists/vacation/{id}", appointmentHandler.RemoveVacation)

	r.Get("/ws/chat", chatHandler.HandleUserChat)
	r.Get("/ws/admin/chat", chatHandler.HandleAdminChat)
	r.Get("/chat/history", chatHandler.GetMyChat)
	r.Get("/admin/chats", chatHandler.GetAllChats)
	r.Get("/admin/chat/{id}", chatHandler.GetChatHistory)
	r.Get("/admin/chat/{id}/export", chatHandler.ExportChatJSON)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	fmt.Println("Appointment service started on port 8081")
	log.Info("appointment service started", zap.Int("port", 8081))

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}
