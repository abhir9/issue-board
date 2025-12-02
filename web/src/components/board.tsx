'use client';

import { DndContext, DragOverlay } from '@dnd-kit/core';
import { createPortal } from 'react-dom';
import { useBoardData } from '@/hooks/useBoardData';
import { useDragAndDrop } from '@/hooks/useDragAndDrop';
import { Column } from './column';
import { IssueCard } from './issue-card';

export function Board() {
  const { issues, isLoading, filters, columns, error } = useBoardData();
  const { sensors, activeIssue, onDragStart, onDragOver, onDragEnd } = useDragAndDrop(
    issues,
    filters
  );

  // Only show error for actual errors, not for empty results
  if (error) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-red-500">
        Unable to load issues. Please try again.
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex h-full gap-4 overflow-x-auto pb-4 px-4 md:px-0">
        {[1, 2, 3, 4, 5].map((i) => (
          <div
            key={i}
            className="shrink-0 w-80 bg-white rounded-xl border border-slate-200 h-full shadow-sm"
          >
            <div className="p-3 border-b border-slate-200 bg-slate-100 rounded-t-xl h-10 animate-pulse" />
            <div className="p-2 space-y-3 bg-slate-50">
              {[1, 2, 3].map((j) => (
                <div
                  key={j}
                  className="h-24 bg-white rounded-lg border border-slate-200/80 animate-pulse"
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
        {columns.map((col) => (
          <div key={col.id} className="snap-center shrink-0">
            <Column id={col.id} title={col.title} issues={col.issues} />
          </div>
        ))}
      </div>
      {createPortal(
        <DragOverlay>{activeIssue && <IssueCard issue={activeIssue} />}</DragOverlay>,
        document.body
      )}
    </DndContext>
  );
}
