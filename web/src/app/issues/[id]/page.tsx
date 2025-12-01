'use client';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { deleteIssue, updateIssue, getUsers, getLabels, getIssue } from '@/lib/api';
import { Issue, IssuePriority, IssueStatus, Label, User } from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Save, Trash2, Check, ChevronsUpDown } from 'lucide-react';
import { useParams, useRouter } from 'next/navigation';
import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { cn } from '@/lib/utils';
import { IssueDetailsSkeleton } from '@/components/issue-details-skeleton';

// Form data type that includes both Issue properties and label_ids for updates
type IssueFormData = Partial<Issue> & {
  label_ids?: string[];
};

interface IssueDetailsPageProps {
  onNavigateBack?: () => void;
}

export default function IssueDetailsPage({ onNavigateBack }: IssueDetailsPageProps = {}) {
  const params = useParams();
  const id = params.id as string;
  const router = useRouter();

  const handleNavigateBack = () => {
    if (onNavigateBack) {
      onNavigateBack();
    } else {
      router.push('/issues');
    }
  };

  // Fetch issue directly or from cache
  const {
    data: issue,
    isLoading,
    error: issueError,
  } = useQuery({
    queryKey: ['issue', id],
    queryFn: () => getIssue(id),
  });

  const { data: users = [], error: usersError } = useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
  });
  const { data: labels = [], error: labelsError } = useQuery({
    queryKey: ['labels'],
    queryFn: getLabels,
  });

  useEffect(() => {
    if (issueError) {
      toast.error('Failed to load issue');
    }
  }, [issueError]);

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

  if (isLoading) return <IssueDetailsSkeleton />;
  if (!issue) return <div className="p-8">Issue not found</div>;
  return (
    <IssueDetailsContent
      key={issue.id}
      issue={issue}
      users={users}
      labels={labels}
      onNavigateBack={handleNavigateBack}
    />
  );
}

interface IssueDetailsContentProps {
  issue: Issue;
  users: User[];
  labels: Label[];
  onNavigateBack: () => void;
}

function IssueDetailsContent({ issue, users, labels, onNavigateBack }: IssueDetailsContentProps) {
  const queryClient = useQueryClient();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [formData, setFormData] = useState<IssueFormData>(() => ({
    ...issue,
    label_ids: issue.labels?.map((l) => l.id) || [],
  }));

  const updateMutation = useMutation({
    mutationFn: (updates: IssueFormData) =>
      updateIssue(issue.id, {
        title: updates.title,
        description: updates.description,
        status: updates.status,
        priority: updates.priority,
        assignee_id: updates.assignee_id,
        label_ids: updates.label_ids,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues'] });
      queryClient.invalidateQueries({ queryKey: ['issue', issue.id] });
      toast.success('Issue updated successfully');
      onNavigateBack();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteIssue(issue.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues'] });
      toast.success('Issue deleted successfully');
      onNavigateBack();
    },
  });

  const handleSave = () => {
    updateMutation.mutate(formData);
  };

  const handleDelete = () => {
    deleteMutation.mutate();
    setShowDeleteDialog(false);
  };

  return (
    <>
      <div className="flex flex-col h-full bg-gray-50">
        <header className="border-b px-6 py-3 flex items-center justify-between bg-white shrink-0">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" onClick={onNavigateBack} aria-label="Go back">
              <ArrowLeft className="h-4 w-4" />
            </Button>
            <h1 className="font-semibold text-lg">Issue Details</h1>
          </div>
        </header>
        <main className="flex-1 p-6 w-full overflow-y-auto">
          <Card>
            <CardHeader className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              <CardTitle className="text-xl flex-1 min-w-0 w-full">
                <Input
                  value={formData.title ?? issue.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  className="text-xl font-bold border-none shadow-none focus-visible:ring-0 px-0 w-full"
                />
              </CardTitle>
              <div className="flex items-center gap-2 shrink-0 self-end sm:self-auto">
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => setShowDeleteDialog(true)}
                  disabled={deleteMutation.isPending}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete
                </Button>
                <Button onClick={handleSave} disabled={updateMutation.isPending} size="sm">
                  <Save className="mr-2 h-4 w-4" />
                  {updateMutation.isPending ? 'Saving...' : 'Save'}
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <label className="text-sm font-medium">Status</label>
                  <Select
                    value={formData.status ?? issue.status}
                    onValueChange={(v) => setFormData({ ...formData, status: v as IssueStatus })}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {['Backlog', 'Todo', 'In Progress', 'Done', 'Canceled'].map((s) => (
                        <SelectItem key={s} value={s}>
                          {s}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Priority</label>
                  <Select
                    value={formData.priority ?? issue.priority}
                    onValueChange={(v) =>
                      setFormData({ ...formData, priority: v as IssuePriority })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {['Low', 'Medium', 'High', 'Critical'].map((p) => (
                        <SelectItem key={p} value={p}>
                          {p}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Assignee</label>
                <Select
                  value={(formData.assignee_id ?? issue.assignee_id) || 'unassigned'}
                  onValueChange={(v) =>
                    setFormData({ ...formData, assignee_id: v === 'unassigned' ? undefined : v })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Unassigned" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="unassigned">Unassigned</SelectItem>
                    {users.map((u) => (
                      <SelectItem key={u.id} value={u.id}>
                        {u.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Labels</label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      role="combobox"
                      className="justify-between h-auto min-h-10 py-2 w-full"
                    >
                      {formData.label_ids && formData.label_ids.length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                          {(formData.label_ids || []).map((labelId) => {
                            const label = labels.find((l) => l.id === labelId);
                            return label ? (
                              <Badge key={label.id} variant="secondary" className="mr-1">
                                {label.name}
                              </Badge>
                            ) : null;
                          })}
                        </div>
                      ) : (
                        'Select labels'
                      )}
                      <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-[400px] p-0">
                    <Command>
                      <CommandInput placeholder="Search labels..." />
                      <CommandList>
                        <CommandEmpty>No label found.</CommandEmpty>
                        <CommandGroup>
                          {labels.map((label) => (
                            <CommandItem
                              key={label.id}
                              value={label.name}
                              onSelect={() => {
                                const currentLabels = formData.label_ids || [];
                                const newLabels = currentLabels.includes(label.id)
                                  ? currentLabels.filter((id) => id !== label.id)
                                  : [...currentLabels, label.id];
                                setFormData({ ...formData, label_ids: newLabels });
                              }}
                            >
                              <Check
                                className={cn(
                                  'mr-2 h-4 w-4',
                                  formData.label_ids?.includes(label.id)
                                    ? 'opacity-100'
                                    : 'opacity-0'
                                )}
                              />
                              {label.name}
                            </CommandItem>
                          ))}
                        </CommandGroup>
                      </CommandList>
                    </Command>
                  </PopoverContent>
                </Popover>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Description</label>
                <Textarea
                  value={formData.description ?? issue.description ?? ''}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="min-h-[150px]"
                />
              </div>

              <div className="space-y-4 pt-4 border-t">
                <h3 className="font-semibold text-sm">Activity</h3>
                <div className="space-y-4">
                  <div className="flex gap-3 text-sm">
                    <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-medium shrink-0">
                      {issue.assignee ? issue.assignee.name[0] : 'U'}
                    </div>
                    <div>
                      <p>
                        <span className="font-medium">
                          {issue.assignee ? issue.assignee.name : 'User'}
                        </span>{' '}
                        created this issue
                      </p>
                      <p className="text-gray-500 text-xs mt-0.5">
                        {new Date(issue.created_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-3 text-sm">
                    <div className="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center text-gray-600 font-medium shrink-0">
                      S
                    </div>
                    <div>
                      <p>
                        <span className="font-medium">System</span> changed status to{' '}
                        <span className="font-medium">{issue.status}</span>
                      </p>
                      <p className="text-gray-500 text-xs mt-0.5">
                        {new Date(issue.updated_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </main>
      </div>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Issue</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{issue.title}&quot;? This action cannot be
              undone and will permanently remove this issue from the board.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete Issue'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
