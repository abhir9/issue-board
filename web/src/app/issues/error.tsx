'use client';

import { useEffect } from 'react';
import { Button } from '@/components/ui/button';

interface IssuesErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function IssuesError({ error, reset }: IssuesErrorProps) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="flex flex-col items-center justify-center h-full gap-4 p-6">
      <p className="text-lg font-semibold">Something went wrong while loading the board.</p>
      <p className="text-sm text-muted-foreground">{error.message}</p>
      <div className="flex gap-2">
        <Button variant="outline" onClick={() => reset()}>
          Try Again
        </Button>
        <Button onClick={() => window.location.reload()}>Reload Page</Button>
      </div>
    </div>
  );
}
