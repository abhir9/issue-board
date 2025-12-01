import { useRef, useEffect, useMemo, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { getIssues } from '@/lib/api';
import { Issue, IssueStatus } from '@/types';
import { ISSUE_COLUMNS } from '@/constants/issues';
import { toast } from 'sonner';

const HIGHLIGHT_DURATION_MS = 2000;

export function useBoardData() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const processedHighlightRef = useRef<string | null>(null);
  const highlightTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Handle highlight effect from URL params
  const highlightId = searchParams.get('highlight');

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
  });

  useEffect(() => {
    if (error) {
      toast.error('Failed to load issues');
    }
  }, [error]);

  const columns = useMemo(() => {
    const cols = new Map<IssueStatus, Issue[]>();
    ISSUE_COLUMNS.forEach((c) => cols.set(c.id, []));

    // Sort by order_index
    const sortedIssues = [...issues].sort((a, b) => a.order_index - b.order_index);

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

  // Highlight state management
  const highlightedId = useHighlight(
    highlightId,
    searchParams,
    router,
    processedHighlightRef,
    highlightTimeoutRef
  );

  return {
    issues,
    isLoading,
    filters,
    columns: columnArrays,
    highlightedId,
    error,
  };
}

function useHighlight(
  highlightId: string | null,
  searchParams: URLSearchParams,
  router: ReturnType<typeof useRouter>,
  processedHighlightRef: React.MutableRefObject<string | null>,
  highlightTimeoutRef: React.MutableRefObject<NodeJS.Timeout | null>
): string | null {
  const [highlightedId, setHighlightedId] = useState<string | null>(null);

  useEffect(() => {
    if (highlightId && highlightId !== processedHighlightRef.current) {
      processedHighlightRef.current = highlightId;

      // Clear any existing timeout to prevent race conditions
      if (highlightTimeoutRef.current) {
        clearTimeout(highlightTimeoutRef.current);
      }

      // Use requestAnimationFrame to avoid setState during render
      requestAnimationFrame(() => {
        setHighlightedId(highlightId);
      });

      // Clear the param without reload
      const newParams = new URLSearchParams(searchParams.toString());
      newParams.delete('highlight');
      router.replace(`/issues?${newParams.toString()}`, { scroll: false });

      // Clear highlight after duration
      highlightTimeoutRef.current = setTimeout(() => {
        setHighlightedId(null);
        processedHighlightRef.current = null;
        highlightTimeoutRef.current = null;
      }, HIGHLIGHT_DURATION_MS);
    }

    // Cleanup on unmount
    return () => {
      if (highlightTimeoutRef.current) {
        clearTimeout(highlightTimeoutRef.current);
      }
    };
    // Refs don't need to be in dependency array as they're stable
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [highlightId, searchParams, router]);

  return highlightedId;
}
