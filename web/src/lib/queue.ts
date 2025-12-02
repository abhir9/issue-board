let chain: Promise<unknown> = Promise.resolve();
let pending = 0;
let dragging = false;

export function getPendingCount() {
  return pending;
}

export function setDragging(value: boolean) {
  dragging = value;
}

export function isDragging() {
  return dragging;
}

export function enqueue<T>(task: () => Promise<T> | T): Promise<T> {
  pending += 1;
  const next = chain
    .then(() => task())
    .finally(() => {
      pending -= 1;
    });
  // Ensure subsequent enqueued tasks wait for this one
  chain = next.catch(() => {});
  return next as Promise<T>;
}

export function waitForIdle(): Promise<void> {
  // Resolves when the current chain completes (i.e., queue drains)
  return chain.then(() => {});
}

export function isFullyIdle(): boolean {
  return pending === 0 && !dragging;
}
