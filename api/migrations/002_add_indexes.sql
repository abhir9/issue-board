-- Add indexes for frequently queried columns

CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
CREATE INDEX IF NOT EXISTS idx_issues_priority ON issues(priority);
CREATE INDEX IF NOT EXISTS idx_issues_assignee_id ON issues(assignee_id);
CREATE INDEX IF NOT EXISTS idx_issues_created_at ON issues(created_at);
CREATE INDEX IF NOT EXISTS idx_issues_order_index ON issues(order_index);
CREATE INDEX IF NOT EXISTS idx_issue_labels_issue_id ON issue_labels(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_labels_label_id ON issue_labels(label_id);
