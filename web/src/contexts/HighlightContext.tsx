'use client';

import { createContext, useContext, useState, useCallback, useRef, useEffect } from 'react';

interface HighlightContextValue {
  highlightedId: string | null;
  triggerHighlight: (id: string) => void;
}

const HighlightContext = createContext<HighlightContextValue | undefined>(undefined);

const HIGHLIGHT_DURATION_MS = 2000;

export function HighlightProvider({ children }: { children: React.ReactNode }) {
  const [highlightedId, setHighlightedId] = useState<string | null>(null);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  const triggerHighlight = useCallback((id: string) => {
    // Clear any existing timeout
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    // Set the highlighted ID
    setHighlightedId(id);

    // Schedule removal of highlight
    timeoutRef.current = setTimeout(() => {
      setHighlightedId(null);
      timeoutRef.current = null;
    }, HIGHLIGHT_DURATION_MS);
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return (
    <HighlightContext.Provider value={{ highlightedId, triggerHighlight }}>
      {children}
    </HighlightContext.Provider>
  );
}

export function useHighlightContext() {
  const context = useContext(HighlightContext);
  if (context === undefined) {
    // Return default values instead of throwing error
    // This allows the hook to be used in portals like DragOverlay
    return { highlightedId: null, triggerHighlight: () => {} };
  }
  return context;
}
