export default function IssuesLoading() {
  return (
    <div className="p-6 space-y-4">
      <div className="h-10 w-64 rounded-lg bg-gray-200 animate-pulse" />
      <div className="flex gap-4 overflow-x-auto">
        {[1, 2, 3, 4, 5].map((col) => (
          <div
            key={col}
            className="w-80 shrink-0 rounded-lg border border-gray-200 bg-white shadow-sm"
          >
            <div className="h-10 border-b border-gray-100 bg-gray-50 rounded-t-lg" />
            <div className="p-3 space-y-3">
              {[1, 2, 3].map((card) => (
                <div key={card} className="h-24 rounded-md bg-gray-100 animate-pulse" />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
