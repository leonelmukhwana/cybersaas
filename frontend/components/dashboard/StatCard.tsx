import type { ReactNode } from "react";

interface StatCardProps {
  title: string;
  value: string | number;
  description?: string;
  icon?: ReactNode;
  trend?: string;
  trendType?: "positive" | "negative" | "neutral";
}

export default function StatCard({
  title,
  value,
  description,
  icon,
  trend,
  trendType = "neutral",
}: StatCardProps) {
  const trendStyles = {
    positive: "text-emerald-600 bg-emerald-50",
    negative: "text-red-600 bg-red-50",
    neutral: "text-slate-600 bg-slate-100",
  };

  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:shadow-md">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="text-sm font-medium text-slate-500">
            {title}
          </p>

          <p className="mt-2 text-2xl font-bold tracking-tight text-slate-950">
            {value}
          </p>
        </div>

        {icon && (
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-[#0757B8]">
            {icon}
          </div>
        )}
      </div>

      {(description || trend) && (
        <div className="mt-4 flex flex-wrap items-center gap-2">
          {trend && (
            <span
              className={`rounded-md px-2 py-1 text-xs font-semibold ${trendStyles[trendType]}`}
            >
              {trend}
            </span>
          )}

          {description && (
            <span className="text-xs text-slate-500">
              {description}
            </span>
          )}
        </div>
      )}
    </div>
  );
}