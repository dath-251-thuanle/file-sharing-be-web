package services_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)


type mockUserRepo struct {
	getByEmailFunc      func(email string) (*models.User, error)
	getByIDFunc         func(id uuid.UUID) (*models.User, error)
	getByUsernameFunc   func(username string) (*models.User, error)
	createFunc          func(user *models.User) error
	updateFunc          func(user *models.User) error
	deleteFunc          func(id uuid.UUID) error
	getAllFunc          func(limit, offset int) ([]models.User, int64, error)
	existsByUsernameFunc func(username string) (bool, error)
	existsByEmailFunc    func(email string) (bool, error)
}

func (m *mockUserRepo) GetByID(id uuid.UUID) (*models.User, error) {
	if m.getByIDFunc == nil {
		return nil, errors.New("not implemented")
	}
	return m.getByIDFunc(id)
}

func (m *mockUserRepo) GetByUsername(username string) (*models.User, error) {
	if m.getByUsernameFunc == nil {
		return nil, errors.New("not implemented")
	}
	return m.getByUsernameFunc(username)
}

func (m *mockUserRepo) GetByEmail(email string) (*models.User, error) {
	if m.getByEmailFunc == nil {
		return nil, errors.New("not implemented")
	}
	return m.getByEmailFunc(email)
}

func (m *mockUserRepo) Create(user *models.User) error {
	if m.createFunc == nil {
		return errors.New("not implemented")
	}
	return m.createFunc(user)
}

func (m *mockUserRepo) Update(user *models.User) error {
	if m.updateFunc == nil {
		return errors.New("not implemented")
	}
	return m.updateFunc(user)
}

func (m *mockUserRepo) Delete(id uuid.UUID) error {
	if m.deleteFunc == nil {
		return errors.New("not implemented")
	}
	return m.deleteFunc(id)
}

func (m *mockUserRepo) GetAll(limit, offset int) ([]models.User, int64, error) {
	if m.getAllFunc == nil {
		return nil, 0, errors.New("not implemented")
	}
	return m.getAllFunc(limit, offset)
}

func (m *mockUserRepo) ExistsByUsername(username string) (bool, error) {
	if m.existsByUsernameFunc == nil {
		return false, errors.New("not implemented")
	}
	return m.existsByUsernameFunc(username)
}

func (m *mockUserRepo) ExistsByEmail(email string) (bool, error) {
	if m.existsByEmailFunc == nil {
		return false, errors.New("not implemented")
	}
	return m.existsByEmailFunc(email)
}

func newAuthTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-minimum-32-characters-long",
		},
		TOTP: config.TOTPConfig{
			Issuer: "FileSharingTest",
			Period: 30,
			Digits: 6,
		},
	}
}

func newTestAuthServiceWithUserRepo(repo repositories.UserRepository) *services.AuthService {
	cfg := newAuthTestConfig()
	return services.NewAuthService(repo, cfg)
}

// ==== Tests cho Login() ====

func TestAuthService_Login_Success_TOTPDisabled(t *testing.T) {
	plainPassword := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate password hash: %v", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Username:     "testuser",
		PasswordHash: string(hash),
	}

	mockRepo := &mockUserRepo{
		getByEmailFunc: func(email string) (*models.User, error) {
			if email != user.Email {
				t.Fatalf("expected email %s, got %s", user.Email, email)
			}
			return user, nil
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	gotUser, totpEnabled, err := authService.Login(user.Email, plainPassword)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotUser == nil {
		t.Fatalf("expected user, got nil")
	}
	if gotUser.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, gotUser.ID)
	}
	if totpEnabled {
		t.Errorf("expected totpEnabled = false, got true")
	}
}

func TestAuthService_Login_Success_TOTPEnabled(t *testing.T) {
	plainPassword := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate password hash: %v", err)
	}

	enabled := true
	user := &models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Username:     "testuser",
		PasswordHash: string(hash),
		TOTPEnabled:  &enabled,
	}

	mockRepo := &mockUserRepo{
		getByEmailFunc: func(email string) (*models.User, error) {
			return user, nil
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	gotUser, totpEnabled, err := authService.Login(user.Email, plainPassword)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotUser == nil {
		t.Fatalf("expected user, got nil")
	}
	if !totpEnabled {
		t.Errorf("expected totpEnabled = true, got false")
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	plainPassword := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate password hash: %v", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Username:     "testuser",
		PasswordHash: string(hash),
	}

	mockRepo := &mockUserRepo{
		getByEmailFunc: func(email string) (*models.User, error) {
			return user, nil
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	gotUser, totpEnabled, err := authService.Login(user.Email, "wrong-password")

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
	if gotUser != nil {
		t.Errorf("expected user nil, got %#v", gotUser)
	}
	if totpEnabled {
		t.Errorf("expected totpEnabled = false, got true")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo := &mockUserRepo{
		getByEmailFunc: func(email string) (*models.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	// Act
	gotUser, totpEnabled, err := authService.Login("notfound@example.com", "any-password")

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
	if gotUser != nil {
		t.Errorf("expected user nil, got %#v", gotUser)
	}
	if totpEnabled {
		t.Errorf("expected totpEnabled = false, got true")
	}
}

func TestAuthService_Login_RepositoryError(t *testing.T) {
	// Arrange: repo trả về lỗi bất kỳ (ví dụ lỗi DB)
	expectedErr := errors.New("db error")
	mockRepo := &mockUserRepo{
		getByEmailFunc: func(email string) (*models.User, error) {
			return nil, expectedErr
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	gotUser, totpEnabled, err := authService.Login("user@example.com", "password123")

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if gotUser != nil {
		t.Errorf("expected user nil, got %#v", gotUser)
	}
	if totpEnabled {
		t.Errorf("expected totpEnabled = false, got true")
	}
}

// ==================== COMPLEX TEST CASES ====================

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := &mockUserRepo{
		existsByUsernameFunc: func(username string) (bool, error) {
			return false, nil
		},
		existsByEmailFunc: func(email string) (bool, error) {
			return false, nil
		},
		createFunc: func(user *models.User) error {
			user.ID = uuid.New()
			return nil
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	user, err := authService.Register("newuser", "newuser@example.com", "password123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil {
		t.Fatalf("expected user, got nil")
	}
	if user.Username != "newuser" {
		t.Errorf("expected username=newuser, got %s", user.Username)
	}
	if user.Email != "newuser@example.com" {
		t.Errorf("expected email=newuser@example.com, got %s", user.Email)
	}
	if user.PasswordHash == "" {
		t.Errorf("expected password hash to be set")
	}
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	mockRepo := &mockUserRepo{
		existsByUsernameFunc: func(username string) (bool, error) {
			return true, nil // Username already exists
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	user, err := authService.Register("existinguser", "new@example.com", "password123")

	if err == nil {
		t.Fatalf("expected error, got nil (user=%+v)", user)
	}
	if !errors.Is(err, services.ErrUserExists) {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	mockRepo := &mockUserRepo{
		existsByUsernameFunc: func(username string) (bool, error) {
			return false, nil
		},
		existsByEmailFunc: func(email string) (bool, error) {
			return true, nil // Email already exists
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	user, err := authService.Register("newuser", "existing@example.com", "password123")

	if err == nil {
		t.Fatalf("expected error, got nil (user=%+v)", user)
	}
	if !errors.Is(err, services.ErrUserExists) {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestAuthService_Register_PasswordTooShort(t *testing.T) {
	mockRepo := &mockUserRepo{}
	authService := newTestAuthServiceWithUserRepo(mockRepo)

	user, err := authService.Register("newuser", "new@example.com", "short")

	if err == nil {
		t.Fatalf("expected error, got nil (user=%+v)", user)
	}
	if !strings.Contains(err.Error(), "password too short") {
		t.Errorf("expected password too short error, got %v", err)
	}
}

func TestAuthService_SetupTOTP_Success(t *testing.T) {
	db := newTestDB(t)
	user := &models.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Username: "user",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	mockRepo := &mockUserRepo{
		getByIDFunc: func(id uuid.UUID) (*models.User, error) {
			return user, nil
		},
		updateFunc: func(u *models.User) error {
			*user = *u
			return nil
		},
	}

	cfg := newAuthTestConfig()
	authService := services.NewAuthService(mockRepo, cfg)

	totpSetup, err := authService.SetupTOTP(user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if totpSetup.Secret == "" {
		t.Errorf("expected TOTP secret to be generated")
	}
	if totpSetup.QRCode == "" {
		t.Errorf("expected QR code to be generated")
	}

	// Verify user has TOTP secret set
	if user.TOTPSecret == nil || *user.TOTPSecret == "" {
		t.Errorf("expected user TOTP secret to be set")
	}
}

func TestAuthService_VerifyTOTP_Success(t *testing.T) {
	db := newTestDB(t)
	secret := "JBSWY3DPEHPK3PXP"
	secretPtr := &secret
	user := &models.User{
		ID:          uuid.New(),
		Email:       "user@example.com",
		Username:    "user",
		TOTPSecret:  secretPtr,
		TOTPEnabled: &[]bool{false}[0],
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	mockRepo := &mockUserRepo{
		getByIDFunc: func(id uuid.UUID) (*models.User, error) {
			return user, nil
		},
		updateFunc: func(u *models.User) error {
			*user = *u
			return nil
		},
	}

	cfg := newAuthTestConfig()
	authService := services.NewAuthService(mockRepo, cfg)

	// Generate valid TOTP code
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	err = authService.VerifyTOTP(user.ID, code)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify TOTP is now enabled
	if user.TOTPEnabled == nil || !*user.TOTPEnabled {
		t.Errorf("expected TOTP to be enabled after verification")
	}
}

func TestAuthService_VerifyTOTP_InvalidCode(t *testing.T) {
	db := newTestDB(t)
	secret := "JBSWY3DPEHPK3PXP"
	secretPtr := &secret
	user := &models.User{
		ID:          uuid.New(),
		Email:       "user@example.com",
		Username:    "user",
		TOTPSecret:  secretPtr,
		TOTPEnabled: &[]bool{false}[0],
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	mockRepo := &mockUserRepo{
		getByIDFunc: func(id uuid.UUID) (*models.User, error) {
			return user, nil
		},
	}

	cfg := newAuthTestConfig()
	authService := services.NewAuthService(mockRepo, cfg)

	err := authService.VerifyTOTP(user.ID, "000000")
	if err == nil {
		t.Fatalf("expected error for invalid code, got nil")
	}
	if !errors.Is(err, services.ErrInvalidTOTPCode) {
		t.Errorf("expected ErrInvalidTOTPCode, got %v", err)
	}
}

func TestAuthService_GenerateAccessToken_Success(t *testing.T) {
	user := &models.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Username: "user",
		Role:     models.RoleUser,
	}

	cfg := newAuthTestConfig()
	authService := services.NewAuthService(nil, cfg)

	token, err := authService.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Errorf("expected token to be generated")
	}

	// Verify token is valid JWT format (has 3 parts separated by dots)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected JWT token format (3 parts), got %d parts", len(parts))
	}
}

func TestAuthService_GetProfile_Success(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Email:    "user@example.com",
		Username: "user",
		Role:     models.RoleUser,
	}

	mockRepo := &mockUserRepo{
		getByIDFunc: func(id uuid.UUID) (*models.User, error) {
			if id == userID {
				return user, nil
			}
			return nil, gorm.ErrRecordNotFound
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	profile, err := authService.GetProfile(userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if profile.ID != userID {
		t.Errorf("expected user ID %s, got %s", userID, profile.ID)
	}
	if profile.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, profile.Email)
	}
}

func TestAuthService_GetProfile_NotFound(t *testing.T) {
	nonExistentID := uuid.New()
	mockRepo := &mockUserRepo{
		getByIDFunc: func(id uuid.UUID) (*models.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	authService := newTestAuthServiceWithUserRepo(mockRepo)

	profile, err := authService.GetProfile(nonExistentID)
	if err == nil {
		t.Fatalf("expected error, got nil (profile=%+v)", profile)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}