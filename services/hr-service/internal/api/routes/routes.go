package routes

import (
	"github.com/erp-system/hr-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	empHandler *handlers.EmployeeHandler,
	payrollHandler *handlers.PayrollHandler,
	timesheetHandler *handlers.TimesheetHandler,
	leaveHandler *handlers.LeaveHandler,
	recruitmentHandler *handlers.RecruitmentHandler,
	performanceHandler *handlers.PerformanceHandler,
	trainingHandler *handlers.TrainingHandler,
	docHandler *handlers.DocumentHandler,
	reportHandler *handlers.ReportHandler,
) {
	v1 := r.Group("/api/v1")
	{
		// Employee Management
		v1.GET("/employees", empHandler.GetEmployees)
		v1.POST("/employees", empHandler.CreateEmployee)
		v1.GET("/employees/:id", empHandler.GetEmployee)
		v1.PUT("/employees/:id", empHandler.UpdateEmployee)
		v1.DELETE("/employees/:id", empHandler.DeleteEmployee)
		v1.POST("/employees/:id/expenses", empHandler.SubmitExpenseClaim)
		v1.GET("/departments", empHandler.GetDepartments)
		v1.POST("/departments", empHandler.CreateDepartment)
		v1.GET("/positions", empHandler.GetPositions)
		v1.POST("/positions", empHandler.CreatePosition)

		// Payroll
		v1.GET("/payroll", payrollHandler.GetPayrollRecords)
		v1.POST("/payroll", payrollHandler.ProcessPayroll)
		v1.GET("/payroll/:id", payrollHandler.GetPayrollRecord)
		v1.PUT("/payroll/:id", payrollHandler.UpdatePayrollRecord)
		v1.GET("/payroll/employee/:id", payrollHandler.GetEmployeePayroll)

		// Time & Attendance
		v1.GET("/timesheet", timesheetHandler.GetTimesheets)
		v1.POST("/timesheet", timesheetHandler.CreateTimesheet)
		v1.GET("/timesheet/:id", timesheetHandler.GetTimesheet)
		v1.PUT("/timesheet/:id", timesheetHandler.UpdateTimesheet)
		v1.POST("/timesheet/:id/submit", timesheetHandler.SubmitTimesheet)
		v1.POST("/timesheet/:id/approve", timesheetHandler.ApproveTimesheet)

		// Leave Management
		v1.GET("/leave-requests", leaveHandler.GetLeaveRequests)
		v1.POST("/leave-requests", leaveHandler.CreateLeaveRequest)
		v1.GET("/leave-requests/:id", leaveHandler.GetLeaveRequest)
		v1.PUT("/leave-requests/:id", leaveHandler.UpdateLeaveRequest)
		v1.POST("/leave-requests/:id/approve", leaveHandler.ApproveLeaveRequest)
		v1.POST("/leave-requests/:id/reject", leaveHandler.RejectLeaveRequest)
		v1.GET("/leave-balances", leaveHandler.GetLeaveBalances)


		// Recruitment
		v1.GET("/recruitment/jobs", recruitmentHandler.GetJobPostings)
		v1.POST("/recruitment/jobs", recruitmentHandler.CreateJobPosting)
		v1.GET("/recruitment/jobs/:id", recruitmentHandler.GetJobPosting)
		v1.PUT("/recruitment/jobs/:id", recruitmentHandler.UpdateJobPosting)
		v1.DELETE("/recruitment/jobs/:id", recruitmentHandler.DeleteJobPosting)

		v1.GET("/recruitment/applications", recruitmentHandler.GetApplications)
		v1.POST("/recruitment/applications", recruitmentHandler.CreateApplication)
		v1.GET("/recruitment/applications/:id", recruitmentHandler.GetApplication)
		v1.PUT("/recruitment/applications/:id", recruitmentHandler.UpdateApplication)

		// Performance
		v1.GET("/performance/reviews", performanceHandler.GetPerformanceReviews)
		v1.POST("/performance/reviews", performanceHandler.CreatePerformanceReview)
		v1.GET("/performance/reviews/:id", performanceHandler.GetPerformanceReview)
		v1.PUT("/performance/reviews/:id", performanceHandler.UpdatePerformanceReview)

		// Training
		v1.GET("/training/programs", trainingHandler.GetTrainingPrograms)
		v1.POST("/training/programs", trainingHandler.CreateTrainingProgram)
		v1.GET("/training/programs/:id", trainingHandler.GetTrainingProgram)
		v1.PUT("/training/programs/:id", trainingHandler.UpdateTrainingProgram)
		v1.POST("/training/programs/:id/enroll", trainingHandler.EnrollEmployee)
		v1.POST("/training/enrollments/:enrollmentId/complete", trainingHandler.CompleteTraining)


		// Document Management
		v1.GET("/employees/:id/documents", docHandler.GetEmployeeDocuments)
		v1.POST("/employees/:id/documents", docHandler.UploadEmployeeDocument)
		v1.DELETE("/employees/:id/documents/:docId", docHandler.DeleteEmployeeDocument)

		// Basic Reporting
		v1.GET("/reports/headcount", reportHandler.GetHeadcountReport)
		v1.GET("/reports/payroll", reportHandler.GetPayrollReport)
		v1.GET("/reports/attendance", reportHandler.GetAttendanceReport)
	}
}
