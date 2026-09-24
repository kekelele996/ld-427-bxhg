package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/renovation/renovation-budget-api/internal/auth"
	"github.com/renovation/renovation-budget-api/internal/config"
	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/repository"
)

// AuthService 认证业务逻辑。
type AuthService struct {
	repo   repository.UserRepository
	cfg    *config.Config
	logger *slog.Logger
}

// NewAuthService 构造认证服务。
func NewAuthService(repo repository.UserRepository, cfg *config.Config, logger *slog.Logger) *AuthService {
	return &AuthService{repo: repo, cfg: cfg, logger: logger}
}

// Login 校验用户并签发 JWT。
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidLogin
		}
		return nil, fmt.Errorf("login user %s: %w", req.Username, err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("login user %s: %w", req.Username, ErrInvalidLogin)
	}
	expire := time.Duration(s.cfg.JWTExpireHours) * time.Hour
	token, err := auth.GenerateToken(s.cfg.JWTSecret, user.ID, user.Username, constants.RoleName(user.Role.Name), expire)
	if err != nil {
		return nil, fmt.Errorf("generate token for %s: %w", req.Username, err)
	}
	return &dto.LoginResponse{
		Token:     token,
		ExpiresIn: int(expire.Seconds()),
		User: dto.UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Role:        constants.RoleName(user.Role.Name),
		},
	}, nil
}
