// services/scm/go.mod
module github.com/sithuhlaing/erp-system/services/scm

go 1.21

require (
    erp-system/shared v0.0.0
    github.com/gin-gonic/gin v1.9.1
)

replace erp-system/shared => ../../shared
