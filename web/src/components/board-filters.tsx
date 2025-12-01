'use client';

import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { getLabels, getUsers } from '@/lib/api';
import { ISSUE_STATUSES } from '@/constants/issues';
import { useQuery } from '@tanstack/react-query';
import { X } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useTransition } from 'react';
import { toast } from 'sonner';

export function BoardFilters() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [, startTransition] = useTransition();

  const {
    data: users = [],
    isLoading: usersLoading,
    error: usersError,
  } = useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
  });

  const {
    data: labels = [],
    isLoading: labelsLoading,
    error: labelsError,
  } = useQuery({
    queryKey: ['labels'],
    queryFn: getLabels,
  });

  useEffect(() => {
    if (usersError) {
      toast.error('Failed to load assignees');
    }
  }, [usersError]);

  useEffect(() => {
    if (labelsError) {
      toast.error('Failed to load labels');
    }
  }, [labelsError]);

  const assignee = searchParams.get('assignee') || 'all';
  const priority = searchParams.get('priority') || 'all';
  const status = searchParams.get('status') || 'all';
  const label = searchParams.get('labels') || 'all';

  const updateFilter = (key: string, value: string) => {
    const params = new URLSearchParams(searchParams.toString());
    if (value && value !== 'all') {
      params.set(key, value);
    } else {
      params.delete(key);
    }
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  };

  const clearFilters = () => {
    startTransition(() => {
      router.replace('/issues');
    });
  };

  const hasFilters =
    assignee !== 'all' || priority !== 'all' || status !== 'all' || label !== 'all';

  return (
    <div className="flex items-center gap-2">
      <Select value={assignee} onValueChange={(v) => updateFilter('assignee', v)}>
        <SelectTrigger className="w-[150px] h-9" aria-label="Filter by assignee">
          <SelectValue placeholder={usersLoading ? 'Loading...' : 'Assignee'} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Assignees</SelectItem>
          {usersLoading ? (
            <SelectItem value="loading" disabled>
              Loading...
            </SelectItem>
          ) : (
            users.map((user) => (
              <SelectItem key={user.id} value={user.id}>
                {user.name}
              </SelectItem>
            ))
          )}
        </SelectContent>
      </Select>

      <Select value={priority} onValueChange={(v) => updateFilter('priority', v)}>
        <SelectTrigger className="w-[150px] h-9" aria-label="Filter by priority">
          <SelectValue placeholder="Priority" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Priorities</SelectItem>
          <SelectItem value="Low">Low</SelectItem>
          <SelectItem value="Medium">Medium</SelectItem>
          <SelectItem value="High">High</SelectItem>
          <SelectItem value="Critical">Critical</SelectItem>
        </SelectContent>
      </Select>

      <Select value={status} onValueChange={(v) => updateFilter('status', v)}>
        <SelectTrigger className="w-[150px] h-9" aria-label="Filter by status">
          <SelectValue placeholder="Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Statuses</SelectItem>
          {ISSUE_STATUSES.map((s) => (
            <SelectItem key={s} value={s}>
              {s}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select value={label} onValueChange={(v) => updateFilter('labels', v)}>
        <SelectTrigger className="w-[150px] h-9" aria-label="Filter by label">
          <SelectValue placeholder={labelsLoading ? 'Loading...' : 'Label'} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Labels</SelectItem>
          {labelsLoading ? (
            <SelectItem value="loading" disabled>
              Loading...
            </SelectItem>
          ) : (
            labels.map((l) => (
              <SelectItem key={l.id} value={l.name}>
                {l.name}
              </SelectItem>
            ))
          )}
        </SelectContent>
      </Select>

      {hasFilters && (
        <Button variant="ghost" size="sm" onClick={clearFilters} className="h-9 px-2 text-gray-500">
          <X className="h-4 w-4 mr-1" /> Clear
        </Button>
      )}
    </div>
  );
}
