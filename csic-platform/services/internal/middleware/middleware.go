package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/csic-platform/services/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	svc *services.Services
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(svc *services.Services) *AuthMiddleware {
	return &AuthMiddleware{svc: svc}
}

// Authenticate 认证请求
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// 验证Bearer令牌格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  "INVALID_TOKEN_FORMAT",
			})
			return
		}

		tokenString := parts[1]

		// 解析和验证JWT令牌
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.svc.Config.Security.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			return
		}

		// 提取声明
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
				"code":  "INVALID_CLAIMS",
			})
			return
		}

		// 验证必需声明
		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing user ID in token",
				"code":  "MISSING_USER_ID",
			})
			return
		}

		userRole, ok := claims["role"].(string)
		if !ok || userRole == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing user role in token",
				"code":  "MISSING_USER_ROLE",
			})
			return
		}

		// 设置上下文中的用户信息
		c.Set("user_id", userID)
		c.Set("user_role", userRole)
		c.Set("username", claims["username"])
		c.Set("department", claims["department"])

		// 继续处理请求
		c.Next()
	}
}

// RequireRole 要求特定角色
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")

		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":   "Insufficient permissions",
			"code":    "FORBIDDEN",
			"message": "You do not have the required role to access this resource",
		})
	}
}

// RequireMFA 要求MFA验证
func (m *AuthMiddleware) RequireMFA() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在实际实现中，这里会检查用户是否已通过MFA验证
		// 可以通过检查会话中的mfa_verified标记

		c.Next()
	}
}

// GenerateToken 生成JWT令牌
func (m *AuthMiddleware) GenerateToken(userID, username, role, department string) (string, error) {
	now := time.Now()
	expiry := now.Add(time.Duration(m.svc.Config.Security.TokenExpiryHours) * time.Hour)
	refreshExpiry := now.Add(time.Duration(m.svc.Config.Security.RefreshExpiryHours) * time.Hour)

	claims := jwt.MapClaims{
		"sub":       userID,
		"username":  username,
		"role":      role,
		"department": department,
		"iat":       now.Unix(),
		"exp":       expiry.Unix(),
		"refresh":   refreshExpiry.Unix(),
		"jti":       generateJTI(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(m.svc.Config.Security.JWTSecret))
}

// generateJTI 生成JWT ID
func generateJTI() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte('a' + i%26)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	wormStorage *services.WORMStorage
}

// NewLoggingMiddleware 创建日志中间件
func NewLoggingMiddleware(wormStorage *services.WORMStorage) *LoggingMiddleware {
	return &LoggingMiddleware{wormStorage: wormStorage}
}

// Logger 日志记录中间件
func (m *LoggingMiddleware) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 记录日志
		logEntry := gin.H{
			"timestamp":    time.Now().UTC(),
			"request_id":   c.GetHeader("X-Request-ID"),
			"user_id":      c.GetString("user_id"),
			"method":       c.Request.Method,
			"path":         path,
			"query":        query,
			"status":       c.Writer.Status(),
			"latency_ms":   latency.Milliseconds(),
			"client_ip":    c.ClientIP(),
			"user_agent":   c.GetHeader("User-Agent"),
		}

		// 记录到标准日志
		if c.Writer.Status() >= 400 {
			// 错误日志
			logEntry["error"] = c.Errors.ByType(gin.ErrorTypePrivate).String()
		}

		// 序列化日志
		logData, _ := json.Marshal(logEntry)
		_ = logData

		// 在实际实现中，这里会将日志发送到日志系统
	}
}

// SecurityLogging 安全事件日志中间件
func (m *LoggingMiddleware) SecurityLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录安全相关事件
		if isSecurityEvent(c) {
			// 记录到WORM存储
			m.logSecurityEvent(c)
		}
		c.Next()
	}
}

// isSecurityEvent 判断是否为安全事件
func isSecurityEvent(c *gin.Context) bool {
	// 检查认证失败
	if c.Writer.Status() == http.StatusUnauthorized {
		return true
	}

	// 检查敏感操作
	path := c.Request.URL.Path
	sensitivePaths := []string{
		"/emergency/",
		"/freeze",
		"/revoke",
		"/security/",
	}

	for _, sp := range sensitivePaths {
		if strings.HasPrefix(path, sp) {
			return true
		}
	}

	return false
}

// logSecurityEvent 记录安全事件
func (m *LoggingMiddleware) logSecurityEvent(c *gin.Context) {
	event := map[string]interface{}{
		"timestamp":   time.Now().UTC(),
		"event_type":  "SECURITY_ALERT",
		"user_id":     c.GetString("user_id"),
		"user_role":   c.GetString("user_role"),
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"status":      c.Writer.Status(),
		"client_ip":   c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
	}

	eventData, _ := json.Marshal(event)
	_ = eventData

	// 在实际实现中，这里会发送到安全信息和事件管理系统(SIEM)
}

// AuditMiddleware 审计中间件
type AuditMiddleware struct {
	wormStorage *services.WORMStorage
}

// NewAuditMiddleware 创建审计中间件
func NewAuditMiddleware(wormStorage *services.WORMStorage) *AuditMiddleware {
	return &AuditMiddleware{wormStorage: wormStorage}
}

// Audit 审计记录中间件
func (m *AuditMiddleware) Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 捕获请求体（限制大小）
		var requestBody string
		if c.Request.ContentLength > 0 && c.Request.ContentLength < 10240 {
			bodyBytes := make([]byte, c.Request.ContentLength)
			c.Request.Body.Read(bodyBytes)
			c.Request.Body = nil
			requestBody = string(bodyBytes)
		}

		// 保存原始状态
		c.Next()

		// 记录审计日志
		auditLog := map[string]interface{}{
			"user_id":       c.GetString("user_id"),
			"user_role":     c.GetString("user_role"),
			"action":        m.getAction(c.Request.Method, c.Request.URL.Path),
			"resource_type": m.getResourceType(c.Request.URL.Path),
			"resource_id":   c.Param("id"),
			"ip_address":    c.ClientIP(),
			"user_agent":    c.GetHeader("User-Agent"),
			"request_body":  requestBody,
			"response_code": c.Writer.Status(),
			"timestamp":     time.Now().UTC(),
		}

		// 记录到WORM存储
		if m.wormStorage != nil {
			// 在实际实现中，这里会将审计日志写入WORM存储
		}
	}
}

// getAction 根据请求获取操作类型
func (m *AuditMiddleware) getAction(method, path string) string {
	actionMap := map[string]string{
		http.MethodGet:    "READ",
		http.MethodPost:   "CREATE",
		http.MethodPut:    "UPDATE",
		http.MethodPatch:  "UPDATE",
		http.MethodDelete: "DELETE",
	}

	if action, ok := actionMap[method]; ok {
		return action
	}
	return "UNKNOWN"
}

// getResourceType 根据路径获取资源类型
func (m *AuditMiddleware) getResourceType(path string) string {
	// 提取资源类型
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 1 {
		return parts[1]
	}
	return "UNKNOWN"
}

// RateLimitMiddleware 速率限制中间件
type RateLimitMiddleware struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimitMiddleware 创建速率限制中间件
func NewRateLimitMiddleware(limit int, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RateLimit 速率限制中间件
func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		m.mu.Lock()

		// 清理过期请求记录
		cutoff := now.Add(-m.window)
		if times, ok := m.requests[ip]; ok {
			var validTimes []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					validTimes = append(validTimes, t)
				}
			}
			m.requests[ip] = validTimes
		}

		// 检查速率限制
		if len(m.requests[ip]) >= m.limit {
			m.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"code":    "RATE_LIMIT",
				"message": "Too many requests, please try again later",
				"retry_after": m.window.Seconds(),
			})
			return
		}

		// 记录请求
		m.requests[ip] = append(m.requests[ip], now)
		m.mu.Unlock()

		c.Next()
	}
}

// CORSMiddleware CORS中间件
type CORSMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
}

// NewCORSMiddleware 创建CORS中间件
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: []string{"https://localhost:3000"}, // 限制为内部网络
		allowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		allowedHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
	}
}

// CORS CORS中间件处理
func (m *CORSMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 验证来源
		allowed := false
		for _, o := range m.allowedOrigins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", strings.Join(m.allowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(m.allowedHeaders, ", "))
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware 安全头中间件
type SecurityHeadersMiddleware struct{}

// NewSecurityHeadersMiddleware 创建安全头中间件
func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{}
}

// SecurityHeaders 添加安全响应头
func (m *SecurityHeadersMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止XSS攻击
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")

		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// 严格传输安全
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}

// RequestIDMiddleware 请求ID中间件
type RequestIDMiddleware struct{}

// NewRequestIDMiddleware 创建请求ID中间件
func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

// RequestID 生成和传递请求ID
func (m *RequestIDMiddleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否已有请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置到上下文和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID 生成唯一请求ID
func generateRequestID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte('a' + i%26)
	}
	return base64.URLEncoding.EncodeToString(b)
}

// HMACMiddleware HMAC验证中间件
type HMACMiddleware struct {
	secret []byte
}

// NewHMACMiddleware 创建HMAC中间件
func NewHMACMiddleware(secret string) *HMACMiddleware {
	return &HMACMiddleware{
		secret: []byte(secret),
	}
}

// VerifyHMAC 验证HMAC签名
func (m *HMACMiddleware) VerifyHMAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取签名
		signature := c.GetHeader("X-Signature")
		if signature == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing signature",
				"code":  "MISSING_SIGNATURE",
			})
			return
		}

		// 获取时间戳
		timestamp := c.GetHeader("X-Timestamp")
		if timestamp == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing timestamp",
				"code":  "MISSING_TIMESTAMP",
			})
			return
		}

		// 验证时间戳（防止重放攻击）
		ts, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid timestamp format",
				"code":  "INVALID_TIMESTAMP",
			})
			return
		}

		if time.Since(ts) > 5*time.Minute {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Request timestamp too old",
				"code":  "TIMESTAMP_EXPIRED",
			})
			return
		}

		// 计算期望的HMAC
		stringToSign := fmt.Sprintf("%s:%s:%s", c.Request.Method, c.Request.URL.Path, timestamp)
		expectedMAC := computeHMAC(stringToSign, m.secret)

		// 验证签名
		if !secureCompare(signature, expectedMAC) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid signature",
				"code":  "INVALID_SIGNATURE",
			})
			return
		}

		c.Next()
	}
}

// computeHMAC 计算HMAC
func computeHMAC(message string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// secureCompare 安全比较（防止时序攻击）
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
