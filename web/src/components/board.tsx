'use client';

import {
  DndContext,
  DragEndEvent,
  DragOverEvent,
  DragOverlay,
  DragStartEvent,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import { arrayMove } from '@dnd-kit/sortable';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { getIssues, moveIssue } from '@/lib/api';
import { Issue, IssueStatus } from '@/types';
import { Column } from './column';
import { IssueCard } from './issue-card';

import { useSearchParams, useRouter } from 'next/navigation';

const COLUMNS: { id: IssueStatus; title: string }[] = [
  { id: 'Backlog', title: 'Backlog' },
  { id: 'Todo', title: 'Todo' },
  { id: 'In Progress', title: 'In Progress' },
  { id: 'Done', title: 'Done' },
  { id: 'Canceled', title: 'Canceled' },
];

export function Board() {
  const queryClient = useQueryClient();
  const [activeIssue, setActiveIssue] = useState<Issue | null>(null);
  const [lastMovedId, setLastMovedId] = useState<string | null>(null);
  const [highlightedId, setHighlightedId] = useState<string | null>(null);
  const searchParams = useSearchParams();
  const router = useRouter();
  const processedHighlightRef = useRef<string | null>(null);

  // Handle highlight effect from URL params
  const highlightId = searchParams.get('highlight');

  useEffect(() => {
    if (highlightId && highlightId !== processedHighlightRef.current) {
      processedHighlightRef.current = highlightId;

      // Use requestAnimationFrame to avoid setState during render
      requestAnimationFrame(() => {
        setHighlightedId(highlightId);
      });

      // Clear the param without reload
      const newParams = new URLSearchParams(searchParams.toString());
      newParams.delete('highlight');
      router.replace(`/issues?${newParams.toString()}`, { scroll: false });

      // Clear highlight after 2 seconds
      setTimeout(() => {
        setHighlightedId(null);
        processedHighlightRef.current = null;
      }, 2000);
    }
  }, [highlightId, searchParams, router]);

  const filters = useMemo(
    () => ({
      assignee: searchParams.get('assignee') || undefined,
      priority: searchParams.get('priority') ? [searchParams.get('priority')!] : undefined,
    }),
    [searchParams]
  );

  const { data: issues = [], isLoading } = useQuery({
    queryKey: ['issues', filters],
    queryFn: () => getIssues(filters),
  });

  const isDraggingRef = useRef(false);

  // Use React Query cache as single source of truth
  const displayIssues = issues;

  const columns = useMemo(() => {
    const cols = new Map<IssueStatus, Issue[]>();
    COLUMNS.forEach((c) => cols.set(c.id, []));

    // Sort by order_index
    const sortedIssues = [...displayIssues].sort((a, b) => a.order_index - b.order_index);

    sortedIssues.forEach((issue) => {
      const col = cols.get(issue.status);
      if (col) col.push(issue);
    });
    return cols;
  }, [displayIssues]);

  // Memoize individual column arrays to prevent unnecessary re-renders
  const columnArrays = useMemo(() => {
    return COLUMNS.map((col) => ({
      id: col.id,
      title: col.title,
      issues: columns.get(col.id) || [],
    }));
  }, [columns]);

  const moveIssueMutation = useMutation({
    mutationFn: ({
      id,
      status,
      orderIndex,
    }: {
      id: string;
      status: IssueStatus;
      orderIndex: number;
    }) => moveIssue(id, status, orderIndex),
    onMutate: async ({ id, status, orderIndex }) => {
      await queryClient.cancelQueries({ queryKey: ['issues'] });
      const previousIssues = queryClient.getQueryData<Issue[]>(['issues']);

      queryClient.setQueryData<Issue[]>(['issues'], (old) => {
        if (!old) return [];
        return old.map((issue) => {
          if (issue.id === id) {
            return { ...issue, status, order_index: orderIndex };
          }
          return issue;
        });
      });

      return { previousIssues };
    },
    onError: (err, newTodo, context) => {
      queryClient.setQueryData(['issues', filters], context?.previousIssues);
    },
    onSuccess: () => {
      // Refetch to get authoritative server state
      queryClient.invalidateQueries({ queryKey: ['issues', filters] });
    },
  });

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 5, // Require 5px movement to start drag, allows clicks to pass through
      },
    })
  );

  function onDragStart(event: DragStartEvent) {
    isDraggingRef.current = true;
    if (event.active.data.current?.type === 'Issue') {
      setActiveIssue(event.active.data.current.issue as Issue);
    }
  }

  function onDragOver(event: DragOverEvent) {
    const { active, over } = event;
    if (!over) return;

    const activeId = active.id as string;
    const overId = over.id as string;

    if (activeId === overId) return;

    // Update React Query cache directly for immediate optimistic updates
    queryClient.setQueryData<Issue[]>(['issues', filters], (old) => {
      if (!old) return old;

      const activeIndex = old.findIndex((i) => i.id === activeId);
      if (activeIndex === -1) return old;

      const activeItem = old[activeIndex];

      // Check if we're over a column or an issue
      const isOverColumn = COLUMNS.some((c) => c.id === overId);

      if (isOverColumn) {
        // Dropped on a column droppable area
        const targetStatus = overId as IssueStatus;

        if (activeItem.status === targetStatus) {
          // Same column, no change needed
          return old;
        }

        // Moving to a different column - place at the end
        const newItems = [...old];
        newItems[activeIndex] = { ...activeItem, status: targetStatus };
        return newItems;
      } else {
        // Dropped on another issue
        const overIndex = old.findIndex((i) => i.id === overId);
        if (overIndex === -1) return old;

        const overItem = old[overIndex];
        const targetStatus = overItem.status;

        // Create new array with the item moved
        const newItems = arrayMove(old, activeIndex, overIndex);

        // Update status if moving to different column
        if (activeItem.status !== targetStatus) {
          const movedIndex = newItems.findIndex((i) => i.id === activeId);
          newItems[movedIndex] = { ...newItems[movedIndex], status: targetStatus };
        }

        return newItems;
      }
    });
  }

  function onDragEnd(event: DragEndEvent) {
    const { active, over } = event;

    isDraggingRef.current = false;
    setActiveIssue(null);

    if (!over) {
      // Revert optimistic update if dropped outside
      queryClient.invalidateQueries({ queryKey: ['issues', filters] });
      return;
    }

    const activeId = active.id as string;

    // Get current state from React Query cache (already updated by onDragOver)
    const currentIssues = queryClient.getQueryData<Issue[]>(['issues', filters]);
    const movedIssue = currentIssues?.find((i) => i.id === activeId);

    // Get original state to compare
    const originalIssues = queryClient.getQueryData<Issue[]>(['issues', filters]);
    const originalIssue = issues.find((i) => i.id === activeId);

    if (!movedIssue || !originalIssue) {
      queryClient.invalidateQueries({ queryKey: ['issues', filters] });
      return;
    }

    // Highlight the card after drop
    setLastMovedId(activeId);
    setTimeout(() => {
      setLastMovedId(null);
    }, 2000);

    // Only persist if status actually changed
    if (movedIssue.status !== originalIssue.status) {
      // Mutation will persist to server
      // onMutate will update cache again (redundant but safe)
      // onSuccess will refetch for authoritative state
      moveIssueMutation.mutate({
        id: activeId,
        status: movedIssue.status,
        orderIndex: movedIssue.order_index,
      });
    }
    // If no status change, cache already has correct state, no action needed
  }

  if (isLoading) {
    return (
      <div className="flex h-full gap-4 overflow-x-auto pb-4 px-4 md:px-0">
        {[1, 2, 3, 4, 5].map((i) => (
          <div
            key={i}
            className="shrink-0 w-80 bg-gray-50/50 rounded-lg border border-gray-200 h-full"
          >
            <div className="p-3 border-b border-gray-100 bg-white rounded-t-lg h-10 animate-pulse bg-gray-100" />
            <div className="p-2 space-y-3">
              {[1, 2, 3].map((j) => (
                <div
                  key={j}
                  className="h-24 bg-white rounded-lg border border-gray-200 animate-pulse"
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    );
  }

  return (
    <DndContext
      sensors={sensors}
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDragEnd={onDragEnd}
    >
      <div className="flex h-full gap-4 overflow-x-auto pb-4 snap-x snap-mandatory px-4 md:px-0">
        {columnArrays.map((col) => (
          <div key={col.id} className="snap-center shrink-0">
            <Column
              id={col.id}
              title={col.title}
              issues={col.issues}
              lastMovedId={lastMovedId}
              highlightedId={highlightedId}
            />
          </div>
        ))}
      </div>
      {typeof window !== 'undefined' &&
        createPortal(
          <DragOverlay>{activeIssue && <IssueCard issue={activeIssue} />}</DragOverlay>,
          document.body
        )}
    </DndContext>
  );
}
