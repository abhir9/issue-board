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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Check, ChevronsUpDown, Loader2 } from 'lucide-react';
import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createIssueSchema, type CreateIssueFormData } from '@/lib/schemas';

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

  const { data: users = [], error: usersError } = useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
  });

  const { data: labels = [], error: labelsError } = useQuery({
    queryKey: ['labels'],
    queryFn: getLabels,
  });

  // React Hook Form with Zod validation
  const form = useForm<CreateIssueFormData>({
    resolver: zodResolver(createIssueSchema),
    defaultValues: {
      title: '',
      description: '',
      status: 'Todo',
      priority: 'Medium',
      label_ids: [],
    },
  });

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors, isSubmitting },
  } = form;

  const createMutation = useMutation({
    mutationFn: createIssue,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues'] });
      setOpen(false);
      setTimeout(() => {
        reset();
      }, 300);
      toast.success('Issue created successfully');
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

  const onSubmit = (data: CreateIssueFormData) => {
    createMutation.mutate(data);
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
      <DialogContent className="sm:max-w-[550px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-semibold">Create New Issue</DialogTitle>
          <DialogDescription className="text-sm text-muted-foreground">
            Add a new issue to track your work. Fill in the details below.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5 py-4">
          {/* Title Field */}
          <div className="space-y-2">
            <Label htmlFor="title" className="text-sm font-medium">
              Title <span className="text-red-500">*</span>
            </Label>
            <Input 
              id="title" 
              {...register('title')} 
              placeholder="Enter issue title..."
              className="h-10"
            />
            {errors.title && <p className="text-xs text-red-500 mt-1">{errors.title.message}</p>}
          </div>

          {/* Description Field */}
          <div className="space-y-2">
            <Label htmlFor="description" className="text-sm font-medium">
              Description
            </Label>
            <Textarea 
              id="description" 
              {...register('description')} 
              placeholder="Add a detailed description..."
              className="min-h-[100px] resize-none"
            />
            {errors.description && (
              <p className="text-xs text-red-500 mt-1">{errors.description.message}</p>
            )}
          </div>

          {/* Status and Priority */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="status" className="text-sm font-medium">
                Status <span className="text-red-500">*</span>
              </Label>
              <Controller
                name="status"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="h-10">
                      <SelectValue placeholder="Select status" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Backlog">📋 Backlog</SelectItem>
                      <SelectItem value="Todo">📝 Todo</SelectItem>
                      <SelectItem value="In Progress">🔄 In Progress</SelectItem>
                      <SelectItem value="Done">✅ Done</SelectItem>
                      <SelectItem value="Canceled">❌ Canceled</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
              {errors.status && <p className="text-xs text-red-500 mt-1">{errors.status.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="priority" className="text-sm font-medium">
                Priority <span className="text-red-500">*</span>
              </Label>
              <Controller
                name="priority"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="h-10">
                      <SelectValue placeholder="Select priority" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Low">
                        <div className="flex items-center gap-2">
                          <span className="inline-block w-2 h-2 rounded-full bg-blue-500"></span>
                          Low
                        </div>
                      </SelectItem>
                      <SelectItem value="Medium">
                        <div className="flex items-center gap-2">
                          <span className="inline-block w-2 h-2 rounded-full bg-yellow-500"></span>
                          Medium
                        </div>
                      </SelectItem>
                      <SelectItem value="High">
                        <div className="flex items-center gap-2">
                          <span className="inline-block w-2 h-2 rounded-full bg-orange-500"></span>
                          High
                        </div>
                      </SelectItem>
                      <SelectItem value="Critical">
                        <div className="flex items-center gap-2">
                          <span className="inline-block w-2 h-2 rounded-full bg-red-500"></span>
                          Critical
                        </div>
                      </SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
              {errors.priority && <p className="text-xs text-red-500 mt-1">{errors.priority.message}</p>}
            </div>
          </div>

          {/* Assignee */}
          <div className="space-y-2">
            <Label htmlFor="assignee" className="text-sm font-medium">
              Assignee
            </Label>
            <Controller
              name="assignee_id"
              control={control}
              render={({ field }) => (
                <Select
                  value={field.value || 'unassigned'}
                  onValueChange={(value) =>
                    field.onChange(value === 'unassigned' ? undefined : value)
                  }
                >
                  <SelectTrigger className="h-10">
                    <SelectValue placeholder="Select assignee" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="unassigned">
                      <div className="flex items-center gap-2">
                        <div className="w-6 h-6 rounded-full bg-gray-200 flex items-center justify-center text-xs">
                          ?
                        </div>
                        Unassigned
                      </div>
                    </SelectItem>
                    {users.map((user) => (
                      <SelectItem key={user.id} value={user.id}>
                        <div className="flex items-center gap-2">
                          <div className="w-6 h-6 rounded-full bg-gradient-to-br from-blue-400 to-purple-500 flex items-center justify-center text-white text-xs font-medium">
                            {user.name[0]}
                          </div>
                          {user.name}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          </div>

          {/* Labels */}
          <div className="space-y-2">
            <Label className="text-sm font-medium">Labels</Label>
            <Controller
              name="label_ids"
              control={control}
              render={({ field }) => (
                <>
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        role="combobox"
                        type="button"
                        className={cn(
                          'w-full justify-between h-10',
                          !field.value?.length && 'text-muted-foreground'
                        )}
                      >
                        {field.value && field.value.length > 0
                          ? `${field.value.length} label${field.value.length > 1 ? 's' : ''} selected`
                          : 'Select labels...'}
                        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0">
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
                                    const currentLabels = field.value || [];
                                    const isSelected = currentLabels.includes(label.id);
                                    field.onChange(
                                      isSelected
                                        ? currentLabels.filter((id: string) => id !== label.id)
                                        : [...currentLabels, label.id]
                                    );
                                  }}
                                >
                                  <Check
                                    className={cn(
                                      'mr-2 h-4 w-4',
                                      field.value?.includes(label.id) ? 'opacity-100' : 'opacity-0'
                                    )}
                                  />
                                  <div className="flex items-center gap-2 flex-1">
                                    <div
                                      className="h-3 w-3 rounded-full shrink-0 ring-1 ring-offset-1 ring-black/10"
                                      style={{ backgroundColor: label.color }}
                                    />
                                    <span className="truncate">{label.name}</span>
                                  </div>
                                </CommandItem>
                              ))}
                            </CommandGroup>
                          </CommandList>
                        </Command>
                      </PopoverContent>
                    </Popover>
                    
                    {/* Selected Labels Display */}
                    {field.value && field.value.length > 0 && (
                      <div className="flex flex-wrap gap-1.5 pt-2">
                        {(field.value || []).map((labelId: string) => {
                          const label = labels.find((l) => l.id === labelId);
                          if (!label) return null;
                          return (
                            <Badge
                              key={label.id}
                              variant="outline"
                              className="gap-1.5 px-2 py-1 text-xs font-medium border"
                              style={{
                                backgroundColor: `${label.color}15`,
                                color: label.color,
                                borderColor: `${label.color}50`,
                              }}
                            >
                              <div
                                className="h-2 w-2 rounded-full"
                                style={{ backgroundColor: label.color }}
                              />
                              {label.name}
                              <button
                                type="button"
                                className="ml-0.5 hover:bg-black/10 rounded-full p-0.5 transition-colors"
                                onClick={() => {
                                  field.onChange(
                                    (field.value || []).filter((id: string) => id !== label.id)
                                  );
                                }}
                              >
                                <Plus className="h-2.5 w-2.5 rotate-45" />
                                <span className="sr-only">Remove</span>
                              </button>
                            </Badge>
                          );
                        })}
                      </div>
                    )}
                  </>
                )}
              />
          </div>

          {/* Submit Button */}
          <DialogFooter className="pt-4 border-t">
            <Button 
              type="submit" 
              disabled={isSubmitting || createMutation.isPending}
              className="w-full sm:w-auto min-w-[120px] h-10"
            >
              {isSubmitting || createMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Plus className="mr-2 h-4 w-4" />
                  Create Issue
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
