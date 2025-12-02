package services

import (
	"errors"
	"time"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo repositories.UserRepository
	cfg  *config.Config
}

func NewUserService(repo repositories.UserRepository, cfg *config.Config) *UserService {
	return &UserService{
		repo: repo,
		cfg:  cfg,
	}
}

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register xử lý logic đăng ký user mới
func (s *UserService) Register(input RegisterInput) (*models.User, error) {
	// 1. Kiểm tra tồn tại
	exists, _ := s.repo.ExistsByEmail(input.Email)
	if exists {
		return nil, errors.New("email already exists")
	}
	exists, _ = s.repo.ExistsByUsername(input.Username)
	if exists {
		return nil, errors.New("username already exists")
	}

	// 2. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Tạo user model
	user := &models.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Role:         models.RoleUser,
		// TOTP mặc định false
	}

	// 4. Lưu vào DB
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login xử lý logic đăng nhập và trả về JWT token
func (s *UserService) Login(input LoginInput) (string, *models.User, error) {
	// 1. Tìm user theo email
	user, err := s.repo.GetByEmail(input.Email)
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	// 2. Kiểm tra password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	// 3. Tạo JWT Token
	token, err := s.generateJWT(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// generateJWT tạo token JWT
func (s *UserService) generateJWT(user *models.User) (string, error) {
	// Parse thời gian hết hạn từ config (ví dụ "15m")
	duration, err := time.ParseDuration(s.cfg.JWT.AccessTokenExpiry)
	if err != nil {
		duration = 15 * time.Minute // Fallback
	}

	claims := jwt.MapClaims{
		"sub":      user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(duration).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}
