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
import { useQuery } from '@tanstack/react-query';
import { X } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';

export function BoardFilters() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const { data: users = [], isLoading: usersLoading } = useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
  });

  const { data: labels = [], isLoading: labelsLoading } = useQuery({
    queryKey: ['labels'],
    queryFn: getLabels,
  });

  const assignee = searchParams.get('assignee') || 'all';
  const priority = searchParams.get('priority') || 'all';
  const label = searchParams.get('labels') || 'all';

  const updateFilter = (key: string, value: string) => {
    const params = new URLSearchParams(searchParams.toString());
    if (value && value !== 'all') {
      params.set(key, value);
    } else {
      params.delete(key);
    }
    router.push(`?${params.toString()}`);
  };

  const clearFilters = () => {
    router.push('/issues');
  };

  const hasFilters = assignee !== 'all' || priority !== 'all' || label !== 'all';

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
