package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/SoVii11/auth-service/internal/infrastructure/postgres"
	_ "github.com/SoVii11/auth-service/services/auth-service/docs"
	"github.com/SoVii11/auth-service/services/auth-service/internal/config"
	handler "github.com/SoVii11/auth-service/services/auth-service/internal/delivery/http"
	"github.com/SoVii11/auth-service/services/auth-service/internal/repository"
	usecase "github.com/SoVii11/auth-service/services/auth-service/internal/usecases"
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

	db, err := postgres.NewDB(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	log.Info("connected to database")

	userRepo := repository.NewUserRepository(db)
	resetCodeRepo := repository.NewResetCodeRepository(db)

	authUsecase := usecase.NewAuthUsecase(userRepo, resetCodeRepo, cfg)
	authHandler := handler.NewAuthController(authUsecase, log)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/send-reset-code", authHandler.SendResetCode)
	r.Post("/auth/reset-password", authHandler.ResetPassword)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	fmt.Println("Auth service started on port 8080")
	log.Info("auth service started", zap.Int("port", 8080))

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}
