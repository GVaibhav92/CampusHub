package router

import (
	"database/sql"
	"log/slog"
	"time"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/handler"
	"campushub/backend/internal/middleware"
	"campushub/backend/internal/repository"
	"campushub/backend/internal/service"

	"github.com/gin-gonic/gin"
)

const tokenTTL = 24 * time.Hour

type Dependencies struct {
	DB        *sql.DB
	Logger    *slog.Logger
	JWTSecret string
}

func New(deps Dependencies) *gin.Engine {
	tenantRepo := repository.NewTenantRepository(deps.DB)
	userRepo := repository.NewUserRepository(deps.DB)
	resourceRepo := repository.NewResourceRepository(deps.DB)
	bookingRepo := repository.NewBookingRepository(deps.DB)
	issueRepo := repository.NewIssueRepository(deps.DB)
	issuer := middleware.NewTokenIssuer(deps.JWTSecret, tokenTTL)
	bookingSvc := service.NewBookingService(deps.DB, bookingRepo, resourceRepo)

	healthHandler := handler.NewHealthHandler(deps.DB)
	tenantHandler := handler.NewTenantHandler(tenantRepo)
	userHandler := handler.NewUserHandler(userRepo)
	authHandler := handler.NewAuthHandler(userRepo, tenantRepo, issuer)
	resourceHandler := handler.NewResourceHandler(resourceRepo)
	bookingHandler := handler.NewBookingHandler(bookingSvc, bookingRepo)
	issueHandler := handler.NewIssueHandler(issueRepo, resourceRepo)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(deps.Logger))

	r.GET("/health", healthHandler.Check)
	r.POST("/tenants", tenantHandler.Create)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth(issuer), middleware.RequireTenant(userRepo))
	protected.GET("/me", handler.Me)

	// Resources: any authenticated tenant member can read; only ADMIN writes.
	protected.GET("/resources", resourceHandler.List)
	protected.GET("/resources/:id", resourceHandler.Get)
	protected.POST("/resources", middleware.RequireRole(domain.RoleAdmin), resourceHandler.Create)
	protected.PUT("/resources/:id", middleware.RequireRole(domain.RoleAdmin), resourceHandler.Update)
	protected.DELETE("/resources/:id", middleware.RequireRole(domain.RoleAdmin), resourceHandler.Delete)

	// Staff-only tenant-wide visibility — both reuse Phase 2 repository
	// methods that already existed but weren't exposed via a route.
	protected.GET("/bookings", middleware.RequireRole(domain.RoleAdmin, domain.RoleTechnician), bookingHandler.ListAll)
	protected.GET("/users", middleware.RequireRole(domain.RoleAdmin), userHandler.List)

	// Bookings: any authenticated tenant member creates/lists/cancels own;
	// ownership + admin-override check happens in the service layer.
	protected.POST("/bookings", bookingHandler.Create)
	protected.GET("/bookings/mine", bookingHandler.ListMine)
	protected.DELETE("/bookings/:id", bookingHandler.Cancel)

	// Issues: any authenticated tenant member can report; only ADMIN/
	// TECHNICIAN can list all or change status. Visibility for a single
	// issue (reporter/assignee/staff) is enforced inside the handler.
	protected.POST("/issues", issueHandler.Create)
	protected.GET("/issues", middleware.RequireRole(domain.RoleAdmin, domain.RoleTechnician), issueHandler.List)
	protected.GET("/issues/mine", issueHandler.ListMine)
	protected.GET("/issues/:id", issueHandler.Get)
	protected.PATCH("/issues/:id/status", middleware.RequireRole(domain.RoleAdmin, domain.RoleTechnician), issueHandler.UpdateStatus)

	return r
}
