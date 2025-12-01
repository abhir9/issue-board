'use client';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
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
import { createIssue, getUsers, getLabels } from '@/lib/api';
import { CreateIssueRequest, IssuePriority, IssueStatus } from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Check, ChevronsUpDown } from 'lucide-react';
import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import { useRouter } from 'next/navigation';

interface CreateIssueModalProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  showTrigger?: boolean;
}

export function CreateIssueModal({
  open: controlledOpen,
  onOpenChange: setControlledOpen,
  showTrigger = true,
}: CreateIssueModalProps = {}) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isControlled = controlledOpen !== undefined && setControlledOpen !== undefined;
  const open = isControlled ? controlledOpen : internalOpen;
  const setOpen = isControlled ? setControlledOpen : setInternalOpen;

  const queryClient = useQueryClient();
  const router = useRouter();

  const { data: users = [], error: usersError } = useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
  });

  const { data: labels = [], error: labelsError } = useQuery({
    queryKey: ['labels'],
    queryFn: getLabels,
  });

  const [formData, setFormData] = useState<CreateIssueRequest>({
    title: '',
    description: '',
    status: 'Todo',
    priority: 'Medium',
    label_ids: [],
  });

  const createMutation = useMutation({
    mutationFn: createIssue,
    onSuccess: (newIssue) => {
      queryClient.invalidateQueries({ queryKey: ['issues'] });
      setOpen(false);
      setTimeout(() => {
        setFormData({
          title: '',
          description: '',
          status: 'Todo',
          priority: 'Medium',
          label_ids: [],
        });
      }, 300);
      toast.success('Issue created successfully');
      // Add highlight param to URL to trigger highlight animation on the board
      router.push(`/issues?highlight=${newIssue.id}`);
    },
    onError: () => {
      toast.error('Failed to create issue');
    },
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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Check if user is not typing in an input
      if (
        document.activeElement?.tagName === 'INPUT' ||
        document.activeElement?.tagName === 'TEXTAREA' ||
        document.activeElement?.getAttribute('contenteditable') === 'true'
      ) {
        return;
      }

      if (e.key.toLowerCase() === 'n') {
        e.preventDefault();
        setOpen(true);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [setOpen]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {showTrigger && (
        <DialogTrigger asChild>
          <Button>
            <Plus className="mr-2 h-4 w-4" /> New Issue
          </Button>
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create Issue</DialogTitle>
          <DialogDescription>
            Add a new issue to the board. Fill in the details below.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="title">Title</Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              required
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label htmlFor="status">Status</Label>
              <Select
                value={formData.status}
                onValueChange={(value) =>
                  setFormData({ ...formData, status: value as IssueStatus })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Backlog">Backlog</SelectItem>
                  <SelectItem value="Todo">Todo</SelectItem>
                  <SelectItem value="In Progress">In Progress</SelectItem>
                  <SelectItem value="Done">Done</SelectItem>
                  <SelectItem value="Canceled">Canceled</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="priority">Priority</Label>
              <Select
                value={formData.priority}
                onValueChange={(value) =>
                  setFormData({ ...formData, priority: value as IssuePriority })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select priority" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Low">Low</SelectItem>
                  <SelectItem value="Medium">Medium</SelectItem>
                  <SelectItem value="High">High</SelectItem>
                  <SelectItem value="Critical">Critical</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label htmlFor="assignee">Assignee</Label>
              <Select
                value={formData.assignee_id || 'unassigned'}
                onValueChange={(value) =>
                  setFormData({
                    ...formData,
                    assignee_id: value === 'unassigned' ? undefined : value,
                  })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select assignee" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="unassigned">Unassigned</SelectItem>
                  {users.map((user) => (
                    <SelectItem key={user.id} value={user.id}>
                      {user.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label>Labels</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    role="combobox"
                    className={cn(
                      'w-full justify-between',
                      !formData.label_ids?.length && 'text-muted-foreground'
                    )}
                  >
                    {formData.label_ids && formData.label_ids.length > 0
                      ? `${formData.label_ids.length} selected`
                      : 'Select labels'}
                    <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-full p-0">
                  <Command>
                    <CommandInput placeholder="Search label..." />
                    <CommandList>
                      <CommandEmpty>No label found.</CommandEmpty>
                      <CommandGroup>
                        {labels.map((label) => (
                          <CommandItem
                            value={label.name}
                            key={label.id}
                            onSelect={() => {
                              const currentLabels = formData.label_ids || [];
                              const isSelected = currentLabels.includes(label.id);
                              setFormData({
                                ...formData,
                                label_ids: isSelected
                                  ? currentLabels.filter((id) => id !== label.id)
                                  : [...currentLabels, label.id],
                              });
                            }}
                          >
                            <Check
                              className={cn(
                                'mr-2 h-4 w-4',
                                formData.label_ids?.includes(label.id) ? 'opacity-100' : 'opacity-0'
                              )}
                            />
                            <div className="flex items-center gap-2">
                              <div
                                className="h-3 w-3 rounded-full"
                                style={{ backgroundColor: label.color }}
                              />
                              {label.name}
                            </div>
                          </CommandItem>
                        ))}
                      </CommandGroup>
                    </CommandList>
                  </Command>
                </PopoverContent>
              </Popover>
              <div className="flex flex-wrap gap-2 mt-2">
                {(formData.label_ids || []).map((labelId) => {
                  const label = labels.find((l) => l.id === labelId);
                  if (!label) return null;
                  return (
                    <Badge
                      key={label.id}
                      variant="secondary"
                      className="gap-1"
                      style={{
                        backgroundColor: `${label.color}20`,
                        color: label.color,
                        borderColor: `${label.color}40`,
                      }}
                    >
                      {label.name}
                      <button
                        type="button"
                        className="ml-1 ring-offset-background rounded-full outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                        onClick={() => {
                          setFormData({
                            ...formData,
                            label_ids: (formData.label_ids || []).filter((id) => id !== label.id),
                          });
                        }}
                      >
                        <Plus className="h-3 w-3 rotate-45" />
                        <span className="sr-only">Remove</span>
                      </button>
                    </Badge>
                  );
                })}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Creating...' : 'Create Issue'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
