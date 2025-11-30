package handlers

import (
	"api/internal/database"
	"api/internal/models"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"api/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Repo *database.Repository
}

func NewHandler(repo *database.Repository) *Handler {
	return &Handler{Repo: repo}
}

// GetIssues godoc
// @Summary Get all issues
// @Description Get a list of issues, optionally filtered by status, assignee, priority, or labels
// @Tags issues
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param assignee query string false "Filter by assignee ID"
// @Param priority query string false "Filter by priority"
// @Param labels query string false "Filter by label name (e.g., ?labels=bug)"
// @Success 200 {array} models.Issue
// @Failure 500 {string} string "Internal Server Error"
// @Router /issues [get]
// @Security ApiKeyAuth
func (h *Handler) GetIssues(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query()["status"]
	assignee := r.URL.Query().Get("assignee")
	priority := r.URL.Query()["priority"]
	labels := r.URL.Query()["labels"]

	// Parse pagination parameters
	page := 1
	pageSize := 0 // 0 means no pagination
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	issues, err := h.Repo.GetIssues(status, assignee, priority, labels, page, pageSize)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch issues", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, issues)
}

// CreateIssue godoc
// @Summary Create a new issue
// @Description Create a new issue with the provided details
// @Tags issues
// @Accept json
// @Produce json
// @Param issue body models.CreateIssueRequest true "Issue content"
// @Success 201 {object} models.Issue
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /issues [post]
// @Security ApiKeyAuth
func (h *Handler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	// Get minimum order_index for this status column to place new issue at the top
	existingIssues, err := h.Repo.GetIssues([]string{req.Status}, "", nil, nil, 1, 0)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch existing issues", map[string]interface{}{"error": err.Error()})
		return
	}

	// Calculate order_index: find min and subtract 1 to place at top
	orderIndex := 0.0
	if len(existingIssues) > 0 {
		minIndex := existingIssues[0].OrderIndex
		for _, issue := range existingIssues {
			if issue.OrderIndex < minIndex {
				minIndex = issue.OrderIndex
			}
		}
		orderIndex = minIndex - 1
	}

	issue := models.Issue{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		AssigneeID:  req.AssigneeID,
		CreatedAt:   now,
		UpdatedAt:   now,
		OrderIndex:  orderIndex,
	}

	if err := h.Repo.CreateIssue(issue); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to create issue", map[string]interface{}{"error": err.Error()})
		return
	}

	if len(req.LabelIDs) > 0 {
		if err := h.Repo.UpdateIssueLabels(id, req.LabelIDs); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to update labels", map[string]interface{}{"error": err.Error()})
			return
		}
	}

	createdIssue, err := h.Repo.GetIssue(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch created issue", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, createdIssue)
}

// GetIssue godoc
// @Summary Get a specific issue
// @Description Get details of a specific issue by ID
// @Tags issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Success 200 {object} models.Issue
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /issues/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetIssue(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	issue, err := h.Repo.GetIssue(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch issue", map[string]interface{}{"error": err.Error()})
		return
	}
	if issue == nil {
		utils.WriteError(w, http.StatusNotFound, "Issue not found", nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, issue)
}

// UpdateIssue godoc
// @Summary Update an issue
// @Description Update details of an existing issue
// @Tags issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Param issue body models.UpdateIssueRequest true "Issue updates"
// @Success 200 {object} models.Issue
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /issues/{id} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req models.UpdateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.AssigneeID != nil {
		updates["assignee_id"] = *req.AssigneeID
	}
	updates["updated_at"] = time.Now()

	if err := h.Repo.UpdateIssue(id, updates); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update issue", map[string]interface{}{"error": err.Error()})
		return
	}

	if req.LabelIDs != nil {
		if err := h.Repo.UpdateIssueLabels(id, req.LabelIDs); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to update labels", map[string]interface{}{"error": err.Error()})
			return
		}
	}

	updatedIssue, err := h.Repo.GetIssue(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch updated issue", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, updatedIssue)
}

// MoveIssue godoc
// @Summary Move an issue
// @Description Move an issue to a new status and/or order
// @Tags issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Param move body models.UpdateIssueRequest true "Move details (status and order_index)"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /issues/{id}/move [patch]
// @Security ApiKeyAuth
func (h *Handler) MoveIssue(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req models.UpdateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.OrderIndex != nil {
		updates["order_index"] = *req.OrderIndex
	}

	if err := h.Repo.UpdateIssue(id, updates); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update issue", map[string]interface{}{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteIssue godoc
// @Summary Delete an issue
// @Description Delete an issue by ID
// @Tags issues
// @Param id path string true "Issue ID"
// @Success 204 {object} nil
// @Failure 500 {string} string "Internal Server Error"
// @Router /issues/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Repo.DeleteIssue(id); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to delete issue", map[string]interface{}{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetUsers godoc
// @Summary Get all users
// @Description Get a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {string} string "Internal Server Error"
// @Router /users [get]
// @Security ApiKeyAuth
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Repo.GetUsers()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch users", map[string]interface{}{"error": err.Error()})
		return
	}
	utils.WriteJSON(w, http.StatusOK, users)
}

// GetLabels godoc
// @Summary Get all labels
// @Description Get a list of all labels
// @Tags labels
// @Accept json
// @Produce json
// @Success 200 {array} models.Label
// @Failure 500 {string} string "Internal Server Error"
// @Router /labels [get]
// @Security ApiKeyAuth
func (h *Handler) GetLabels(w http.ResponseWriter, r *http.Request) {
	labels, err := h.Repo.GetLabels()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch labels", map[string]interface{}{"error": err.Error()})
		return
	}
	utils.WriteJSON(w, http.StatusOK, labels)
}
