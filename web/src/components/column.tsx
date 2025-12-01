'use client';

import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { Issue, IssueStatus } from '@/types';
import { IssueCard } from './issue-card';
import { memo } from 'react';

interface ColumnProps {
  id: IssueStatus;
  title: string;
  issues: Issue[];
}

function ColumnComponent({ id, title, issues }: ColumnProps) {
  const { setNodeRef } = useDroppable({ id });

  return (
    <div className="flex flex-col bg-white rounded-xl border border-slate-200 shadow-sm w-80 shrink-0">
      <div className="p-3 font-semibold text-sm text-slate-900 flex justify-between items-center border-b border-slate-200 bg-slate-50 rounded-t-xl">
        {title}
        <span className="text-xs font-medium text-slate-800 bg-slate-200/80 px-2 py-0.5 rounded-full">
          {issues.length}
        </span>
      </div>
      <div className="p-2">
        <SortableContext
          id={id}
          items={issues.map((i) => i.id)}
          strategy={verticalListSortingStrategy}
        >
          <div ref={setNodeRef} className="min-h-[150px]">
            {issues.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center gap-1 text-slate-600 p-4 min-h-[150px] border-2 border-dashed border-slate-300 rounded-lg bg-slate-50">
                <span className="text-sm font-medium text-slate-800">No issues</span>
                <span className="text-xs text-slate-600 text-center">
                  Drag or create an issue here to populate this column
                </span>
              </div>
            ) : (
              issues.map((issue) => <IssueCard key={issue.id} issue={issue} />)
            )}
          </div>
        </SortableContext>
      </div>
    </div>
  );
}

// Memoize to prevent unnecessary re-renders
export const Column = memo(ColumnComponent);
