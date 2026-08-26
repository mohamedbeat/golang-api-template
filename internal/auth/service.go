package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"golang-api-template/internal/auth/dtos"
	"golang-api-template/internal/config"
	"golang-api-template/internal/middleware"
	"golang-api-template/internal/users"
	"golang-api-template/pkg/apperror"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	//user
	LoginUser(ctx context.Context, req *dtos.LoginRequest) (access string, refresh string, err error)
	RegisterUser(ctx context.Context, req *dtos.CreateUserRequest) (*users.User, error)
	RefreshSession(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error

	// Password hashing
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool

	// JWT operations
	GenerateAccessToken(userID uuid.UUID, email, name, typ string) (string, error)
	ValidateAccessToken(tokenString string) (*middleware.Claims, error)
	GenerateRefreshToken() (string, string)
}

// refreshRaceWindow bounds how long an already-rotated token is treated as
// "probably a concurrent request racing itself" rather than "reused token,
// likely stolen". Keep this short — long enough to absorb network jitter
// and a couple of parallel browser tabs, short enough that it doesn't
// meaningfully widen the window for an actual attacker.
const refreshRaceWindow = 10 * time.Second

type authService struct {
	sessionRepo   SessionRepository
	userRepo      users.UserRepository
	jwtSecret     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	refreshCache  *refreshCache
}

func NewAuthService(cfg *config.Config, sessionRepo SessionRepository, userRepo users.UserRepository) AuthService {
	return &authService{
		jwtSecret:     []byte(cfg.Auth.JwtSecret),
		accessExpiry:  time.Duration(cfg.Auth.JwtDuration) * time.Second,
		refreshExpiry: time.Duration(cfg.Auth.RefreshDuration) * time.Hour,
		sessionRepo:   sessionRepo,
		userRepo:      userRepo,
		refreshCache:  newRefreshCache(refreshRaceWindow),
	}
}

////////////////////////
//
// user related
//
////////////////////////

func (s *authService) RefreshSession(ctx context.Context, refreshToken string) (string, string, error) {
	hash := s.hashToken(refreshToken)
	cacheKey := "user:" + hash

	// Fast path: an identical request already rotated this exact token
	// within the race window — hand back the same result instead of
	// hitting the DB (and instead of tripping reuse detection below).
	if cached, ok := s.refreshCache.Get(cacheKey); ok {
		log.Println("cached refresh found")
		return cached.Access, cached.Refresh, nil
	}

	log.Println("no cached refresh found")
	var accessToken, newRaw string

	err := s.sessionRepo.WithTx(ctx, func(tx *sqlx.Tx) error { // tx begin
		session, err := s.sessionRepo.GetForUpdate(ctx, tx, hash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperror.Unauthorized("Session not found")
			}
			return err
		}

		if session.RevokedAt != nil {
			return apperror.Unauthorized("Session revoked")
		}

		if session.IsExpired() {
			return apperror.Unauthorized("Session expired")
		}

		if session.ReplacedBy != nil {
			if time.Since(session.UpdatedAt) < refreshRaceWindow {
				// Almost certainly a second concurrent request that missed
				// the in-memory cache (e.g. landed on a different process).
				// Reject rather than serve stale data; the client already
				// has the winning request's new tokens stored.
				return apperror.Unauthorized("Session already refreshed")
			}
			// An old, already-rotated token reappearing well outside the
			// race window is the signature of a stolen token being
			// reused. Kill the whole chain, not just this token.
			_ = s.sessionRepo.RevokeFamily(ctx, tx, session.FamilyID)
			return apperror.Unauthorized("refresh token reuse detected")
		}

		newTokenRaw, newHash := s.GenerateRefreshToken()

		newSession := &Session{
			UserID:           session.UserID,
			FamilyID:         session.FamilyID,
			RefreshTokenHash: newHash,
			ExpiresAt:        time.Now().UTC().Add(s.refreshExpiry),
		}
		if err := s.sessionRepo.RotateSession(ctx, tx, session.ID, newSession); err != nil {
			return err
		}

		user, err := s.userRepo.GetByID(ctx, session.UserID)
		if err != nil {
			return err
		}
		if user == nil {
			return apperror.Unauthorized("User not found")
		}

		access, err := s.GenerateAccessToken(user.ID, user.Email, user.FullName, user.Type)
		if err != nil {
			return err
		}

		accessToken = access
		newRaw = newTokenRaw
		s.refreshCache.Set(cacheKey, refreshResult{Access: access, Refresh: newTokenRaw})
		return nil
	}) // tx end

	if err != nil {
		return "", "", err
	}
	return accessToken, newRaw, nil
}

func (s *authService) LoginUser(ctx context.Context, req *dtos.LoginRequest) (string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", "", err
	}

	if user == nil {
		return "", "", apperror.Unauthorized("User not found")
	}

	isPassCorrect := s.CheckPasswordHash(req.Password, user.Password_hash)
	if !isPassCorrect {
		return "", "", apperror.Unauthorized("Invalid credentials")
	}

	rawToken, hash := s.GenerateRefreshToken()

	session := &Session{
		UserID:           user.ID,
		RefreshTokenHash: hash,
		ExpiresAt:        time.Now().UTC().Add(s.refreshExpiry),
	}
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return "", "", err
	}
	accessToken, err := s.GenerateAccessToken(user.ID, user.Email, user.FullName, user.Type)
	if err != nil {
		return "", "", err
	}
	return accessToken, rawToken, nil
}

func (s *authService) RegisterUser(ctx context.Context, req *dtos.CreateUserRequest) (*users.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return nil, apperror.Conflict("User already exist")
	}

	hashed_pass, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	req.Password = hashed_pass
	created, err := s.userRepo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	hash := s.hashToken(refreshToken)
	session, err := s.sessionRepo.FindByToken(hash)
	if err != nil {
		return err
	}
	if session == nil {
		return nil // already gone/expired — logout is idempotent, not an error
	}
	return s.sessionRepo.RevokeByID(ctx, session.ID)
}

////////////////////////
// auth related
////////////////////////

func (s *authService) hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(h[:])
}

// HashPassword hashes a plain text password
func (s *authService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash compares a password with its hash
func (s *authService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Generate access Token for users
func (s *authService) GenerateAccessToken(userID uuid.UUID, email, name, typ string) (string, error) {
	claims := middleware.Claims{
		UserID: userID,
		Email:  email,
		Type:   typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "golang-api-template",
			Subject:   name,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateAccessToken validates and parses a JWT token
func (s *authService) ValidateAccessToken(tokenString string) (*middleware.Claims, error) {
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, s.keyFunc)
	if err != nil || !token.Valid {
		return nil, apperror.Unauthorized("invalid or expired token")
	}
	return claims, nil
}

// helper for token sign validations
func (s *authService) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}
	return s.jwtSecret, nil
}

func (s *authService) GenerateRefreshToken() (refreshToken string, hashedToken string) {

	b := make([]byte, 32)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)

	hash := sha256.Sum256([]byte(token))
	return token, base64.URLEncoding.EncodeToString(hash[:])
}
