package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/grownmind/backend/internal/user"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
	blocklistPrefix = "blocklist:"
	oauthStatePrefix = "oauth_state:"
)

type Service struct {
	users              *user.Repository
	redis              *redis.Client
	jwtSecret          []byte
	jwtRefreshSecret   []byte
	googleClientID     string
	googleClientSecret string
	googleRedirectURI  string
}

func NewService(users *user.Repository, rdb *redis.Client, jwtSecret, jwtRefreshSecret, googleClientID, googleClientSecret, googleRedirectURI string) *Service {
	return &Service{
		users:              users,
		redis:              rdb,
		jwtSecret:          []byte(jwtSecret),
		jwtRefreshSecret:   []byte(jwtRefreshSecret),
		googleClientID:     googleClientID,
		googleClientSecret: googleClientSecret,
		googleRedirectURI:  googleRedirectURI,
	}
}

func (s *Service) Register(ctx context.Context, email, password, fullName, username string) (*user.User, string, string, error) {
	if existing, _ := s.users.GetByEmail(ctx, email); existing != nil {
		return nil, "", "", errors.New("email already registered")
	}
	if existing, _ := s.users.GetByUsername(ctx, username); existing != nil {
		return nil, "", "", errors.New("username already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("hash password: %w", err)
	}
	hashStr := string(hash)

	created, err := s.users.Create(ctx, &user.User{
		Email:        email,
		FullName:     fullName,
		Username:     username,
		PasswordHash: &hashStr,
		Provider:     "email",
	})
	if err != nil {
		return nil, "", "", err
	}
	access, refresh, err := s.generateTokens(created.ID)
	return created, access, refresh, err
}

func (s *Service) Login(ctx context.Context, emailOrUsername, password string) (*user.User, string, string, error) {
	var (
		u   *user.User
		err error
	)
	if strings.Contains(emailOrUsername, "@") {
		u, err = s.users.GetByEmail(ctx, emailOrUsername)
	} else {
		u, err = s.users.GetByUsername(ctx, emailOrUsername)
	}
	if err != nil {
		return nil, "", "", err
	}
	if u == nil || u.PasswordHash == nil {
		return nil, "", "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}
	access, refresh, err := s.generateTokens(u.ID)
	return u, access, refresh, err
}

func (s *Service) GoogleSignIn(ctx context.Context, accessToken string) (*user.User, string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, "", "", fmt.Errorf("google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", "", errors.New("invalid google access token")
	}

	var info struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, "", "", err
	}
	if info.Email == "" {
		return nil, "", "", errors.New("google token missing email")
	}

	username := strings.Split(info.Email, "@")[0]
	var avatarURL *string
	if info.Picture != "" {
		avatarURL = &info.Picture
	}

	upserted, err := s.users.UpsertByEmail(ctx, &user.User{
		Email:     info.Email,
		FullName:  info.Name,
		Username:  username,
		AvatarURL: avatarURL,
		Provider:  "google",
	})
	if err != nil {
		return nil, "", "", err
	}
	access, refresh, err := s.generateTokens(upserted.ID)
	return upserted, access, refresh, err
}

func (s *Service) GoogleOAuthURL(ctx context.Context) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)
	if err := s.redis.Set(ctx, oauthStatePrefix+state, "1", 5*time.Minute).Err(); err != nil {
		return "", err
	}
	params := url.Values{
		"client_id":     {s.googleClientID},
		"redirect_uri":  {s.googleRedirectURI},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"online"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode(), nil
}

func (s *Service) GoogleOAuthCallback(ctx context.Context, code, state string) (*user.User, string, string, error) {
	if _, err := s.redis.GetDel(ctx, oauthStatePrefix+state).Result(); err != nil {
		return nil, "", "", errors.New("invalid or expired oauth state")
	}
	form := url.Values{
		"code":          {code},
		"client_id":     {s.googleClientID},
		"client_secret": {s.googleClientSecret},
		"redirect_uri":  {s.googleRedirectURI},
		"grant_type":    {"authorization_code"},
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		return nil, "", "", fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", "", errors.New("google token exchange failed")
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, "", "", err
	}
	if tokenResp.AccessToken == "" {
		return nil, "", "", errors.New("empty access token from google")
	}
	return s.GoogleSignIn(ctx, tokenResp.AccessToken)
}

func (s *Service) AppleSignIn(ctx context.Context, identityToken, fullName string) (*user.User, string, string, error) {
	jwks, err := getAppleJWKS()
	if err != nil {
		return nil, "", "", fmt.Errorf("apple jwks: %w", err)
	}

	token, err := jwt.Parse(identityToken, jwks.Keyfunc)
	if err != nil || !token.Valid {
		return nil, "", "", fmt.Errorf("invalid apple token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, "", "", errors.New("invalid apple token claims")
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	if sub == "" {
		return nil, "", "", errors.New("apple token missing sub")
	}
	if email == "" {
		email = sub[:8] + "@privaterelay.appleid.com"
	}

	upserted, err := s.users.UpsertByEmail(ctx, &user.User{
		Email:    email,
		FullName: fullName,
		Username: "apple_" + sub[:8],
		Provider: "apple",
	})
	if err != nil {
		return nil, "", "", err
	}
	access, refresh, err := s.generateTokens(upserted.ID)
	return upserted, access, refresh, err
}

// RefreshToken validates a refresh token, checks it is not revoked, and issues new tokens.
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := s.parseRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	jti, _ := claims["jti"].(string)
	if jti != "" {
		blocked, _ := s.redis.Exists(ctx, blocklistPrefix+jti).Result()
		if blocked > 0 {
			return "", "", errors.New("token has been revoked")
		}
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", "", errors.New("invalid token subject")
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil || u == nil {
		return "", "", errors.New("user not found")
	}

	// Revoke the old refresh token before issuing new ones.
	if jti != "" {
		exp, _ := claims["exp"].(float64)
		ttl := time.Until(time.Unix(int64(exp), 0))
		if ttl > 0 {
			s.redis.Set(ctx, blocklistPrefix+jti, "1", ttl)
		}
	}

	return s.generateTokens(userID)
}

// Logout revokes the refresh token by adding its JTI to the Redis blocklist.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.parseRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return nil
	}
	exp, _ := claims["exp"].(float64)
	ttl := time.Until(time.Unix(int64(exp), 0))
	if ttl <= 0 {
		return nil
	}
	return s.redis.Set(ctx, blocklistPrefix+jti, "1", ttl).Err()
}

func (s *Service) GetMe(ctx context.Context, userID string) (*user.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *Service) parseRefreshToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtRefreshSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	if typ, _ := claims["typ"].(string); typ != "refresh" {
		return nil, errors.New("not a refresh token")
	}
	return claims, nil
}

func (s *Service) generateTokens(userID string) (accessToken, refreshToken string, err error) {
	now := time.Now()

	at, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"jti": uuid.NewString(),
		"iat": now.Unix(),
		"exp": now.Add(accessTokenTTL).Unix(),
		"typ": "access",
	}).SignedString(s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}

	rt, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"jti": uuid.NewString(),
		"iat": now.Unix(),
		"exp": now.Add(refreshTokenTTL).Unix(),
		"typ": "refresh",
	}).SignedString(s.jwtRefreshSecret)
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}

	return at, rt, nil
}

var (
	appleJWKS     *keyfunc.JWKS
	appleJWKSOnce sync.Once
	appleJWKSErr  error
)

func getAppleJWKS() (*keyfunc.JWKS, error) {
	appleJWKSOnce.Do(func() {
		appleJWKS, appleJWKSErr = keyfunc.Get("https://appleid.apple.com/auth/keys", keyfunc.Options{
			RefreshInterval: time.Hour,
		})
	})
	return appleJWKS, appleJWKSErr
}
