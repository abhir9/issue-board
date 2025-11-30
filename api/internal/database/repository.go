package database

import (
	"api/internal/models"
	"database/sql"
	"fmt"
	"strings"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// GetIssues retrieves issues with optional filters and pagination
func (r *Repository) GetIssues(status []string, assigneeID string, priority []string, labels []string, page, pageSize int) ([]models.Issue, error) {
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
		// Filter issues that have at least one of the specified labels
		query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM issue_labels il WHERE il.issue_id = i.id AND il.label_id IN (%s))", strings.Join(placeholders, ","))
	}

	query += " ORDER BY i.order_index ASC"

	// Add pagination
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		query += " LIMIT ? OFFSET ?"
		args = append(args, pageSize, offset)
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []models.Issue
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
			return nil, err
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

		// Fetch labels for each issue (N+1 problem, but fine for small scale)
		labels, err := r.GetLabelsForIssue(i.ID)
		if err != nil {
			return nil, err
		}
		i.Labels = labels

		issues = append(issues, i)
	}
	return issues, nil
}

func (r *Repository) GetLabelsForIssue(issueID string) ([]models.Label, error) {
	query := `
		SELECT l.id, l.name, l.color
		FROM labels l
		JOIN issue_labels il ON l.id = il.label_id
		WHERE il.issue_id = ?
	`
	rows, err := r.DB.Query(query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []models.Label
	for rows.Next() {
		var l models.Label
		if err := rows.Scan(&l.ID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, nil
}

func (r *Repository) CreateIssue(issue models.Issue) error {
	query := `
		INSERT INTO issues (id, title, description, status, priority, assignee_id, created_at, updated_at, order_index)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.DB.Exec(query, issue.ID, issue.Title, issue.Description, issue.Status, issue.Priority, issue.AssigneeID, issue.CreatedAt, issue.UpdatedAt, issue.OrderIndex)
	return err
}

func (r *Repository) GetIssue(id string) (*models.Issue, error) {
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

	err := r.DB.QueryRow(query, id).Scan(
		&i.ID, &i.Title, &i.Description, &i.Status, &i.Priority, &assigneeID, &i.CreatedAt, &i.UpdatedAt, &i.OrderIndex,
		&userID, &userName, &userAvatar,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
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

	labels, err := r.GetLabelsForIssue(i.ID)
	if err != nil {
		return nil, err
	}
	i.Labels = labels

	return &i, nil
}

func (r *Repository) UpdateIssue(id string, updates map[string]interface{}) error {
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

	_, err := r.DB.Exec(query, args...)
	return err
}

func (r *Repository) UpdateIssueLabels(issueID string, labelIDs []string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	// Delete existing
	_, err = tx.Exec("DELETE FROM issue_labels WHERE issue_id = ?", issueID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert new
	stmt, err := tx.Prepare("INSERT INTO issue_labels (issue_id, label_id) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, labelID := range labelIDs {
		_, err = stmt.Exec(issueID, labelID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) DeleteIssue(id string) error {
	_, err := r.DB.Exec("DELETE FROM issues WHERE id = ?", id)
	return err
}

func (r *Repository) GetUsers() ([]models.User, error) {
	rows, err := r.DB.Query("SELECT id, name, avatar_url FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var avatarURL sql.NullString
		if err := rows.Scan(&u.ID, &u.Name, &avatarURL); err != nil {
			return nil, err
		}
		if avatarURL.Valid {
			u.AvatarURL = avatarURL.String
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *Repository) GetLabels() ([]models.Label, error) {
	rows, err := r.DB.Query("SELECT id, name, color FROM labels")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []models.Label
	for rows.Next() {
		var l models.Label
		if err := rows.Scan(&l.ID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, nil
}
