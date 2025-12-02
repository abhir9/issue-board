import { useEffect, useMemo } from 'react';
import { useSearchParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { getIssues } from '@/lib/api';
import { Issue, IssueStatus } from '@/types';
import { ISSUE_COLUMNS } from '@/constants/issues';
import { toast } from 'sonner';

export function useBoardData() {
  const searchParams = useSearchParams();

  const filters = useMemo(() => {
    const priorityParam = searchParams.get('priority');
    const labelParam = searchParams.get('labels');
    const statusParam = searchParams.get('status');

    return {
      assignee: searchParams.get('assignee') || undefined,
      priority: priorityParam && priorityParam !== 'all' ? [priorityParam] : undefined,
      labels: labelParam && labelParam !== 'all' ? [labelParam] : undefined,
      status: statusParam && statusParam !== 'all' ? [statusParam] : undefined,
    };
  }, [searchParams]);

  const {
    data: issues = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ['issues', filters],
    queryFn: () => getIssues(filters),
    retry: 1, // Only retry once to avoid excessive retries
    // Don't treat empty results as errors
    retryOnMount: false,
  });

  useEffect(() => {
    if (error) {
      // Log error details in development
      if (process.env.NODE_ENV === 'development') {
        console.error('Board data error:', error);
      }
      // Only show error toast for actual API errors, not for empty results
      toast.error('Failed to load issues. Please try again.');
    }
  }, [error]);

  const columns = useMemo(() => {
    const cols = new Map<IssueStatus, Issue[]>();
    ISSUE_COLUMNS.forEach((c) => cols.set(c.id, []));

    // Sort by order_index - ensure issues is an array
    const sortedIssues = (issues || []).sort((a, b) => a.order_index - b.order_index);

    sortedIssues.forEach((issue) => {
      const col = cols.get(issue.status);
      if (col) col.push(issue);
    });
    return cols;
  }, [issues]);

  const columnArrays = useMemo(() => {
    return ISSUE_COLUMNS.map((col) => ({
      id: col.id,
      title: col.title,
      issues: columns.get(col.id) || [],
    }));
  }, [columns]);

  return {
    issues,
    isLoading,
    filters,
    columns: columnArrays,
    error,
  };
}
