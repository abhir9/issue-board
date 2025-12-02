'use client';

import { Suspense } from 'react';
import dynamic from 'next/dynamic';
import { Board } from '@/components/board';
import { BoardFilters } from '@/components/board-filters';

const CreateIssueModal = dynamic(
  () => import('@/components/create-issue-modal').then((mod) => mod.CreateIssueModal),
  {
    ssr: false,
    loading: () => null,
  }
);

export default function IssuesPage() {
  return (
    <div className="h-screen flex flex-col">
      <header className="border-b px-4 md:px-6 py-3 flex flex-col md:flex-row items-start md:items-center justify-between bg-white shrink-0 gap-4">
        <div className="flex flex-col md:flex-row items-start md:items-center gap-4 w-full md:w-auto">
          <h1 className="font-semibold text-lg shrink-0">Issue Board</h1>
          <div className="w-full md:w-auto overflow-x-auto pb-1 md:pb-0">
            <Suspense fallback={<div>Loading filters...</div>}>
              <BoardFilters />
            </Suspense>
          </div>
        </div>
        <div className="self-end md:self-auto">
          <CreateIssueModal />
        </div>
      </header>
      <main className="flex-1 overflow-hidden bg-slate-100 h-full">
        <div className="h-full w-full overflow-x-auto overflow-y-hidden p-6">
          <Suspense fallback={<div>Loading board...</div>}>
            <Board />
          </Suspense>
        </div>
      </main>
    </div>
  );
}
