'use client';

import { Dialog, DialogContent, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { useRouter } from 'next/navigation';

export function Modal({ children }: { children: React.ReactNode }) {
  const router = useRouter();

  const onDismiss = () => {
    router.back();
  };

  return (
    <Dialog open={true} onOpenChange={(open) => !open && onDismiss()}>
      <DialogContent
        className="max-w-3xl h-[90vh] overflow-y-auto p-0 border-none bg-transparent shadow-none"
        aria-describedby="modal-description"
      >
        <DialogTitle className="sr-only">Issue Details</DialogTitle>
        <DialogDescription id="modal-description" className="sr-only">
          Details of the selected issue
        </DialogDescription>
        <div className="bg-white rounded-lg h-full overflow-hidden flex flex-col">{children}</div>
      </DialogContent>
    </Dialog>
  );
}
