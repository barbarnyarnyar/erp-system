// File: services/auth-service/go.mod
module auth-service

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/go-redis/redis/v8 v8.11.5
	golang.org/x/crypto v0.10.0
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
	github.com/joho/godotenv v1.4.0
	github.com/go-playground/validator/v10 v10.15.5
)

// File: services/auth-service/.env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=erp_auth_db
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-super-secret-development-key-change-in-production
JWT_EXPIRY_HOURS=24
REFRESH_TOKEN_EXPIRY_DAYS=7

# Server Configuration
PORT=8090
GIN_MODE=debug

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080