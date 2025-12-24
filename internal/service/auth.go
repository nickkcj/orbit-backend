package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/nickkcj/orbit-backend/internal/database"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type AuthService struct {
	db                 *database.Queries
	jwtSecret          []byte
	googleClientID     string
	googleClientSecret string
	googleRedirectURL  string
	frontendURL        string
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	FrontendURL  string
}

func NewAuthService(db *database.Queries, jwtSecret string, googleConfig *GoogleOAuthConfig) *AuthService {
	svc := &AuthService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
	if googleConfig != nil {
		svc.googleClientID = googleConfig.ClientID
		svc.googleClientSecret = googleConfig.ClientSecret
		svc.googleRedirectURL = googleConfig.RedirectURL
		svc.frontendURL = googleConfig.FrontendURL
	}
	return svc
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type AuthResponse struct {
	Token string        `json:"token"`
	User  database.User `json:"user"`
}

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	// Check if user exists
	_, err := s.db.GetUserByEmail(ctx, input.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := s.db.CreateUser(ctx, database.CreateUserParams{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Name:         input.Name,
		AvatarUrl:    sql.NullString{},
	})
	if err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	// Get user
	user, err := s.db.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check user status
	if user.Status != "active" {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) generateToken(user database.User) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
	return s.db.GetUserByID(ctx, id)
}

// ============================================================================
// Google OAuth
// ============================================================================

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
	googleUserURL  = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// GoogleUserInfo represents the user info from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// GenerateOAuthState generates a random state for CSRF protection
func (s *AuthService) GenerateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetGoogleAuthURL returns the Google OAuth authorization URL
func (s *AuthService) GetGoogleAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", s.googleClientID)
	params.Add("redirect_uri", s.googleRedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")

	return googleAuthURL + "?" + params.Encode()
}

// GetFrontendURL returns the frontend URL for redirects
func (s *AuthService) GetFrontendURL() string {
	return s.frontendURL
}

// ExchangeGoogleCode exchanges the authorization code for tokens and user info
func (s *AuthService) ExchangeGoogleCode(ctx context.Context, code string) (*GoogleUserInfo, error) {
	// Exchange code for tokens
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.googleClientID)
	data.Set("client_secret", s.googleClientSecret)
	data.Set("redirect_uri", s.googleRedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", googleTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Get user info
	userReq, err := http.NewRequestWithContext(ctx, "GET", googleUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := client.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userResp.Body)
		return nil, fmt.Errorf("userinfo request failed: %s", string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// LoginOrRegisterWithGoogle creates or finds a user from Google OAuth and returns auth response
func (s *AuthService) LoginOrRegisterWithGoogle(ctx context.Context, googleUser *GoogleUserInfo) (*AuthResponse, error) {
	// Try to find existing user by email
	user, err := s.db.GetUserByEmail(ctx, googleUser.Email)
	if err == nil {
		// User exists - generate token and return
		token, err := s.generateToken(user)
		if err != nil {
			return nil, err
		}
		return &AuthResponse{Token: token, User: user}, nil
	}

	// User doesn't exist - create new user (no password for OAuth users)
	user, err = s.db.CreateUser(ctx, database.CreateUserParams{
		Email:        googleUser.Email,
		PasswordHash: "", // No password for OAuth users
		Name:         googleUser.Name,
		AvatarUrl:    sql.NullString{String: googleUser.Picture, Valid: googleUser.Picture != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

// IsGoogleOAuthConfigured returns true if Google OAuth is configured
func (s *AuthService) IsGoogleOAuthConfigured() bool {
	return s.googleClientID != "" && s.googleClientSecret != ""
}
