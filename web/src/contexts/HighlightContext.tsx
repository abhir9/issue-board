'use client';

import { createContext, useContext } from 'react';

interface HighlightContextValue {
  highlightedId: string | null;
}

const HighlightContext = createContext<HighlightContextValue | undefined>(undefined);

export function HighlightProvider({
  children,
  highlightedId,
}: {
  children: React.ReactNode;
  highlightedId: string | null;
}) {
  return (
    <HighlightContext.Provider value={{ highlightedId }}>{children}</HighlightContext.Provider>
  );
}

export function useHighlightContext() {
  const context = useContext(HighlightContext);
  if (context === undefined) {
    throw new Error('useHighlightContext must be used within a HighlightProvider');
  }
  return context;
}
