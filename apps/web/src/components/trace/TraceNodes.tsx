import {
    AlertCircle,
    Clock,
    Cpu,
    Database,
    ExternalLink,
    Globe,
    MessageSquare,
    Play,
    Server,
    Square,
    Zap,
} from "lucide-react";
import { Handle, Position } from "react-flow-renderer";

// ─── Node Type Definitions ─────────────────────────────────────────────────

export interface TraceNodeData {
  id: string;
  label: string;
  type: string;
  status: string;
  serviceName: string;
  durationMs: number;
  startAt?: string;
  endAt?: string;
  httpMethod?: string;
  httpPath?: string;
  httpStatus?: number;
  dbSystem?: string;
  dbQuery?: string;
  dbRows?: number;
  metadata?: string;
  isRoot?: boolean;
  isTerminal?: boolean;
}

interface TraceNodeProps {
  data: TraceNodeData;
  selected?: boolean;
}

// ─── Style Helpers ──────────────────────────────────────────────────────────

const statusColors = {
  OK: { bg: "#10b981", light: "#d1fae5", text: "#065f46" },
  ERROR: { bg: "#ef4444", light: "#fee2e2", text: "#991b1b" },
};

function getStatusColors(status: string) {
  return (
    statusColors[status as keyof typeof statusColors] || {
      bg: "#6b7280",
      light: "#f3f4f6",
      text: "#374151",
    }
  );
}

function getDurationColor(ms: number): string {
  if (ms < 100) return "#10b981";
  if (ms < 500) return "#f59e0b";
  if (ms < 1000) return "#f97316";
  return "#ef4444";
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

function getMethodColor(method: string): string {
  const colors: Record<string, string> = {
    GET: "#3b82f6",
    POST: "#10b981",
    PUT: "#f59e0b",
    DELETE: "#ef4444",
    PATCH: "#8b5cf6",
  };
  return colors[method] || "#6b7280";
}

// ─── Node Type Registry ─────────────────────────────────────────────────────

function getNodeIcon(type: string) {
  const iconMap: Record<string, React.ElementType> = {
    HTTP: Globe,
    DB: Database,
    EXTERNAL_API: ExternalLink,
    CUSTOM: Cpu,
    QUEUE: MessageSquare,
    CACHE: Zap,
    INTERNAL: Server,
  };
  return iconMap[type] || Cpu;
}

function getNodeStyles(type: string, status: string) {
  const statusClr = getStatusColors(status);
  const isOk = status === "OK";

  const baseStyles: Record<string, React.CSSProperties> = {
    HTTP: {
      border: `2px solid ${isOk ? "#3b82f6" : statusClr.bg}`,
      background: `linear-gradient(135deg, #ffffff 0%, #f0f9ff 100%)`,
    },
    DB: {
      border: `2px solid ${isOk ? "#8b5cf6" : statusClr.bg}`,
      background: `linear-gradient(135deg, #ffffff 0%, #faf5ff 100%)`,
    },
    EXTERNAL_API: {
      border: `2px solid ${isOk ? "#f97316" : statusClr.bg}`,
      background: `linear-gradient(135deg, #ffffff 0%, #fff7ed 100%)`,
    },
    QUEUE: {
      border: `2px solid ${isOk ? "#06b6d4" : statusClr.bg}`,
      background: `linear-gradient(135deg, #ffffff 0%, #ecfeff 100%)`,
    },
    CACHE: {
      border: `2px solid ${isOk ? "#14b8a6" : statusClr.bg}`,
      background: `linear-gradient(135deg, #ffffff 0%, #f0fdfa 100%)`,
    },
    CUSTOM: {
      border: `2px solid ${isOk ? "#6366f1" : statusClr.bg}`,
      background: `linear-gradient(135deg, #ffffff 0%, #eef2ff 100%)`,
    },
    INTERNAL: {
      border: `2px solid ${isOk ? "#64748b" : statusClr.bg}`,
      background: `linear-gradient(135deg, #ffffff 0%, #f8fafc 100%)`,
    },
  };

  return baseStyles[type] || baseStyles.CUSTOM;
}

// ─── Custom Node Component ──────────────────────────────────────────────────

export function TraceNode({ data, selected }: TraceNodeProps) {
  const Icon = getNodeIcon(data.type);
  const nodeStyles = getNodeStyles(data.type, data.status);
  const statusClr = getStatusColors(data.status);
  const isOk = data.status === "OK";
  const durationColor = getDurationColor(data.durationMs);

  // Root/Start node
  if (data.isRoot) {
    return (
      <div
        style={{
          width: 180,
          padding: "12px 14px",
          borderRadius: 12,
          border: "2px solid #6366f1",
          background: "linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)",
          color: "#fff",
          boxShadow: "0 4px 12px rgba(99, 102, 241, 0.3)",
        }}
      >
        <Handle type="source" position={Position.Right} />
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <div
            style={{
              width: 28,
              height: 28,
              borderRadius: 8,
              background: "rgba(255,255,255,0.2)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <Play size={14} fill="#fff" />
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 11, fontWeight: 700, opacity: 0.9 }}>
              INÍCIO
            </div>
            <div style={{ fontSize: 9, opacity: 0.7, marginTop: 2 }}>
              {data.serviceName}
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Terminal/End node
  if (data.isTerminal) {
    return (
      <div
        style={{
          width: 140,
          padding: "10px 12px",
          borderRadius: 10,
          border: "2px solid #64748b",
          background: "#f8fafc",
        }}
      >
        <Handle type="target" position={Position.Left} />
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <Square size={12} fill="#64748b" color="#64748b" />
          <span style={{ fontSize: 10, fontWeight: 600, color: "#64748b" }}>
            FIM
          </span>
        </div>
      </div>
    );
  }

  // Regular nodes
  return (
    <div
      style={{
        width: 240,
        ...nodeStyles,
        borderRadius: 12,
        padding: 0,
        overflow: "hidden",
        boxShadow: selected
          ? "0 0 0 2px #6366f1, 0 8px 24px rgba(0,0,0,0.12)"
          : "0 2px 8px rgba(0,0,0,0.06)",
        transition: "box-shadow 0.2s",
      }}
    >
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />

      {/* Header */}
      <div
        style={{
          padding: "8px 12px",
          borderBottom: "1px solid #e2e8f0",
          display: "flex",
          alignItems: "center",
          gap: 8,
        }}
      >
        {/* Status indicator */}
        <div
          style={{
            width: 3,
            height: 24,
            borderRadius: 2,
            background: isOk ? "#10b981" : "#ef4444",
          }}
        />

        {/* Icon */}
        <div
          style={{
            width: 28,
            height: 28,
            borderRadius: 8,
            background: isOk ? `${statusClr.bg}15` : `${statusClr.bg}20`,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <Icon size={14} color={isOk ? statusClr.bg : statusClr.bg} />
        </div>

        {/* Label & Service */}
        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            style={{
              fontSize: 11,
              fontWeight: 600,
              color: "#0f172a",
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {data.label}
          </div>
          <div
            style={{
              fontSize: 9,
              color: "#64748b",
              marginTop: 1,
            }}
          >
            {data.serviceName}
          </div>
        </div>

        {/* Duration badge */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 3,
            padding: "2px 6px",
            borderRadius: 6,
            background: `${durationColor}12`,
            fontSize: 10,
            fontWeight: 600,
            color: durationColor,
          }}
        >
          <Clock size={9} />
          {formatDuration(data.durationMs)}
        </div>
      </div>

      {/* Content */}
      <div style={{ padding: "6px 12px 8px" }}>
        {/* HTTP info */}
        {data.httpMethod && data.httpPath && (
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 6,
              fontSize: 9,
            }}
          >
            <span
              style={{
                padding: "2px 5px",
                borderRadius: 4,
                background: getMethodColor(data.httpMethod),
                color: "#fff",
                fontWeight: 600,
                fontSize: 8,
              }}
            >
              {data.httpMethod}
            </span>
            <span
              style={{
                color: "#475569",
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
                flex: 1,
                fontFamily: "monospace",
              }}
            >
              {data.httpPath}
            </span>
            {data.httpStatus && (
              <span
                style={{
                  color: data.httpStatus < 400 ? "#10b981" : "#ef4444",
                  fontWeight: 600,
                }}
              >
                {data.httpStatus}
              </span>
            )}
          </div>
        )}

        {/* DB info */}
        {data.dbQuery && (
          <div style={{ fontSize: 9 }}>
            <div
              style={{
                color: "#64748b",
                marginBottom: 2,
                display: "flex",
                alignItems: "center",
                gap: 4,
              }}
            >
              <Database size={8} />
              {data.dbSystem || "database"}
            </div>
            <div
              style={{
                color: "#475569",
                fontFamily: "monospace",
                fontSize: 8,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
            >
              {data.dbQuery.substring(0, 60)}...
            </div>
            {data.dbRows !== undefined && (
              <div style={{ color: "#94a3b8", marginTop: 2 }}>
                {data.dbRows} row(s)
              </div>
            )}
          </div>
        )}

        {/* Error indicator */}
        {!isOk && (
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 4,
              marginTop: 4,
              padding: "4px 6px",
              borderRadius: 6,
              background: "#fee2e2",
              fontSize: 9,
              color: "#991b1b",
              fontWeight: 500,
            }}
          >
            <AlertCircle size={10} />
            Error detected
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Node Type Map for ReactFlow ────────────────────────────────────────────

export const nodeTypes = {
  traceNode: TraceNode,
};
