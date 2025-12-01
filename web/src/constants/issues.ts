import { IssueStatus } from '@/types';

export const ISSUE_STATUSES: IssueStatus[] = ['Backlog', 'Todo', 'In Progress', 'Done', 'Canceled'];

export const ISSUE_COLUMNS = ISSUE_STATUSES.map((status) => ({
  id: status,
  title: status,
}));
