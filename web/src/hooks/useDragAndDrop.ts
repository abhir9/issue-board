import { useRef, useState, useEffect } from 'react';
import { useSensor, useSensors, PointerSensor } from '@dnd-kit/core';
import { arrayMove } from '@dnd-kit/sortable';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { DragStartEvent, DragEndEvent } from '@dnd-kit/core';
import { moveIssue } from '@/lib/api';
import { enqueue, waitForIdle, setDragging, isFullyIdle } from '@/lib/queue';
import { Issue, IssueStatus } from '@/types';
import { ISSUE_COLUMNS } from '@/constants/issues';

const DRAG_ACTIVATION_DISTANCE_PX = 5;
const REFETCH_DELAY_MS = 1000; // Debounce refetch to avoid race conditions

export function useDragAndDrop(
  issues: Issue[],
  filters: {
    assignee?: string;
    priority?: string[];
    labels?: string[];
  }
) {
  const queryClient = useQueryClient();
  const [activeIssue, setActiveIssue] = useState<Issue | null>(null);
  const isDraggingRef = useRef(false);
  const refetchTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (refetchTimeoutRef.current) {
        clearTimeout(refetchTimeoutRef.current);
      }
    };
  }, []);

  // Schedule a debounced refetch to sync with backend
  const scheduleRefetch = () => {
    // Clear any existing timeout
    if (refetchTimeoutRef.current) {
      clearTimeout(refetchTimeoutRef.current);
    }

    // Schedule new refetch after delay
    refetchTimeoutRef.current = setTimeout(() => {
      queryClient.invalidateQueries({ queryKey: ['issues', filters] });
      refetchTimeoutRef.current = null;
    }, REFETCH_DELAY_MS);
  };

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
    onMutate: async () => {
      // Cancel any outgoing refetches to avoid overwriting our optimistic update
      await queryClient.cancelQueries({ queryKey: ['issues', filters] });

      // Save current state for rollback, but DON'T re-apply optimistic update
      // (already done in onDragEnd to avoid double-updates causing reversals)
      const previousIssues = queryClient.getQueryData<Issue[]>(['issues', filters]);

      return { previousIssues };
    },
    onError: (err, newTodo, context) => {
      // Revert to previous state on error
      queryClient.setQueryData(['issues', filters], context?.previousIssues);
      queryClient.invalidateQueries({ queryKey: ['issues', filters] });
    },
    // Do not refetch on every success; we will refetch when the queue is idle
  });

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: DRAG_ACTIVATION_DISTANCE_PX,
      },
    })
  );

  function onDragStart(event: DragStartEvent) {
    isDraggingRef.current = true;
    setDragging(true);

    // Cancel any pending refetch when a new drag starts
    if (refetchTimeoutRef.current) {
      clearTimeout(refetchTimeoutRef.current);
      refetchTimeoutRef.current = null;
    }

    if (event.active.data.current?.type === 'Issue') {
      setActiveIssue(event.active.data.current.issue as Issue);
    }
  }

  function onDragOver() {
    // Removed optimistic updates from onDragOver to prevent race conditions
    // All state updates now happen in onDragEnd only
    return;
  }

  function onDragEnd(event: DragEndEvent) {
    const { active, over } = event;

    isDraggingRef.current = false;
    setDragging(false);
    setActiveIssue(null);

    if (!over) {
      return;
    }

    const activeId = active.id as string;
    const overId = over.id as string;

    if (activeId === overId) return;

    // Get fresh issues from cache to avoid stale closure
    const currentIssues = queryClient.getQueryData<Issue[]>(['issues', filters]) || [];

    const originalIssue = currentIssues.find((i) => i.id === activeId);
    if (!originalIssue) {
      return;
    }

    let targetStatus: IssueStatus = originalIssue.status;
    const isOverColumn = ISSUE_COLUMNS.some((c) => c.id === overId);

    if (isOverColumn) {
      targetStatus = overId as IssueStatus;
    } else {
      const overIssue = currentIssues.find((i) => i.id === overId);
      if (overIssue) {
        targetStatus = overIssue.status;
      }
    }

    queryClient.setQueryData<Issue[]>(['issues', filters], (old) => {
      if (!old) return old;

      const activeIndex = old.findIndex((i) => i.id === activeId);
      if (activeIndex === -1) return old;

      let newItems = [...old];

      if (!isOverColumn) {
        const overIndex = newItems.findIndex((i) => i.id === overId);
        if (overIndex !== -1) {
          newItems = arrayMove(newItems, activeIndex, overIndex);
        }
      }

      const finalIndex = newItems.findIndex((i) => i.id === activeId);
      if (newItems[finalIndex].status !== targetStatus) {
        newItems[finalIndex] = { ...newItems[finalIndex], status: targetStatus };
      }

      // Build ordered list of items in target column as they appear in the array
      const columnItemsInOrder: {
        id: string;
        originalOrderIndex: number;
        arrayPosition: number;
      }[] = [];
      newItems.forEach((item, idx) => {
        if (item.status === targetStatus) {
          columnItemsInOrder.push({
            id: item.id,
            originalOrderIndex: item.order_index,
            arrayPosition: idx,
          });
        }
      });

      // Find the position of the moved item in the column
      const movedItemPosInColumn = columnItemsInOrder.findIndex((item) => item.id === activeId);

      if (movedItemPosInColumn === -1) {
        // Should not happen, but fallback
        return newItems;
      }

      let newOrderIndex: number;
      if (columnItemsInOrder.length === 1) {
        // Only item in column
        newOrderIndex = 0;
      } else if (movedItemPosInColumn === 0) {
        // Moving to the top
        const nextItem = newItems.find((i) => i.id === columnItemsInOrder[1].id);
        newOrderIndex = nextItem ? nextItem.order_index - 1 : 0;
      } else if (movedItemPosInColumn === columnItemsInOrder.length - 1) {
        // Moving to the bottom
        const prevItem = newItems.find(
          (i) => i.id === columnItemsInOrder[movedItemPosInColumn - 1].id
        );
        newOrderIndex = prevItem ? prevItem.order_index + 1 : columnItemsInOrder.length - 1;
      } else {
        // Moving between two items
        const prevItem = newItems.find(
          (i) => i.id === columnItemsInOrder[movedItemPosInColumn - 1].id
        );
        const nextItem = newItems.find(
          (i) => i.id === columnItemsInOrder[movedItemPosInColumn + 1].id
        );
        if (prevItem && nextItem) {
          newOrderIndex = (prevItem.order_index + nextItem.order_index) / 2;
        } else {
          // Fallback
          newOrderIndex = movedItemPosInColumn;
        }
      }

      newItems[finalIndex] = { ...newItems[finalIndex], order_index: newOrderIndex };

      return newItems;
    });

    const updatedIssues = queryClient.getQueryData<Issue[]>(['issues', filters]);
    const movedIssue = updatedIssues?.find((i) => i.id === activeId);

    if (movedIssue) {
      // Enqueue network updates to run strictly sequentially.
      void enqueue(() =>
        moveIssueMutation.mutateAsync({
          id: activeId,
          status: movedIssue.status,
          orderIndex: movedIssue.order_index,
        })
      );

      // Refetch only after BOTH queue is empty AND no active drags
      void waitForIdle().then(() => {
        // Double-check we're fully idle before refetching
        if (isFullyIdle()) {
          scheduleRefetch();
        }
      });
    }
  }

  return {
    sensors,
    activeIssue,
    onDragStart,
    onDragOver,
    onDragEnd,
  };
}
