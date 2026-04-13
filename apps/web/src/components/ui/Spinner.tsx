import { cva, type VariantProps } from "class-variance-authority";
import { Loader2 } from "lucide-react";

const spinnerVariants = cva("animate-spin", {
  variants: {
    size: {
      sm: "h-4 w-4",
      md: "h-6 w-6",
      lg: "h-8 w-8",
      xl: "h-12 w-12",
    },
  },
  defaultVariants: {
    size: "md",
  },
});

interface SpinnerProps extends VariantProps<typeof spinnerVariants> {
  className?: string;
  label?: string;
}

export function Spinner({ size, className, label }: SpinnerProps) {
  return (
    <div className="flex flex-col items-center gap-2">
      <Loader2 className={spinnerVariants({ size, className })} />
      {label && (
        <span className="text-sm text-slate-500 dark:text-slate-400">
          {label}
        </span>
      )}
    </div>
  );
}
