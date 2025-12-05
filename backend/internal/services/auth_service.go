package services

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"time"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists           = errors.New("user already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrTOTPNotEnabled       = errors.New("totp not enabled")
	ErrInvalidTOTPCode      = errors.New("invalid totp code")
	ErrTOTPSecretNotCreated = errors.New("totp secret not created")
	ErrLoginSessionExpired  = errors.New("login session expired")
)

type TokenClaims struct {
	UserID      string        `json:"sub"`
	Email       string        `json:"email"`
	Username    string        `json:"username"`
	Role        models.UserRole `json:"role"`
	TOTPEnabled bool          `json:"totpEnabled"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo             repositories.UserRepository
	loginSessionRepo     repositories.LoginSessionRepository
	cfg                  *config.Config
	loginSessionTTL      time.Duration
	maxTOTPFailedAttempt int
}

type TOTPSetup struct {
	Secret string
	QRCode string
}

func NewAuthService(repo repositories.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:             repo,
		loginSessionRepo:     nil,
		cfg:                  cfg,
		loginSessionTTL:      5 * time.Minute,
		maxTOTPFailedAttempt: 5,
	}
}

// NewAuthServiceWithLoginSessions allows injecting a LoginSessionRepository for TOTP login flow.
func NewAuthServiceWithLoginSessions(
	userRepo repositories.UserRepository,
	loginSessionRepo repositories.LoginSessionRepository,
	cfg *config.Config,
) *AuthService {
	svc := NewAuthService(userRepo, cfg)
	svc.loginSessionRepo = loginSessionRepo
	return svc
}

func (s *AuthService) Register(username, email, password string) (*models.User, error) {
	if len(password) < 8 {
		return nil, fmt.Errorf("password too short")
	}

	usernameExists, err := s.userRepo.ExistsByUsername(username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, ErrUserExists
	}

	emailExists, err := s.userRepo.ExistsByEmail(email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RoleUser,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(email, password string) (*models.User, bool, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, false, err
	}
	if user == nil {
		return nil, false, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, false, ErrInvalidCredentials
	}

	totpEnabled := user.TOTPEnabled != nil && *user.TOTPEnabled
	return user, totpEnabled, nil
}


func (s *AuthService) CreateLoginSession(userID uuid.UUID) (uuid.UUID, error) {
	if s.loginSessionRepo == nil {
		return uuid.Nil, fmt.Errorf("login session repository not configured")
	}
	session, err := s.loginSessionRepo.Create(userID, s.loginSessionTTL)
	if err != nil {
		return uuid.Nil, err
	}
	return session.ID, nil
}

func (s *AuthService) LoginWithTOTP(userID uuid.UUID, code string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if user.TOTPEnabled == nil || !*user.TOTPEnabled {
		return nil, ErrTOTPNotEnabled
	}
	if user.TOTPSecret == nil {
		return nil, ErrTOTPSecretNotCreated
	}

	if !s.validateTOTP(*user.TOTPSecret, code) {
		return nil, ErrInvalidTOTPCode
	}

	return user, nil
}

func (s *AuthService) LoginWithTOTPSession(cid uuid.UUID, code string) (*models.User, error) {
	if s.loginSessionRepo == nil {
		return nil, fmt.Errorf("login session repository not configured")
	}

	now := time.Now().UTC()
	session, err := s.loginSessionRepo.GetActiveByID(cid, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLoginSessionExpired
		}
		return nil, err
	}

	user, err := s.userRepo.GetByID(session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if user.TOTPEnabled == nil || !*user.TOTPEnabled {
		return nil, ErrTOTPNotEnabled
	}
	if user.TOTPSecret == nil {
		return nil, ErrTOTPSecretNotCreated
	}

	// Validate TOTP
	if !s.validateTOTP(*user.TOTPSecret, code) {
		_ = s.loginSessionRepo.IncrementFailedAttempts(session.ID)
		return nil, ErrInvalidTOTPCode
	}

	if err := s.loginSessionRepo.MarkConsumed(session.ID, now); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) SetupTOTP(userID uuid.UUID) (*TOTPSetup, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	opts := totp.GenerateOpts{
		Issuer:      s.cfg.TOTP.Issuer,
		AccountName: user.Email,
		Period:      s.cfg.TOTP.Period,
		Digits:      totpDigits(s.cfg.TOTP.Digits),
	}

	key, err := totp.Generate(opts)
	if err != nil {
		return nil, err
	}

	img, err := key.Image(200, 200)
	if err != nil {
		return nil, err
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return nil, err
	}
	qrBase64 := base64.StdEncoding.EncodeToString(pngBuf.Bytes())

	secret := key.Secret()
	user.TOTPSecret = &secret
	enabled := false
	user.TOTPEnabled = &enabled

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return &TOTPSetup{
		Secret: secret,
		QRCode: "data:image/png;base64," + qrBase64,
	}, nil
}

func (s *AuthService) VerifyTOTP(userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if user.TOTPSecret == nil {
		return ErrTOTPSecretNotCreated
	}

	if !s.validateTOTP(*user.TOTPSecret, code) {
		return ErrInvalidTOTPCode
	}

	enabled := true
	user.TOTPEnabled = &enabled
	return s.userRepo.Update(user)
}

func (s *AuthService) GenerateAccessToken(user *models.User) (string, error) {
	accessTTL, err := s.cfg.JWT.GetAccessTokenExpiry()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	claims := TokenClaims{
		UserID:      user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		Role:        user.Role,
		TOTPEnabled: user.TOTPEnabled != nil && *user.TOTPEnabled,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *AuthService) GetProfile(userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *AuthService) DisableTOTP(userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if user.TOTPEnabled == nil || !*user.TOTPEnabled {
		return ErrTOTPNotEnabled
	}
	if user.TOTPSecret == nil {
		return ErrTOTPSecretNotCreated
	}

	if !s.validateTOTP(*user.TOTPSecret, code) {
		return ErrInvalidTOTPCode
	}

	enabled := false
	user.TOTPEnabled = &enabled
	return s.userRepo.Update(user)
}

func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, totpCode, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("password too short, minimum 8 characters required")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	totpEnabled := user.TOTPEnabled != nil && *user.TOTPEnabled

	if totpEnabled && totpCode != "" {
		if user.TOTPSecret == nil {
			return ErrTOTPSecretNotCreated
		}
		if !s.validateTOTP(*user.TOTPSecret, totpCode) {
			return ErrInvalidTOTPCode
		}
	} else {
		if oldPassword == "" {
			return fmt.Errorf("old password is required")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
			return ErrInvalidCredentials
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return s.userRepo.Update(user)
}

func (s *AuthService) validateTOTP(secret, code string) bool {
	valid, err := totp.ValidateCustom(
		code,
		secret,
		time.Now(),
		totp.ValidateOpts{
			Period:    s.cfg.TOTP.Period,
			Skew:      1,
			Digits:    totpDigits(s.cfg.TOTP.Digits),
			Algorithm: otp.AlgorithmSHA1,
		},
	)
	if err != nil {
		return false
	}
	return valid
}

func totpDigits(d uint) otp.Digits {
	switch d {
	case 6:
		return otp.DigitsSix
	case 8:
		return otp.DigitsEight
	default:
		return otp.DigitsSix
	}
}
