package routes

import (
	"github.com/erp-system/hr-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, hrHandler *handlers.HrHandler) {
	v1 := r.Group("/api/v1")
	{
		// Department Management
		v1.POST("/departments", hrHandler.CreateDepartment)
		v1.GET("/departments", hrHandler.GetDepartments)
		v1.GET("/departments/:id", hrHandler.GetDepartment)
		v1.PUT("/departments/:id", hrHandler.UpdateDepartment)

		// Employee Management
		v1.POST("/employees", hrHandler.HireEmployee)
		v1.GET("/employees", hrHandler.GetEmployees)
		v1.GET("/employees/:id", hrHandler.GetEmployee)
		v1.PUT("/employees/:id", hrHandler.UpdateEmployee)
		v1.DELETE("/employees/:id", hrHandler.TerminateEmployee)
		v1.PUT("/employees/:id/compensation", hrHandler.UpdateCompensation)
		v1.GET("/employees/:id/management-chain", hrHandler.GetManagementChain)

		// Payroll Run Management
		v1.POST("/payroll/initiate", hrHandler.InitiatePayrollRun)
		v1.POST("/payroll/calculate/:id", hrHandler.ExecutePayrollCalculations)
		v1.POST("/payroll/approve/:id", hrHandler.CloseAndApprovePayroll)
		v1.GET("/payroll/runs", hrHandler.GetPayrollRuns)
		v1.GET("/payroll/runs/:id", hrHandler.GetPayrollRun)

		// Expense Claim Management
		v1.POST("/expenses", hrHandler.SubmitExpenseClaim)
		v1.GET("/expenses", hrHandler.GetExpenseClaims)
		v1.GET("/expenses/:id", hrHandler.GetExpenseClaim)
		v1.GET("/expenses/:id/lines", hrHandler.GetExpenseClaimLines)
		v1.POST("/expenses/:id/approve", hrHandler.VerifyAndApproveClaim)
		v1.POST("/expenses/:id/pay", hrHandler.PayClaim)
	}
}
