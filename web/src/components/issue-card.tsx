'use client';

import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Issue } from '@/types';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import Link from 'next/link';
import { useHighlightContext } from '@/contexts/HighlightContext';

interface IssueCardProps {
  issue: Issue;
}

export function IssueCard({ issue }: IssueCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: issue.id,
    data: { type: 'Issue', issue },
  });

  // Get highlighted state from context
  const { highlightedId } = useHighlightContext();
  const isHighlighted = issue.id === highlightedId;

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.3 : 1, // Make it semi-transparent
    border: isDragging ? '2px dashed #3b82f6' : undefined, // Add dashed border
    background: isDragging ? '#eff6ff' : undefined, // Light blue background
    height: isDragging ? 'auto' : undefined,
  };

  const priorityColors = {
    Low: 'bg-blue-100 text-blue-800',
    Medium: 'bg-yellow-100 text-yellow-800',
    High: 'bg-orange-100 text-orange-800',
    Critical: 'bg-red-100 text-red-800',
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className="mb-3 touch-none"
      role="button"
      aria-label={`${issue.title}, ${issue.priority} priority, ${issue.status} status. Press to drag and reorder.`}
      tabIndex={0}
    >
      <Link href={`/issues/${issue.id}`} scroll={false} className="block">
        <Card
          className={`transition-all duration-300 group relative overflow-hidden ${
            isHighlighted
              ? 'bg-yellow-50/50 shadow-[inset_0_0_0_4px_rgb(234,179,8),0_0_20px_rgba(234,179,8,0.6)] scale-[1.02]'
              : 'border-2 border-transparent hover:border-gray-200 hover:shadow-md'
          }`}
        >
          <CardHeader className="p-3 pb-0 flex flex-row items-start justify-between space-y-0">
            <span className="font-medium text-sm line-clamp-2">{issue.title}</span>
          </CardHeader>
          <CardContent className="p-3 pt-2">
            <div className="flex flex-wrap gap-1 mb-2">
              {issue.labels?.map((label) => (
                <Badge
                  key={label.id}
                  variant="outline"
                  style={{ borderColor: label.color }}
                  className="text-[10px] px-1 py-0 h-5"
                >
                  {label.name}
                </Badge>
              ))}
            </div>
            <div className="flex items-center justify-between mt-2">
              <Badge
                variant="secondary"
                className={`text-[10px] px-1 py-0 h-5 ${priorityColors[issue.priority] || ''}`}
              >
                {issue.priority}
              </Badge>
              {issue.assignee && (
                <Avatar className="h-6 w-6">
                  <AvatarImage
                    src={issue.assignee.avatar_url}
                    alt={`${issue.assignee.name}'s avatar`}
                  />
                  <AvatarFallback>{issue.assignee.name[0]}</AvatarFallback>
                </Avatar>
              )}
            </div>
          </CardContent>
        </Card>
      </Link>
    </div>
  );
}
