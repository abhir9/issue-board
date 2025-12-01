'use client';

import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { Issue, IssueStatus } from '@/types';
import { IssueCard } from './issue-card';

interface ColumnProps {
  id: IssueStatus;
  title: string;
  issues: Issue[];
  highlightedId?: string | null;
}

export function Column({ id, title, issues, highlightedId }: ColumnProps) {
  const { setNodeRef } = useDroppable({ id });

  return (
    <div className="flex flex-col bg-gray-50/50 rounded-lg border border-gray-200 w-80 shrink-0">
      <div className="p-3 font-semibold text-sm flex justify-between items-center border-b border-gray-100 bg-white rounded-t-lg">
        {title}
        <span className="text-xs text-gray-500 bg-gray-100 px-2 py-0.5 rounded-full">
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
              <div className="h-full flex flex-col items-center justify-center text-gray-400 p-4 min-h-[150px] border-2 border-dashed border-gray-200 rounded-lg">
                <span className="text-sm">No issues</span>
              </div>
            ) : (
              issues.map((issue) => (
                <IssueCard
                  key={issue.id}
                  issue={issue}
                  isHighlighted={issue.id === highlightedId}
                />
              ))
            )}
          </div>
        </SortableContext>
      </div>
    </div>
  );
}
