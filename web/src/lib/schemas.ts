import { z } from 'zod';

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
