import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';

export function IssueDetailsSkeleton() {
    return (
        <div className="flex flex-col h-full bg-gray-50">
            <header className="border-b px-6 py-3 flex items-center justify-between bg-white shrink-0">
                <div className="flex items-center gap-4">
                    <Skeleton className="h-9 w-9 rounded-md" />
                    <Skeleton className="h-6 w-32" />
                </div>
            </header>
            <main className="flex-1 p-6 w-full overflow-y-auto">
                <Card>
                    <CardHeader className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                        <div className="flex-1 w-full">
                            <Skeleton className="h-10 w-full max-w-2xl" />
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                            <Skeleton className="h-9 w-24" />
                            <Skeleton className="h-9 w-20" />
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-6">
                        {/* Status and Priority */}
                        <div className="grid grid-cols-2 gap-6">
                            <div className="space-y-2">
                                <Skeleton className="h-4 w-12" />
                                <Skeleton className="h-10 w-full" />
                            </div>
                            <div className="space-y-2">
                                <Skeleton className="h-4 w-12" />
                                <Skeleton className="h-10 w-full" />
                            </div>
                        </div>

                        {/* Assignee */}
                        <div className="space-y-2">
                            <Skeleton className="h-4 w-16" />
                            <Skeleton className="h-10 w-full" />
                        </div>

                        {/* Labels */}
                        <div className="space-y-2">
                            <Skeleton className="h-4 w-12" />
                            <Skeleton className="h-10 w-full" />
                        </div>

                        {/* Description */}
                        <div className="space-y-2">
                            <Skeleton className="h-4 w-20" />
                            <Skeleton className="h-[150px] w-full" />
                        </div>

                        {/* Activity */}
                        <div className="space-y-4 pt-4 border-t">
                            <Skeleton className="h-5 w-16" />
                            <div className="space-y-4">
                                <div className="flex gap-3">
                                    <Skeleton className="w-8 h-8 rounded-full shrink-0" />
                                    <div className="flex-1 space-y-2">
                                        <Skeleton className="h-4 w-48" />
                                        <Skeleton className="h-3 w-20" />
                                    </div>
                                </div>
                                <div className="flex gap-3">
                                    <Skeleton className="w-8 h-8 rounded-full shrink-0" />
                                    <div className="flex-1 space-y-2">
                                        <Skeleton className="h-4 w-56" />
                                        <Skeleton className="h-3 w-16" />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </main>
        </div>
    );
}
