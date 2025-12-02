import axios from 'axios';
import { CreateIssueRequest, Issue, Label, UpdateIssueRequest, User } from '@/types';
import { issuesArraySchema, issueSchema, usersArraySchema, labelsArraySchema } from '@/lib/schemas';
import { toast } from 'sonner';

const api = axios.create({
  baseURL: '/api/proxy',
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Don't show toast for 404 errors - let the component handle it
    if (error.response?.status === 404) {
      return Promise.reject(error);
    }

    // Log error details for debugging
    if (process.env.NODE_ENV === 'development') {
      console.error('API Error:', {
        status: error.response?.status,
        data: error.response?.data,
        url: error.config?.url,
      });
    }

    const payload = error.response?.data;
    let message = error.message || 'An error occurred';

    if (payload) {
      if (typeof payload === 'string') {
        message = payload;
      } else if (typeof payload === 'object') {
        const maybeMessage =
          (payload as { message?: string; error?: string }).message ||
          (payload as { message?: string; error?: string }).error;
        message = maybeMessage || JSON.stringify(payload);
      }
    }

    toast.error(message);
    return Promise.reject(error);
  }
);

export const getIssues = async (filters?: {
  status?: string[];
  assignee?: string;
  priority?: string[];
  labels?: string[];
}) => {
  const params = new URLSearchParams();
  filters?.status?.forEach((s) => params.append('status', s));
  if (filters?.assignee) params.append('assignee', filters.assignee);
  filters?.priority?.forEach((p) => params.append('priority', p));
  filters?.labels?.forEach((l) => params.append('labels', l));

  const { data } = await api.get<Issue[]>('/issues', { params });
  
  // Handle null or undefined as empty array
  const normalizedData = data ?? [];
  
  // Validate API response
  const validated = issuesArraySchema.safeParse(normalizedData);
  if (!validated.success) {
    console.error('Invalid issues response:', validated.error);
    // For empty results, return empty array instead of throwing
    if (normalizedData.length === 0 || !Array.isArray(normalizedData)) {
      return [];
    }
    throw new Error('Invalid data received from server');
  }
  return validated.data;
};

export const createIssue = async (issue: CreateIssueRequest) => {
  const { data } = await api.post<Issue>('/issues', issue);
  // Validate API response
  const validated = issueSchema.safeParse(data);
  if (!validated.success) {
    console.error('Invalid issue response:', validated.error);
    throw new Error('Invalid data received from server');
  }
  return validated.data;
};

export const getIssue = async (id: string) => {
  const { data } = await api.get<Issue>(`/issues/${id}`);
  // Validate API response
  const validated = issueSchema.safeParse(data);
  if (!validated.success) {
    console.error('Invalid issue response:', validated.error);
    throw new Error('Invalid data received from server');
  }
  return validated.data;
};

export const updateIssue = async (id: string, updates: UpdateIssueRequest) => {
  const { data } = await api.patch<Issue>(`/issues/${id}`, updates);
  // Validate API response
  const validated = issueSchema.safeParse(data);
  if (!validated.success) {
    console.error('Invalid issue response:', validated.error);
    throw new Error('Invalid data received from server');
  }
  return validated.data;
};

export const moveIssue = async (id: string, status: string, orderIndex: number) => {
  await api.patch(`/issues/${id}/move`, { status, order_index: orderIndex });
};

export const deleteIssue = async (id: string) => {
  await api.delete(`/issues/${id}`);
};

export const getUsers = async () => {
  const { data } = await api.get<User[]>('/users');
  // Validate API response
  const validated = usersArraySchema.safeParse(data);
  if (!validated.success) {
    console.error('Invalid users response:', validated.error);
    throw new Error('Invalid data received from server');
  }
  return validated.data;
};

export const getLabels = async () => {
  const { data } = await api.get<Label[]>('/labels');
  // Validate API response
  const validated = labelsArraySchema.safeParse(data);
  if (!validated.success) {
    console.error('Invalid labels response:', validated.error);
    throw new Error('Invalid data received from server');
  }
  return validated.data;
};
