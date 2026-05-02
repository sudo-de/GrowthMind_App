package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
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
	accessTokenTTL   = 15 * time.Minute
	refreshTokenTTL  = 30 * 24 * time.Hour
	blocklistPrefix  = "blocklist:"
	oauthStatePrefix = "oauth_state:"
	otpKeyPrefix     = "reg_pending:"
	otpTTL           = 10 * time.Minute
	maxOTPAttempts   = 5
)

type Service struct {
	users              *user.Repository
	redis              *redis.Client
	jwtSecret          []byte
	jwtRefreshSecret   []byte
	googleClientID     string
	googleClientSecret string
	googleRedirectURI  string
	smtpHost           string
	smtpPort           string
	smtpUsername       string
	smtpPassword       string
	smtpFrom           string
}

func NewService(
	users *user.Repository,
	rdb *redis.Client,
	jwtSecret, jwtRefreshSecret,
	googleClientID, googleClientSecret, googleRedirectURI,
	smtpHost, smtpPort, smtpUsername, smtpPassword, smtpFrom string,
) *Service {
	return &Service{
		users:              users,
		redis:              rdb,
		jwtSecret:          []byte(jwtSecret),
		jwtRefreshSecret:   []byte(jwtRefreshSecret),
		googleClientID:     googleClientID,
		googleClientSecret: googleClientSecret,
		googleRedirectURI:  googleRedirectURI,
		smtpHost:           smtpHost,
		smtpPort:           smtpPort,
		smtpUsername:       smtpUsername,
		smtpPassword:       smtpPassword,
		smtpFrom:           smtpFrom,
	}
}

// ── OTP registration flow ─────────────────────────────────────────────────────

type pendingRegistration struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Username     string `json:"username"`
	OTP          string `json:"otp"`
	Attempts     int    `json:"attempts"`
}

// InitiateRegistration stores a pending registration in Redis, generates a
// 6-digit OTP, and sends it to the user's email. Returns a session token that
// must be passed to VerifyRegistrationOTP.
func (s *Service) InitiateRegistration(ctx context.Context, email, password, username string) (string, error) {
	if existing, _ := s.users.GetByEmail(ctx, email); existing != nil {
		return "", errors.New("email already registered")
	}
	if existing, _ := s.users.GetByUsername(ctx, username); existing != nil {
		return "", errors.New("username already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	otp, err := generateOTP()
	if err != nil {
		return "", err
	}

	sessionToken, err := generateSecureToken()
	if err != nil {
		return "", err
	}

	pending := pendingRegistration{
		Email:        email,
		PasswordHash: string(hash),
		Username:     username,
		OTP:          otp,
		Attempts:     0,
	}
	data, _ := json.Marshal(pending)
	if err := s.redis.Set(ctx, otpKeyPrefix+sessionToken, data, otpTTL).Err(); err != nil {
		return "", fmt.Errorf("store pending registration: %w", err)
	}

	if err := s.sendOTPEmail(email, otp); err != nil {
		log.Printf("WARN: failed to send OTP email to %s: %v", email, err)
	}

	return sessionToken, nil
}

// VerifyRegistrationOTP checks the OTP, creates the user account, and returns JWT tokens.
func (s *Service) VerifyRegistrationOTP(ctx context.Context, sessionToken, otp string) (*user.User, string, string, error) {
	key := otpKeyPrefix + sessionToken
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, "", "", errors.New("code expired or invalid — please start over")
	}

	var pending pendingRegistration
	if err := json.Unmarshal(data, &pending); err != nil {
		return nil, "", "", errors.New("invalid session")
	}

	pending.Attempts++
	if pending.Attempts > maxOTPAttempts {
		s.redis.Del(ctx, key)
		return nil, "", "", errors.New("too many incorrect attempts — please register again")
	}

	if pending.OTP != otp {
		updated, _ := json.Marshal(pending)
		ttl, _ := s.redis.TTL(ctx, key).Result()
		if ttl <= 0 {
			ttl = otpTTL
		}
		s.redis.Set(ctx, key, updated, ttl)
		remaining := maxOTPAttempts - pending.Attempts
		return nil, "", "", fmt.Errorf("incorrect code — %d attempt(s) remaining", remaining)
	}

	s.redis.Del(ctx, key)

	created, err := s.users.Create(ctx, &user.User{
		Email:        pending.Email,
		FullName:     pending.Username,
		Username:     pending.Username,
		PasswordHash: &pending.PasswordHash,
		Provider:     "email",
	})
	if err != nil {
		return nil, "", "", err
	}

	access, refresh, err := s.generateTokens(created.ID)
	return created, access, refresh, err
}

func (s *Service) sendOTPEmail(to, otp string) error {
	if s.smtpHost == "" {
		log.Printf("DEV — OTP for %s: %s", to, otp)
		return nil
	}

	from := s.smtpFrom
	if from == "" {
		from = s.smtpUsername
	}

	body := fmt.Sprintf(
		"From: GrowthMind <%s>\r\nTo: %s\r\nSubject: Your verification code\r\n\r\n"+
			"Your GrowthMind verification code is: %s\r\n\r\nThis code expires in 10 minutes.\r\n",
		from, to, otp,
	)

	addr := s.smtpHost + ":" + s.smtpPort
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(body))
}

func generateOTP() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	n := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", n%1000000), nil
}

func generateSecureToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ── Standard register (kept for backwards compatibility) ──────────────────────

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

func (s *Service) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	return u != nil, nil
}

func (s *Service) IsUsernameTaken(ctx context.Context, username string) (bool, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return false, err
	}
	return u != nil, nil
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
	if s.googleClientID == "" || s.googleRedirectURI == "" {
		return "", errors.New("google oauth not configured: missing GOOGLE_CLIENT_ID or GOOGLE_REDIRECT_URI")
	}
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

	if jti != "" {
		exp, _ := claims["exp"].(float64)
		ttl := time.Until(time.Unix(int64(exp), 0))
		if ttl > 0 {
			s.redis.Set(ctx, blocklistPrefix+jti, "1", ttl)
		}
	}

	return s.generateTokens(userID)
}

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
