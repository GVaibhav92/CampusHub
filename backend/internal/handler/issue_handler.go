package handler

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"campushub/backend/internal/cloud"
	"campushub/backend/internal/domain"
	"campushub/backend/internal/middleware"
	"campushub/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IssueHandler struct {
	issues    repository.IssueRepository
	resources repository.ResourceRepository
	publisher *cloud.EventPublisher
	logger    *slog.Logger
}

func NewIssueHandler(issues repository.IssueRepository, resources repository.ResourceRepository, publisher *cloud.EventPublisher,
	logger *slog.Logger) *IssueHandler {
	return &IssueHandler{issues: issues, resources: resources, publisher: publisher, logger: logger}
}

// validTransitions is the entire status-transition business rule.
var validTransitions = map[domain.IssueStatus][]domain.IssueStatus{
	domain.IssueOpen:       {domain.IssueInProgress, domain.IssueClosed},
	domain.IssueInProgress: {domain.IssueResolved, domain.IssueOpen, domain.IssueClosed},
	domain.IssueResolved:   {domain.IssueClosed, domain.IssueInProgress},
	domain.IssueClosed:     {}, // terminal transitions
}

func isValidTransition(from, to domain.IssueStatus) bool {
	for _, allowed := range validTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

type createIssueRequest struct {
	ResourceID  string `json:"resource_id" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Priority    string `json:"priority"`
	Description string `json:"description" binding:"required"`
}

func (h *IssueHandler) Create(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	userID, _ := middleware.UserIDFromContext(c)

	var req createIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource_id"})
		return
	}

	// confirms the resource exists in this tenant
	resource, err := h.resources.GetByID(c.Request.Context(), tenantID, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify resource"})
		return
	}
	if resource == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		return
	}

	priority := domain.IssuePriority(req.Priority)
	if priority == "" {
		priority = domain.PriorityMedium
	}

	issue := &domain.Issue{
		ID: uuid.New(), TenantID: tenantID, ResourceID: resourceID, ReportedBy: userID,
		Category: req.Category, Priority: priority, Description: req.Description, Status: domain.IssueOpen,
	}
	if err := h.issues.Create(c.Request.Context(), issue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create issue"})
		return
	}
	if h.publisher != nil {
		event := cloud.NotificationEvent{
			Subject: "New Maintenance Issue Reported",
			Message: fmt.Sprintf("Issue reported for %q: %s (priority: %s, category: %s).", resource.Name, req.Description, priority, req.Category),
		}
		if err := h.publisher.Publish(c.Request.Context(), event); err != nil {
			h.logger.Error("failed to publish issue event", "error", err, "issue_id", issue.ID)
		}
	}
	c.JSON(http.StatusCreated, issue)
}

// List returns every issue in the tenant
func (h *IssueHandler) List(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	issues, err := h.issues.ListByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list issues"})
		return
	}
	c.JSON(http.StatusOK, issues)
}

func (h *IssueHandler) ListMine(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	userID, _ := middleware.UserIDFromContext(c)
	issues, err := h.issues.ListByReporter(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list issues"})
		return
	}
	c.JSON(http.StatusOK, issues)
}

// Get is visible to: the reporter, the assigned technician, or any
// ADMIN/TECHNICIAN. Everyone else gets 403.
func (h *IssueHandler) Get(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	role, _ := middleware.RoleFromContext(c)
	userID, _ := middleware.UserIDFromContext(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue id"})
		return
	}

	issue, err := h.issues.GetByID(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get issue"})
		return
	}
	if issue == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "issue not found"})
		return
	}

	isReporter := issue.ReportedBy == userID
	isAssigned := issue.AssignedTo != nil && *issue.AssignedTo == userID
	isStaff := role == domain.RoleAdmin || role == domain.RoleTechnician
	if !isReporter && !isAssigned && !isStaff {
		c.JSON(http.StatusForbidden, gin.H{"error": "not permitted to view this issue"})
		return
	}
	c.JSON(http.StatusOK, issue)
}

type updateStatusRequest struct {
	Status     string  `json:"status" binding:"required"`
	AssignedTo *string `json:"assigned_to"`
}

// UpdateStatus is route-gated to ADMIN/TECHNICIAN. It re-fetches the
// current issue to validate the transition against its actual current status
func (h *IssueHandler) UpdateStatus(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue id"})
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newStatus := domain.IssueStatus(req.Status)

	issue, err := h.issues.GetByID(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get issue"})
		return
	}
	if issue == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "issue not found"})
		return
	}
	if !isValidTransition(issue.Status, newStatus) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition from " + string(issue.Status) + " to " + string(newStatus)})
		return
	}

	var assignedTo *uuid.UUID
	if req.AssignedTo != nil {
		parsed, err := uuid.Parse(*req.AssignedTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assigned_to id"})
			return
		}
		assignedTo = &parsed
	}

	err = h.issues.UpdateStatus(c.Request.Context(), tenantID, id, issue.Status, newStatus, assignedTo)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusConflict, gin.H{"error": "issue status changed concurrently, please retry"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update issue status"})
		return
	}

	updated, err := h.issues.GetByID(c.Request.Context(), tenantID, id)
	if err != nil || updated == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "status updated but failed to reload issue"})
		return
	}
	c.JSON(http.StatusOK, updated)
}
