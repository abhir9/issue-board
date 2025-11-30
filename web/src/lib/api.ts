import axios from 'axios';
import { CreateIssueRequest, Issue, Label, UpdateIssueRequest, User } from '@/types';

import { toast } from 'sonner';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': process.env.NEXT_PUBLIC_API_KEY || '',
  },
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data || error.message || 'An error occurred';
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
  return data;
};

export const createIssue = async (issue: CreateIssueRequest) => {
  const { data } = await api.post<Issue>('/issues', issue);
  return data;
};

export const updateIssue = async (id: string, updates: UpdateIssueRequest) => {
  const { data } = await api.patch<Issue>(`/issues/${id}`, updates);
  return data;
};

export const moveIssue = async (id: string, status: string, orderIndex: number) => {
  await api.patch(`/issues/${id}/move`, { status, order_index: orderIndex });
};

export const deleteIssue = async (id: string) => {
  await api.delete(`/issues/${id}`);
};

export const getUsers = async () => {
  const { data } = await api.get<User[]>('/users');
  return data;
};

export const getLabels = async () => {
  const { data } = await api.get<Label[]>('/labels');
  return data;
};
