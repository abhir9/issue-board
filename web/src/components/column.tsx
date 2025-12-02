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
          <div ref={setNodeRef} className="min-h-[150px] space-y-2">
            {issues.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center gap-1.5 text-slate-500 p-4 min-h-[150px] border-2 border-dashed border-slate-200 rounded-lg bg-slate-50/50">
                <svg
                  className="w-8 h-8 text-slate-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={1.5}
                    d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                  />
                </svg>
                <span className="text-sm font-medium text-slate-700">No issues</span>
                <span className="text-xs text-slate-500 text-center max-w-[180px]">
                  No issues in this column.
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
