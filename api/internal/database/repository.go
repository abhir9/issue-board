package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/abhir9/issue-board/api/internal/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// GetIssues retrieves issues with optional filters and pagination
func (r *Repository) GetIssues(ctx context.Context, status []string, assigneeID string, priority []string, labels []string, page, pageSize int) ([]models.Issue, error) {
	query := `
		SELECT i.id, i.title, i.description, i.status, i.priority, i.assignee_id, i.created_at, i.updated_at, i.order_index,
		       u.id, u.name, u.avatar_url
		FROM issues i
		LEFT JOIN users u ON i.assignee_id = u.id
		WHERE 1=1
	`
	var args []interface{}

	if len(status) > 0 {
		placeholders := make([]string, len(status))
		for i, s := range status {
			placeholders[i] = "?"
			args = append(args, s)
		}
		query += fmt.Sprintf(" AND i.status IN (%s)", strings.Join(placeholders, ","))
	}

	if assigneeID != "" {
		query += " AND i.assignee_id = ?"
		args = append(args, assigneeID)
	}

	if len(priority) > 0 {
		placeholders := make([]string, len(priority))
		for i, p := range priority {
			placeholders[i] = "?"
			args = append(args, p)
		}
		query += fmt.Sprintf(" AND i.priority IN (%s)", strings.Join(placeholders, ","))
	}

	if len(labels) > 0 {
		placeholders := make([]string, len(labels))
		for i, l := range labels {
			placeholders[i] = "?"
			args = append(args, l)
		}
		// Filter issues that have at least one of the specified labels (by label name)
		query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM issue_labels il JOIN labels l ON il.label_id = l.id WHERE il.issue_id = i.id AND l.name IN (%s))", strings.Join(placeholders, ","))
	}

	query += " ORDER BY i.order_index ASC"

	// Add pagination
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		query += " LIMIT ? OFFSET ?"
		args = append(args, pageSize, offset)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()

	var issues []models.Issue
	issueIDs := make([]string, 0)
	
	for rows.Next() {
		var i models.Issue
		var u models.User
		var assigneeID sql.NullString
		var userID sql.NullString
		var userName sql.NullString
		var userAvatar sql.NullString

		err := rows.Scan(
			&i.ID, &i.Title, &i.Description, &i.Status, &i.Priority, &assigneeID, &i.CreatedAt, &i.UpdatedAt, &i.OrderIndex,
			&userID, &userName, &userAvatar,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}

		if assigneeID.Valid {
			i.AssigneeID = &assigneeID.String
			if userID.Valid {
				u.ID = userID.String
				u.Name = userName.String
				u.AvatarURL = userAvatar.String
				i.Assignee = &u
			}
		}

		issues = append(issues, i)
		issueIDs = append(issueIDs, i.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issues: %w", err)
	}

	// Fetch all labels for all issues in one query (solves N+1 problem)
	if len(issueIDs) > 0 {
		labelMap, err := r.GetLabelsForIssues(ctx, issueIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch labels: %w", err)
		}

		// Attach labels to issues
		for i := range issues {
			if labels, ok := labelMap[issues[i].ID]; ok {
				issues[i].Labels = labels
			} else {
				issues[i].Labels = []models.Label{}
			}
		}
	}

	return issues, nil
}

func (r *Repository) GetLabelsForIssue(ctx context.Context, issueID string) ([]models.Label, error) {
	query := `
		SELECT l.id, l.name, l.color
		FROM labels l
		JOIN issue_labels il ON l.id = il.label_id
		WHERE il.issue_id = ?
	`
	rows, err := r.DB.QueryContext(ctx, query, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to query labels for issue: %w", err)
	}
	defer rows.Close()

	var labels []models.Label
	for rows.Next() {
		var l models.Label
		if err := rows.Scan(&l.ID, &l.Name, &l.Color); err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, l)
	}
	return labels, nil
}

// GetLabelsForIssues retrieves labels for multiple issues in a single query (batch fetch)
func (r *Repository) GetLabelsForIssues(ctx context.Context, issueIDs []string) (map[string][]models.Label, error) {
	if len(issueIDs) == 0 {
		return make(map[string][]models.Label), nil
	}

	placeholders := make([]string, len(issueIDs))
	args := make([]interface{}, len(issueIDs))
	for i, id := range issueIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT il.issue_id, l.id, l.name, l.color
		FROM labels l
		JOIN issue_labels il ON l.id = il.label_id
		WHERE il.issue_id IN (%s)
		ORDER BY il.issue_id, l.name
	`, strings.Join(placeholders, ","))

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query labels for issues: %w", err)
	}
	defer rows.Close()

	labelMap := make(map[string][]models.Label)
	for rows.Next() {
		var issueID string
		var l models.Label
		if err := rows.Scan(&issueID, &l.ID, &l.Name, &l.Color); err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labelMap[issueID] = append(labelMap[issueID], l)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating labels: %w", err)
	}

	return labelMap, nil
}

func (r *Repository) CreateIssue(ctx context.Context, issue models.Issue) error {
	query := `
		INSERT INTO issues (id, title, description, status, priority, assignee_id, created_at, updated_at, order_index)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.DB.ExecContext(ctx, query, issue.ID, issue.Title, issue.Description, issue.Status, issue.Priority, issue.AssigneeID, issue.CreatedAt, issue.UpdatedAt, issue.OrderIndex)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}
	return nil
}

func (r *Repository) GetIssue(ctx context.Context, id string) (*models.Issue, error) {
	query := `
		SELECT i.id, i.title, i.description, i.status, i.priority, i.assignee_id, i.created_at, i.updated_at, i.order_index,
		       u.id, u.name, u.avatar_url
		FROM issues i
		LEFT JOIN users u ON i.assignee_id = u.id
		WHERE i.id = ?
	`
	var i models.Issue
	var u models.User
	var assigneeID sql.NullString
	var userID sql.NullString
	var userName sql.NullString
	var userAvatar sql.NullString

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&i.ID, &i.Title, &i.Description, &i.Status, &i.Priority, &assigneeID, &i.CreatedAt, &i.UpdatedAt, &i.OrderIndex,
		&userID, &userName, &userAvatar,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	if assigneeID.Valid {
		i.AssigneeID = &assigneeID.String
		if userID.Valid {
			u.ID = userID.String
			u.Name = userName.String
			u.AvatarURL = userAvatar.String
			i.Assignee = &u
		}
	}

	labels, err := r.GetLabelsForIssue(ctx, i.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels for issue: %w", err)
	}
	i.Labels = labels

	return &i, nil
}

func (r *Repository) UpdateIssue(ctx context.Context, id string, updates map[string]interface{}) error {
	// Dynamic update query
	query := "UPDATE issues SET "
	var args []interface{}
	var parts []string

	for k, v := range updates {
		parts = append(parts, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	if len(parts) == 0 {
		return nil
	}

	query += strings.Join(parts, ", ") + " WHERE id = ?"
	args = append(args, id)

	result, err := r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("issue not found")
	}

	return nil
}

func (r *Repository) UpdateIssueLabels(ctx context.Context, issueID string, labelIDs []string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing
	_, err = tx.ExecContext(ctx, "DELETE FROM issue_labels WHERE issue_id = ?", issueID)
	if err != nil {
		return fmt.Errorf("failed to delete existing labels: %w", err)
	}

	// Insert new
	if len(labelIDs) > 0 {
		stmt, err := tx.PrepareContext(ctx, "INSERT INTO issue_labels (issue_id, label_id) VALUES (?, ?)")
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, labelID := range labelIDs {
			_, err = stmt.ExecContext(ctx, issueID, labelID)
			if err != nil {
				return fmt.Errorf("failed to insert label: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) DeleteIssue(ctx context.Context, id string) error {
	result, err := r.DB.ExecContext(ctx, "DELETE FROM issues WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("issue not found")
	}

	return nil
}

func (r *Repository) GetUsers(ctx context.Context) ([]models.User, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, avatar_url FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var avatarURL sql.NullString
		if err := rows.Scan(&u.ID, &u.Name, &avatarURL); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		if avatarURL.Valid {
			u.AvatarURL = avatarURL.String
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *Repository) GetLabels(ctx context.Context) ([]models.Label, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, color FROM labels")
	if err != nil {
		return nil, fmt.Errorf("failed to query labels: %w", err)
	}
	defer rows.Close()

	var labels []models.Label
	for rows.Next() {
		var l models.Label
		if err := rows.Scan(&l.ID, &l.Name, &l.Color); err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, l)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating labels: %w", err)
	}

	return labels, nil
}
