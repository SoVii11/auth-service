package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/SoVii11/auth-service/config"
	_ "github.com/SoVii11/auth-service/docs"
	"github.com/SoVii11/auth-service/internal/controllers"
	"github.com/SoVii11/auth-service/internal/infrastructure/postgres"
	"github.com/SoVii11/auth-service/internal/infrastructure/repository"
	"github.com/SoVii11/auth-service/internal/usecases"
	"github.com/SoVii11/shared/pkg/logger"
)

// @title           Auth Service API
// @version         1.0
// @description     Сервис авторизации для психологической платформы
// @host            localhost:8080
// @BasePath        /
func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if err := logger.Init(cfg.AppMode); err != nil {
		panic(err)
	}
	defer logger.Sync()

	log := logger.Get()

	db, err := postgres.NewDB(cfg)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	log.Info("connected to database")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	resetCodeRepo := repository.NewResetCodeRepository(db)
	appointmentRepo := repository.NewAppointmentRepository(db)
	psychologistRepo := repository.NewPsychologistRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)

	// Usecases
	authUsecase := usecases.NewAuthUsecase(userRepo, resetCodeRepo, cfg)
	appointmentUsecase := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)

	// Controllers
	authController := controllers.NewAuthController(authUsecase, log)
	appointmentController := controllers.NewAppointmentController(appointmentUsecase, log, cfg.JWTSecret)
	chatController := controllers.NewChatController(messageRepo, log, cfg.JWTSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Auth routes
	r.Post("/auth/register", authController.Register)
	r.Post("/auth/login", authController.Login)
	r.Post("/auth/send-reset-code", authController.SendResetCode)
	r.Post("/auth/reset-password", authController.ResetPassword)

	// Psychologist routes
	r.Get("/psychologists", appointmentController.GetPsychologists)
	r.Get("/psychologists/{id}", appointmentController.GetPsychologistByID)
	r.Get("/psychologists/{id}/vacations", appointmentController.GetVacations)
	r.Get("/psychologists/{id}/availability", appointmentController.CheckAvailability)

	// Appointment routes
	r.Post("/appointments", appointmentController.CreateAppointment)
	r.Get("/appointments/my", appointmentController.GetMyAppointments)

	// Admin routes
	r.Get("/admin/appointments", appointmentController.GetAllAppointments)
	r.Patch("/admin/appointments/{id}/approve", appointmentController.ApproveAppointment)
	r.Patch("/admin/appointments/{id}/reject", appointmentController.RejectAppointment)
	r.Post("/admin/psychologists/{id}/vacation", appointmentController.AddVacation)
	r.Delete("/admin/psychologists/vacation/{id}", appointmentController.RemoveVacation)

	// Chat WebSocket
	r.Get("/ws/chat", chatController.HandleUserChat)
	r.Get("/ws/admin/chat", chatController.HandleAdminChat)

	// Chat REST
	r.Get("/chat/history", chatController.GetMyChat)
	r.Get("/admin/chats", chatController.GetAllChats)
	r.Get("/admin/chat/{id}", chatController.GetChatHistory)
	r.Get("/admin/chat/{id}/export", chatController.ExportChatJSON)

	// Swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	fmt.Println("Server started on port 8080")
	log.Info("server started", zap.Int("port", 8080))

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}
