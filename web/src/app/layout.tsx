import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import Providers from '@/components/providers';
import { Toaster } from '@/components/ui/sonner';
import { CommandPalette } from '@/components/command-palette';
import { ErrorBoundary } from '@/components/error-boundary';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Issue Board',
  description: 'Kanban Issue Tracker',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <ErrorBoundary>
          <Providers>
            {children}
            <Toaster position="top-right" richColors theme="dark" />
            <CommandPalette />
          </Providers>
        </ErrorBoundary>
      </body>
    </html>
  );
}
