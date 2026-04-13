import { AnimatePresence, motion } from "framer-motion";
import {
    AlertCircle,
    Calendar,
    Clock,
    Database,
    ExternalLink,
    Globe,
    Server,
    X,
} from "lucide-react";

interface NodeDetailsProps {
  data: any;
  onClose: () => void;
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return "N/A";
  return new Date(dateStr).toLocaleString("pt-BR");
}

function getDurationBadge(ms: number) {
  if (ms < 100) return { bg: "#d1fae5", text: "#065f46", label: "Fast" };
  if (ms < 500) return { bg: "#fef3c7", text: "#92400e", label: "Normal" };
  if (ms < 1000) return { bg: "#fed7aa", text: "#9a3412", label: "Slow" };
  return { bg: "#fee2e2", text: "#991b1b", label: "Critical" };
}

export function NodeDetails({ data, onClose }: NodeDetailsProps) {
  const d = data?.data || {};
  const durationBadge = getDurationBadge(d.durationMs || 0);

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/20 backdrop-blur-sm z-40"
        onClick={onClose}
      />
      <AnimatePresence>
        <motion.div
          initial={{ x: 400, opacity: 0 }}
          animate={{ x: 0, opacity: 1 }}
          exit={{ x: 400, opacity: 0 }}
          transition={{ type: "spring", stiffness: 300, damping: 30 }}
          className="fixed right-0 top-0 bottom-0 w-[420px] bg-white dark:bg-slate-900 border-l border-slate-200 dark:border-slate-800 shadow-2xl z-50 overflow-y-auto"
        >
          {/* Header */}
          <div className="sticky top-0 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 px-6 py-4 flex items-center justify-between z-10">
            <div className="flex items-center gap-3">
              <div
                className="w-10 h-10 rounded-xl flex items-center justify-center"
                style={{
                  background: d.status === "ERROR" ? "#fee2e2" : "#eef2ff",
                }}
              >
                {d.type === "HTTP" && <Globe size={20} color="#3b82f6" />}
                {d.type === "DB" && <Database size={20} color="#8b5cf6" />}
                {d.type === "EXTERNAL_API" && (
                  <ExternalLink size={20} color="#f97316" />
                )}
                {d.type === "CUSTOM" && <Server size={20} color="#6366f1" />}
                {!["HTTP", "DB", "EXTERNAL_API", "CUSTOM"].includes(d.type) && (
                  <Server size={20} color="#64748b" />
                )}
              </div>
              <div>
                <h3 className="font-semibold text-slate-900 dark:text-white text-sm">
                  {d.label}
                </h3>
                <p className="text-xs text-slate-500 dark:text-slate-400">
                  {d.serviceName}
                </p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              <X size={18} className="text-slate-500" />
            </button>
          </div>

          {/* Content */}
          <div className="p-6 space-y-6">
            {/* Status & Duration */}
            <div className="grid grid-cols-2 gap-3">
              <div className="bg-slate-50 dark:bg-slate-800 rounded-xl p-4">
                <div className="text-xs text-slate-500 dark:text-slate-400 mb-1">
                  Status
                </div>
                <div className="flex items-center gap-2">
                  <div
                    className="w-2.5 h-2.5 rounded-full"
                    style={{
                      background: d.status === "ERROR" ? "#ef4444" : "#10b981",
                    }}
                  />
                  <span
                    className={`text-sm font-semibold ${
                      d.status === "ERROR"
                        ? "text-red-600 dark:text-red-400"
                        : "text-emerald-600 dark:text-emerald-400"
                    }`}
                  >
                    {d.status}
                  </span>
                </div>
              </div>

              <div className="bg-slate-50 dark:bg-slate-800 rounded-xl p-4">
                <div className="text-xs text-slate-500 dark:text-slate-400 mb-1">
                  Duration
                </div>
                <div className="flex items-center gap-2">
                  <Clock size={14} className="text-slate-400" />
                  <span className="text-sm font-bold text-slate-900 dark:text-white">
                    {formatDuration(d.durationMs)}
                  </span>
                </div>
                <div
                  className="text-[10px] font-medium mt-1 px-1.5 py-0.5 rounded inline-block"
                  style={{
                    background: durationBadge.bg,
                    color: durationBadge.text,
                  }}
                >
                  {durationBadge.label}
                </div>
              </div>
            </div>

            {/* Timestamps */}
            <div className="bg-slate-50 dark:bg-slate-800 rounded-xl p-4 space-y-2">
              <div className="text-xs font-medium text-slate-700 dark:text-slate-300 mb-2">
                Timestamps
              </div>
              <div className="flex items-start gap-2">
                <Calendar size={12} className="text-slate-400 mt-0.5" />
                <div className="flex-1">
                  <div className="text-[10px] text-slate-500 dark:text-slate-400">
                    Start
                  </div>
                  <div className="text-xs text-slate-900 dark:text-white font-mono">
                    {formatDate(d.startAt)}
                  </div>
                </div>
              </div>
              <div className="flex items-start gap-2">
                <Calendar size={12} className="text-slate-400 mt-0.5" />
                <div className="flex-1">
                  <div className="text-[10px] text-slate-500 dark:text-slate-400">
                    End
                  </div>
                  <div className="text-xs text-slate-900 dark:text-white font-mono">
                    {formatDate(d.endAt)}
                  </div>
                </div>
              </div>
            </div>

            {/* HTTP Details */}
            {d.httpMethod && d.httpPath && (
              <div className="bg-slate-50 dark:bg-slate-800 rounded-xl p-4 space-y-3">
                <div className="text-xs font-medium text-slate-700 dark:text-slate-300 flex items-center gap-2">
                  <Globe size={12} />
                  HTTP Request
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className="px-2 py-1 rounded text-[10px] font-bold text-white"
                    style={{
                      background: (() => {
                        const methodColors: Record<string, string> = {
                          GET: "#3b82f6",
                          POST: "#10b981",
                          PUT: "#f59e0b",
                          DELETE: "#ef4444",
                          PATCH: "#8b5cf6",
                        };
                        return methodColors[d.httpMethod] || "#6b7280";
                      })(),
                    }}
                  >
                    {d.httpMethod}
                  </span>
                  <span className="text-xs font-mono text-slate-900 dark:text-white flex-1 truncate">
                    {d.httpPath}
                  </span>
                </div>
                {d.httpStatus && (
                  <div className="flex items-center gap-2">
                    <span className="text-[10px] text-slate-500 dark:text-slate-400">
                      Status:
                    </span>
                    <span
                      className={`text-sm font-bold ${
                        d.httpStatus < 400
                          ? "text-emerald-600 dark:text-emerald-400"
                          : "text-red-600 dark:text-red-400"
                      }`}
                    >
                      {d.httpStatus}
                    </span>
                  </div>
                )}
              </div>
            )}

            {/* Database Details */}
            {d.dbQuery && (
              <div className="bg-slate-50 dark:bg-slate-800 rounded-xl p-4 space-y-3">
                <div className="text-xs font-medium text-slate-700 dark:text-slate-300 flex items-center gap-2">
                  <Database size={12} />
                  Database Query
                </div>
                {d.dbSystem && (
                  <div className="text-[10px] text-slate-500 dark:text-slate-400">
                    System:{" "}
                    <span className="text-slate-900 dark:text-white font-medium">
                      {d.dbSystem}
                    </span>
                  </div>
                )}
                <div className="bg-white dark:bg-slate-950 rounded-lg p-3 border border-slate-200 dark:border-slate-700">
                  <pre className="text-[10px] font-mono text-slate-700 dark:text-slate-300 whitespace-pre-wrap break-all max-h-32 overflow-y-auto">
                    {d.dbQuery}
                  </pre>
                </div>
                {d.dbRows !== undefined && (
                  <div className="text-[10px] text-slate-500 dark:text-slate-400">
                    Rows affected:{" "}
                    <span className="text-slate-900 dark:text-white font-medium">
                      {d.dbRows}
                    </span>
                  </div>
                )}
              </div>
            )}

            {/* Error Details */}
            {d.status === "ERROR" && (
              <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-900/50 rounded-xl p-4 space-y-2">
                <div className="flex items-center gap-2 text-red-700 dark:text-red-400">
                  <AlertCircle size={14} />
                  <span className="text-xs font-semibold">Error Detected</span>
                </div>
                <p className="text-[10px] text-red-600 dark:text-red-400/80">
                  This node encountered an error during execution. Check logs
                  for more details.
                </p>
              </div>
            )}

            {/* Metadata */}
            {d.metadata && (
              <div className="bg-slate-50 dark:bg-slate-800 rounded-xl p-4">
                <div className="text-xs font-medium text-slate-700 dark:text-slate-300 mb-2">
                  Metadata
                </div>
                <pre className="text-[10px] font-mono text-slate-600 dark:text-slate-400 whitespace-pre-wrap break-all max-h-48 overflow-y-auto">
                  {typeof d.metadata === "string"
                    ? d.metadata
                    : JSON.stringify(d.metadata, null, 2)}
                </pre>
              </div>
            )}
          </div>
        </motion.div>
      </AnimatePresence>
    </AnimatePresence>
  );
}
