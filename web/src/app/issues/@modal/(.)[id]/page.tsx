'use client';

import IssueDetailsPage from '@/app/issues/[id]/page';
import { Modal } from '@/components/modal';
import { useRouter } from 'next/navigation';

export default function IssueModalPage() {
  const router = useRouter();

  const handleClose = () => {
    router.back();
  };

  return (
    <Modal>
      <IssueDetailsPage onNavigateBack={handleClose} />
    </Modal>
  );
}
