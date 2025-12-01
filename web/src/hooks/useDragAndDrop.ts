import { useRef, useState } from 'react';
import { useSensor, useSensors, PointerSensor } from '@dnd-kit/core';
import { arrayMove } from '@dnd-kit/sortable';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { DragStartEvent, DragOverEvent, DragEndEvent } from '@dnd-kit/core';
import { moveIssue } from '@/lib/api';
import { Issue, IssueStatus } from '@/types';

const DRAG_ACTIVATION_DISTANCE_PX = 5;

const COLUMNS: { id: IssueStatus; title: string }[] = [
  { id: 'Backlog', title: 'Backlog' },
  { id: 'Todo', title: 'Todo' },
  { id: 'In Progress', title: 'In Progress' },
  { id: 'Done', title: 'Done' },
  { id: 'Canceled', title: 'Canceled' },
];

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
      await queryClient.cancelQueries({ queryKey: ['issues', filters] });
      const previousIssues = queryClient.getQueryData<Issue[]>(['issues', filters]);

      queryClient.setQueryData<Issue[]>(['issues', filters], (old) => {
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
      queryClient.invalidateQueries({ queryKey: ['issues', filters] });
    },
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

    const activeIssue = issues.find((i) => i.id === activeId);
    if (!activeIssue) return;

    const isOverColumn = COLUMNS.some((c) => c.id === overId);

    if (isOverColumn) {
      const targetStatus = overId as IssueStatus;

      if (activeIssue.status !== targetStatus) {
        queryClient.setQueryData<Issue[]>(['issues', filters], (old) => {
          if (!old) return old;

          const activeIndex = old.findIndex((i) => i.id === activeId);
          if (activeIndex === -1) return old;

          const newItems = [...old];
          newItems[activeIndex] = { ...newItems[activeIndex], status: targetStatus };
          return newItems;
        });
      }
    } else {
      const overIssue = issues.find((i) => i.id === overId);
      if (!overIssue) return;

      if (activeIssue.status !== overIssue.status) {
        queryClient.setQueryData<Issue[]>(['issues', filters], (old) => {
          if (!old) return old;

          const activeIndex = old.findIndex((i) => i.id === activeId);
          if (activeIndex === -1) return old;

          const newItems = [...old];
          newItems[activeIndex] = { ...newItems[activeIndex], status: overIssue.status };
          return newItems;
        });
      }
    }
  }

  function onDragEnd(event: DragEndEvent) {
    const { active, over } = event;

    isDraggingRef.current = false;
    setActiveIssue(null);

    if (!over) {
      queryClient.invalidateQueries({ queryKey: ['issues', filters] });
      return;
    }

    const activeId = active.id as string;
    const overId = over.id as string;

    if (activeId === overId) return;

    const originalIssue = issues.find((i) => i.id === activeId);
    if (!originalIssue) {
      queryClient.invalidateQueries({ queryKey: ['issues', filters] });
      return;
    }

    let targetStatus: IssueStatus = originalIssue.status;
    const isOverColumn = COLUMNS.some((c) => c.id === overId);

    if (isOverColumn) {
      targetStatus = overId as IssueStatus;
    } else {
      const overIssue = issues.find((i) => i.id === overId);
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

      const columnItems = newItems.filter((i) => i.id === activeId || i.status === targetStatus);
      const movedItemIndexInColumn = columnItems.findIndex((i) => i.id === activeId);

      let newOrderIndex: number;
      if (movedItemIndexInColumn === 0) {
        const nextItem = columnItems[1];
        newOrderIndex = nextItem ? nextItem.order_index - 1 : 0;
      } else if (movedItemIndexInColumn === columnItems.length - 1) {
        const prevItem = columnItems[movedItemIndexInColumn - 1];
        newOrderIndex = prevItem ? prevItem.order_index + 1 : columnItems.length - 1;
      } else {
        const prevItem = columnItems[movedItemIndexInColumn - 1];
        const nextItem = columnItems[movedItemIndexInColumn + 1];
        newOrderIndex = (prevItem.order_index + nextItem.order_index) / 2;
      }

      newItems[finalIndex] = { ...newItems[finalIndex], order_index: newOrderIndex };

      return newItems;
    });

    const updatedIssues = queryClient.getQueryData<Issue[]>(['issues', filters]);
    const movedIssue = updatedIssues?.find((i) => i.id === activeId);

    if (movedIssue) {
      moveIssueMutation.mutate({
        id: activeId,
        status: movedIssue.status,
        orderIndex: movedIssue.order_index,
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
