package router

import (
	"database/sql"
	"log/slog"

	"campushub/backend/internal/auth"
	"campushub/backend/internal/cloud"
	"campushub/backend/internal/domain"
	"campushub/backend/internal/handler"
	"campushub/backend/internal/middleware"
	"campushub/backend/internal/repository"
	"campushub/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	DB             *sql.DB
	Logger         *slog.Logger
	AuthProvider   auth.Provider
	TokenValidator middleware.TokenValidator
	Publisher      *cloud.EventPublisher
}

func New(deps Dependencies) *gin.Engine {
	tenantRepo := repository.NewTenantRepository(deps.DB)
	userRepo := repository.NewUserRepository(deps.DB)
	resourceRepo := repository.NewResourceRepository(deps.DB)
	bookingRepo := repository.NewBookingRepository(deps.DB)
	issueRepo := repository.NewIssueRepository(deps.DB)
	enrollmentRepo := repository.NewEnrollmentRepository(deps.DB)

	bookingSvc := service.NewBookingService(deps.DB, bookingRepo, resourceRepo, deps.Publisher, deps.Logger)
	enrollmentSvc := service.NewEnrollmentService(deps.DB, enrollmentRepo, bookingRepo, resourceRepo, deps.Logger)

	healthHandler := handler.NewHealthHandler(deps.DB)
	tenantHandler := handler.NewTenantHandler(tenantRepo)
	authHandler := handler.NewAuthHandler(userRepo, tenantRepo, deps.AuthProvider)
	resourceHandler := handler.NewResourceHandler(resourceRepo)
	bookingHandler := handler.NewBookingHandler(bookingSvc, bookingRepo)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentSvc)
	issueHandler := handler.NewIssueHandler(issueRepo, resourceRepo, deps.Publisher, deps.Logger)
	userHandler := handler.NewUserHandler(userRepo)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(deps.Logger), middleware.CORS())

	r.GET("/health", healthHandler.Check)
	r.GET("/tenants", tenantHandler.List)
	r.POST("/tenants", tenantHandler.Create)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)
	r.POST("/admin/promote-user", authHandler.PromoteUser)

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth(deps.TokenValidator), middleware.RequireTenant(userRepo))
	protected.GET("/me", handler.Me)
	protected.GET("/resources", resourceHandler.List)
	protected.GET("/resources/:id", resourceHandler.Get)
	protected.POST("/resources", middleware.RequireRole(domain.RoleAdmin), resourceHandler.Create)
	protected.PUT("/resources/:id", middleware.RequireRole(domain.RoleAdmin), resourceHandler.Update)
	protected.DELETE("/resources/:id", middleware.RequireRole(domain.RoleAdmin), resourceHandler.Delete)

	// Facility Bookings: Only Faculty and Admin can reserve physical resources
	protected.POST("/bookings", middleware.RequireRole(domain.RoleFaculty, domain.RoleAdmin), bookingHandler.Create)
	protected.GET("/bookings/mine", bookingHandler.ListMine)
	protected.GET("/bookings", middleware.RequireRole(domain.RoleAdmin, domain.RoleTechnician), bookingHandler.ListAll)
	protected.DELETE("/bookings/:id", bookingHandler.Cancel)

	// Event Discovery & Student Enrollment
	protected.GET("/events", enrollmentHandler.ListEvents)
	protected.GET("/events/enrolled", middleware.RequireRole(domain.RoleStudent), enrollmentHandler.ListMyEnrolled)
	protected.POST("/events/:id/enroll", middleware.RequireRole(domain.RoleStudent), enrollmentHandler.Enroll)
	protected.DELETE("/events/:id/enroll", middleware.RequireRole(domain.RoleStudent), enrollmentHandler.Unenroll)

	// Faculty Roster Management
	protected.GET("/events/:id/roster", middleware.RequireRole(domain.RoleFaculty, domain.RoleAdmin), enrollmentHandler.GetRoster)
	protected.DELETE("/events/:id/roster/:student_id", middleware.RequireRole(domain.RoleFaculty, domain.RoleAdmin), enrollmentHandler.RemoveStudent)

	protected.POST("/issues", issueHandler.Create)
	protected.GET("/issues", middleware.RequireRole(domain.RoleAdmin, domain.RoleTechnician), issueHandler.List)
	protected.GET("/issues/mine", issueHandler.ListMine)
	protected.GET("/issues/:id", issueHandler.Get)
	protected.PATCH("/issues/:id/status", middleware.RequireRole(domain.RoleAdmin, domain.RoleTechnician), issueHandler.UpdateStatus)
	protected.GET("/users", middleware.RequireRole(domain.RoleAdmin), userHandler.List)

	return r
}
