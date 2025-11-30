'use client';

import * as React from 'react';
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command';
import { useRouter } from 'next/navigation';
import { Plus, LayoutDashboard, Search } from 'lucide-react';
import dynamic from 'next/dynamic';

const CreateIssueModal = dynamic(
  () => import('./create-issue-modal').then((mod) => mod.CreateIssueModal),
  {
    ssr: false,
  }
);

export function CommandPalette() {
  const [open, setOpen] = React.useState(false);
  const router = useRouter();
  const [showCreateModal, setShowCreateModal] = React.useState(false);

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };

    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, []);

  const runCommand = React.useCallback((command: () => unknown) => {
    setOpen(false);
    command();
  }, []);

  return (
    <>
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Type a command or search..." />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
          <CommandGroup heading="Suggestions">
            <CommandItem
              onSelect={() => {
                runCommand(() => setShowCreateModal(true));
              }}
            >
              <Plus className="mr-2 h-4 w-4" />
              <span>Create New Issue</span>
            </CommandItem>
            <CommandItem
              onSelect={() => {
                runCommand(() => router.push('/issues'));
              }}
            >
              <LayoutDashboard className="mr-2 h-4 w-4" />
              <span>Go to Board</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Filters">
            <CommandItem
              onSelect={() => {
                runCommand(() => router.push('/issues?priority=High'));
              }}
            >
              <Search className="mr-2 h-4 w-4" />
              <span>Filter by High Priority</span>
            </CommandItem>
            <CommandItem
              onSelect={() => {
                runCommand(() => router.push('/issues?status=Todo'));
              }}
            >
              <Search className="mr-2 h-4 w-4" />
              <span>Filter by Todo</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </CommandDialog>

      <CreateIssueModal
        open={showCreateModal}
        onOpenChange={setShowCreateModal}
        showTrigger={false}
      />
    </>
  );
}
