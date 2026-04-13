import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

const badgeVariants = cva(
  "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold transition-colors",
  {
    variants: {
      variant: {
        default:
          "bg-indigo-100 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-400",
        success:
          "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400",
        destructive:
          "bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400",
        warning:
          "bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400",
        info: "bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400",
        outline:
          "border border-slate-200 text-slate-700 dark:border-slate-800 dark:text-slate-300",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={badgeVariants({ variant, className })} {...props} />
  );
}

export { Badge, badgeVariants };
