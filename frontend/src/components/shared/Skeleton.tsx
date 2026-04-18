import clsx from 'clsx';

interface SkeletonProps {
  className?: string;
  rows?: number;
  height?: string;
}

export default function Skeleton({ className, rows = 1, height = 'h-4' }: SkeletonProps) {
  return (
    <div className={clsx('space-y-3', className)}>
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className={clsx('rounded bg-zinc-800 skeleton-pulse', height, i === rows - 1 && rows > 1 && 'w-3/4')}
        />
      ))}
    </div>
  );
}
