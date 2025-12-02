export type User = {
  id: string;
  name: string;
  avatar_url?: string;
};

export type Label = {
  id: string;
  name: string;
  color: string;
};

export type IssueStatus = 'Backlog' | 'Todo' | 'In Progress' | 'Done' | 'Canceled';
export type IssuePriority = 'Low' | 'Medium' | 'High' | 'Critical';

export type Issue = {
  id: string;
  title: string;
  description: string;
  status: IssueStatus;
  priority: IssuePriority;
  assignee_id?: string | null;
  assignee?: User | null;
  labels?: Label[];
  created_at: string;
  updated_at: string;
  order_index: number;
};

export type CreateIssueRequest = {
  title: string;
  description: string;
  status: IssueStatus;
  priority: IssuePriority;
  assignee_id?: string | null;
  label_ids: string[];
};

export type UpdateIssueRequest = Partial<CreateIssueRequest>;
