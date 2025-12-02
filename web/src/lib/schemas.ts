import { z } from 'zod';

// Form validation schemas
export const createIssueSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200, 'Title must be less than 200 characters'),
  description: z.string(),
  status: z.enum(['Backlog', 'Todo', 'In Progress', 'Done', 'Canceled'], {
    message: 'Status is required',
  }),
  priority: z.enum(['Low', 'Medium', 'High', 'Critical'], {
    message: 'Priority is required',
  }),
  assignee_id: z.string().optional(),
  label_ids: z.array(z.string()),
});

export type CreateIssueFormData = z.infer<typeof createIssueSchema>;

// API response validation schemas
export const labelSchema = z.object({
  id: z.string(),
  name: z.string(),
  color: z.string(),
});

export const userSchema = z.object({
  id: z.string(),
  name: z.string(),
  email: z.string().email().optional(),
  avatar_url: z.string().url().optional(),
});

export const issueSchema = z.object({
  id: z.string(),
  title: z.string(),
  description: z.string(),
  status: z.enum(['Backlog', 'Todo', 'In Progress', 'Done', 'Canceled']),
  priority: z.enum(['Low', 'Medium', 'High', 'Critical']),
  order_index: z.number(),
  assignee_id: z.string().nullable().optional(),
  assignee: userSchema.nullable().optional(),
  labels: z.array(labelSchema).optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const issuesArraySchema = z.array(issueSchema);
export const usersArraySchema = z.array(userSchema);
export const labelsArraySchema = z.array(labelSchema);
