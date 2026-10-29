package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleOrganizer UserRole = "organizer"
	RoleViewer    UserRole = "viewer"
)

type OrganizerClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Role   UserRole `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	jwtSecret   []byte
	salt        string
	adminEmail  string
	adminPass   string
	viewerEmail string
	viewerPass  string
}

func NewAuthService(secret, salt, adminEmail, adminPass, viewerEmail, viewerPass string) *AuthService {
	return &AuthService{
		jwtSecret:   []byte(secret),
		salt:        salt,
		adminEmail:  adminEmail,
		adminPass:   adminPass,
		viewerEmail: viewerEmail,
		viewerPass:  viewerPass,
	}
}

// HashSensorKey creates a strong deterministic SHA-256 hash of sensor key + salt
func (s *AuthService) HashSensorKey(apiKey string) string {
	h := sha256.New()
	h.Write([]byte(apiKey + ":" + s.salt))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateOrganizerToken issues a signed JWT for organizers
func (s *AuthService) GenerateOrganizerToken(userID, email string, role UserRole) (string, time.Time, error) {
	exp := time.Now().Add(24 * time.Hour)
	claims := OrganizerClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
			Issuer:    "campus-sensor-monitor",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// ValidateToken verifies the signature and claims of an organizer JWT
func (s *AuthService) ValidateToken(tokenString string) (*OrganizerClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &OrganizerClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*OrganizerClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AuthenticateOrganizer validates login credentials and returns token
func (s *AuthService) AuthenticateOrganizer(email, password string) (string, time.Time, UserRole, error) {
	if email == s.adminEmail && password == s.adminPass {
		token, exp, err := s.GenerateOrganizerToken("org-admin-1", email, RoleAdmin)
		return token, exp, RoleAdmin, err
	}
	if email == s.viewerEmail && password == s.viewerPass {
		token, exp, err := s.GenerateOrganizerToken("org-viewer-1", email, RoleViewer)
		return token, exp, RoleViewer, err
	}
	return "", time.Time{}, "", errors.New("invalid email or password")
}

// ExtractTokenFromRequest extracts bearer token from header or cookie
func ExtractTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	if cookie, err := r.Cookie("access_token"); err == nil {
		return cookie.Value
	}
	return ""
}
