import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { FileQuestion } from 'lucide-react';

export default function IssueNotFound() {
  return (
    <div className="flex items-center justify-center min-h-screen p-4 bg-gray-50">
      <Card className="max-w-md w-full">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-gray-700">
            <FileQuestion className="h-6 w-6" />
            Issue Not Found
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-gray-600">
            The issue you&apos;re looking for doesn&apos;t exist or may have been deleted.
          </p>
          <Link href="/issues">
            <Button className="w-full">Back to Issue Board</Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}
