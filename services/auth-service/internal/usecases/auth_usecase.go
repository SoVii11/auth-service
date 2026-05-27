package usecase

import (
	"crypto/rand"

	"errors"
	"fmt"
	"math/big"
	"net/smtp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/SoVii11/auth-service/services/auth-service/internal/config"
	"github.com/SoVii11/auth-service/services/auth-service/internal/domain"
	sharedJWT "github.com/SoVii11/shared/pkg/jwt"
)

type AuthUsecase struct {
	userRepo      UserRepo
	resetCodeRepo ResetCodeRepo
	cfg           *config.Config
}

func NewAuthUsecase(userRepo UserRepo, resetCodeRepo ResetCodeRepo, cfg *config.Config) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo, resetCodeRepo: resetCodeRepo, cfg: cfg}
}

func (u *AuthUsecase) Register(email, password string) (*domain.User, error) {
	_, err := u.userRepo.FindByEmail(email)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{Email: email, Password: string(hashed), Role: "user"}
	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *AuthUsecase) Login(email, password string) (string, error) {
	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", errors.New("invalid email or password")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	return sharedJWT.GenerateToken(user.ID, user.Email, user.Role, u.cfg.JWTSecret, 24*time.Hour)
}

func (u *AuthUsecase) SendResetCode(email string) error {
	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	code, err := generateCode()
	if err != nil {
		return err
	}

	_ = u.resetCodeRepo.DeleteByUserID(user.ID)

	resetCode := &domain.ResetCode{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := u.resetCodeRepo.Create(resetCode); err != nil {
		return err
	}

	return u.sendEmail(email, code)
}

func (u *AuthUsecase) ResetPassword(email, code, newPassword string) error {
	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	resetCode, err := u.resetCodeRepo.FindByUserIDAndCode(user.ID, code)
	if err != nil {
		return errors.New("invalid code")
	}

	if time.Now().After(resetCode.ExpiresAt) {
		return errors.New("code expired")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := u.userRepo.UpdatePassword(user.ID, string(hashed)); err != nil {
		return err
	}

	return u.resetCodeRepo.DeleteByUserID(user.ID)
}

func generateCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[n.Int64()]
	}
	return string(code), nil
}

func (u *AuthUsecase) sendEmail(to, code string) error {
	auth := smtp.PlainAuth("", u.cfg.SMTPUser, u.cfg.SMTPPass, u.cfg.SMTPHost)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: Код восстановления пароля\r\n\r\nВаш код: %s\r\nКод действителен 15 минут.",
		u.cfg.SMTPUser, to, code,
	)
	return smtp.SendMail(fmt.Sprintf("%s:%s", u.cfg.SMTPHost, u.cfg.SMTPPort), auth, u.cfg.SMTPUser, []string{to}, []byte(msg))
}
