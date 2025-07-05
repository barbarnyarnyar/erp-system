// File: services/auth-service/internal/handlers/auth.go
package handlers

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"auth-service/internal/models"
	"auth-service/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
	jwtSecret   string
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         *models.User `json:"user"`
}

type JWTClaims struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func NewAuthHandler(authService *services.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user
	user, err := h.authService.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Update last login
	user.LastLogin = &time.Time{}
	*user.LastLogin = time.Now()
	h.authService.UpdateUser(user)

	// Create session
	session := &models.Session{
		UserID:    user.ID,
		Token:     accessToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
	h.authService.CreateSession(session)

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400, // 24 hours
		User:         user,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No token provided"})
		return
	}

	// Remove "Bearer " prefix
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Invalidate session
	err := h.authService.InvalidateSession(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) ValidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No token provided"})
		return
	}

	// Remove "Bearer " prefix
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Validate token
	claims, err := h.validateJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Check if session is still active
	session, err := h.authService.GetSessionByToken(token)
	if err != nil || !session.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":  true,
		"claims": claims,
	})
}

func (h *AuthHandler) GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) CheckPermission(c *gin.Context) {
	service := c.Query("service")
	resource := c.Query("resource")
	action := c.Query("action")

	if service == "" || resource == "" || action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	hasPermission, err := h.authService.CheckPermission(userID.(uint), service, resource, action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_permission": hasPermission,
		"service":        service,
		"resource":       resource,
		"action":         action,
	})
}

func (h *AuthHandler) generateTokens(user *models.User) (string, string, error) {
	// Get user roles and permissions
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	permissions := make([]string, 0)
	for _, perm := range user.Permissions {
		permissions = append(permissions, perm.Service+":"+perm.Resource+":"+perm.Action)
	}
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			permissions = append(permissions, perm.Service+":"+perm.Resource+":"+perm.Action)
		}
	}

	// Create access token claims
	accessClaims := JWTClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Create refresh token claims (longer expiry, fewer claims)
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   user.Username,
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (h *AuthHandler) validateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}