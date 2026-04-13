import * as React from "react";

function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={`animate-pulse rounded-md bg-slate-200 dark:bg-slate-800 ${className || ""}`}
      {...props}
    />
  );
}

function SkeletonLine({
  className,
  lines = 1,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & { lines?: number }) {
  return (
    <div className="space-y-2" {...props}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          className={`h-4 w-full ${className || ""}`}
        />
      ))}
    </div>
  );
}

function SkeletonCard({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={`space-y-3 rounded-xl border border-slate-200 p-6 dark:border-slate-800 ${className || ""}`} {...props}>
      <Skeleton className="h-5 w-1/3" />
      <SkeletonLine lines={2} />
    </div>
  );
}

function SkeletonTable({ rows = 5, className, ...props }: React.HTMLAttributes<HTMLDivElement> & { rows?: number }) {
  return (
    <div className={`space-y-3 ${className || ""}`} {...props}>
      <Skeleton className="h-10 w-full" />
      {Array.from({ length: rows }).map((_, i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
    </div>
  );
}

export { Skeleton, SkeletonLine, SkeletonCard, SkeletonTable };
