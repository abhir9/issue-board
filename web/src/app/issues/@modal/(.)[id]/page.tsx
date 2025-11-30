import IssueDetailsPage from '@/app/issues/[id]/page';
import { Modal } from '@/components/modal';

export default function IssueModalPage() {
  return (
    <Modal>
      <IssueDetailsPage />
    </Modal>
  );
}
