// services/finance/go.mod
module github.com/sithuhlaing/erp-system/services/finance

go 1.21

require (
    github.com/sithuhlaing/erp-system/shared v0.0.0
    github.com/gin-gonic/gin v1.9.1
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
)

replace github.com/sithuhlaing/erp-system/shared => ../../shared