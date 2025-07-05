// shared/go.mod
module github.com/your-org/erp-microservices/shared

go 1.21

require (
  github.com/gin-gonic/gin v1.9.1
  github.com/golang-jwt/jwt/v5 v5.0.0
  gorm.io/gorm v1.25.5
  gorm.io/driver/postgres v1.5.4
  github.com/go-redis/redis/v8 v8.11.5
  github.com/streadway/amqp v1.1.0
  github.com/sirupsen/logrus v1.9.3
  github.com/spf13/viper v1.17.0
  github.com/google/uuid v1.4.0
)