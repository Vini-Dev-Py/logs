import dagre from "@dagrejs/dagre";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import ReactFlow, {
    Background,
    Controls,
    MarkerType,
    MiniMap,
} from "react-flow-renderer";
import { useParams } from "react-router-dom";
import { api } from "../api/client";
import { NodeDetails } from "../components/trace/NodeDetails";
import { nodeTypes } from "../components/trace/TraceNodes";

function getLayoutedElements(nodes: any[], edges: any[], direction = "LR") {
  const g = new dagre.graphlib.Graph();
  g.setGraph({ rankdir: direction, nodesep: 80, ranksep: 120 });
  g.setDefaultEdgeLabel(() => ({}));

  nodes.forEach((node) => {
    g.setNode(node.id, { width: 240, height: 100 });
  });

  edges.forEach((edge) => {
    g.setEdge(edge.source, edge.target);
  });

  dagre.layout(g);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = g.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - 120,
        y: nodeWithPosition.y - 50,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
}

export function TraceViewerPage() {
  const { traceId = "" } = useParams();
  const [note, setNote] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedNode, setSelectedNode] = useState<any>(null);
  const queryClient = useQueryClient();
  const isMountedRef = useRef(true);

  const q = useQuery({
    queryKey: ["trace", traceId],
    queryFn: async () => (await api.get(`/traces/${traceId}`)).data,
  });

  const searchQ = useQuery({
    queryKey: ["search", searchQuery],
    queryFn: async () => {
      if (!searchQuery) return null;
      const params = new URLSearchParams({ query: searchQuery });
      return (await api.get(`/search?${params.toString()}`)).data.items || [];
    },
    enabled: searchQuery.length > 2,
  });

  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
      // React Query gerencia cache automaticamente - não cancelar queries no cleanup
    };
  }, []);

  const { nodes, edges } = useMemo(() => {
    const matchedNodeIds = new Set(
      (searchQ.data || [])
        .filter((r: any) => r.traceId === traceId)
        .map((r: any) => r.nodeId),
    );
    const isSearching = searchQuery.length > 2;

    const nodesData = q.data?.nodes || [];
    const edgesData = q.data?.edges || [];

    // Build nodes with unique IDs and find root node
    const hasParent = new Set<string>();
    edgesData.forEach((e: any) => {
      if (e.source) hasParent.add(e.target);
    });

    const baseNodes = nodesData.map((n: any, i: number) => {
      let isMatch = false;
      if (isSearching && matchedNodeIds.has(n.id)) {
        isMatch = true;
      }

      const isRoot = !hasParent.has(n.id);
      const isTerminal = n.data?.status === "TERMINAL";

      let borderStyle =
        n.data?.status === "ERROR"
          ? "2px solid #dc2626"
          : n.data?.durationMs > 1000
            ? "2px solid #f59e0b"
            : "1px solid #cbd5e1";

      let opacity = 1;
      let boxShadow = "none";

      if (isSearching) {
        if (isMatch) {
          borderStyle = "3px solid #6366f1";
          boxShadow = "0 0 20px rgba(99, 102, 241, 0.5)";
        } else {
          opacity = 0.35;
        }
      }

      return {
        id: `node-${n.id}-${i}`,
        originalId: n.id,
        type: "traceNode",
        position: { x: 0, y: 0 }, // Will be auto-laid out
        data: {
          label: n.data?.label || n.data?.name || "Unknown",
          type: n.data?.type || "CUSTOM",
          status: n.data?.status || "OK",
          serviceName: n.data?.serviceName || "unknown",
          durationMs: n.data?.durationMs || 0,
          startAt: n.data?.startAt,
          endAt: n.data?.endAt,
          httpMethod: n.data?.httpMethod,
          httpPath: n.data?.httpPath,
          httpStatus: n.data?.httpStatus,
          dbSystem: n.data?.dbSystem,
          dbQuery: n.data?.dbQuery,
          dbRows: n.data?.dbRows,
          metadata: n.data?.metadata,
          isRoot,
          isTerminal,
          isMatch,
        },
        style: {
          opacity,
          filter: isSearching && !isMatch ? "grayscale(0.5)" : undefined,
        },
      };
    });

    // Build edges mapping
    const nodeIdMap = new Map<string, string>();
    baseNodes.forEach((n: any) => {
      nodeIdMap.set(n.originalId, n.id);
    });

    const edges = edgesData
      .map((e: any) => {
        const sourceId = nodeIdMap.get(e.source);
        const targetId = nodeIdMap.get(e.target);
        if (sourceId && targetId) {
          return {
            id: `edge-${sourceId}-${targetId}`,
            source: sourceId,
            target: targetId,
            type: "smoothstep",
            markerEnd: {
              type: MarkerType.ArrowClosed,
              width: 16,
              height: 16,
              color: "#6366f1",
            },
            style: {
              stroke: "#6366f1",
              strokeWidth: 2,
              strokeDasharray: "none",
            },
            animated: false,
          };
        }
        return null;
      })
      .filter(Boolean);

    return getLayoutedElements(baseNodes, edges);
  }, [q.data, searchQ.data, searchQuery, traceId]);

  const onNodeClick = useCallback((_: any, node: any) => {
    setSelectedNode(node);
  }, []);

  async function addAnnotation() {
    if (!note.trim() || !isMountedRef.current) return;
    try {
      await api.post(`/traces/${traceId}/annotations`, {
        nodeId: "annotation",
        text: note,
        x: 80,
        y: 80,
      });
      setNote("");
      if (isMountedRef.current) {
        q.refetch();
      }
    } catch (error) {
      if (!isMountedRef.current) return;
      throw error;
    }
  }

  // Stats
  const totalDuration = useMemo(() => {
    if (!q.data?.nodes?.length) return 0;
    const nodes = q.data.nodes;
    const startTimes = nodes
      .map((n: any) => new Date(n.data?.startAt).getTime())
      .filter((t: number) => !isNaN(t));
    const endTimes = nodes
      .map((n: any) => new Date(n.data?.endAt).getTime())
      .filter((t: number) => !isNaN(t));
    if (startTimes.length === 0 || endTimes.length === 0) return 0;
    return Math.max(...endTimes) - Math.min(...startTimes);
  }, [q.data]);

  const errorCount = useMemo(
    () =>
      q.data?.nodes?.filter((n: any) => n.data?.status === "ERROR").length || 0,
    [q.data],
  );

  return (
    <section className="space-y-4">
      {/* Header */}
      <header className="flex items-start justify-between flex-wrap gap-4">
        <div>
          <h3 className="text-xl font-bold text-slate-900 dark:text-white">
            Trace Details
          </h3>
          <p className="text-xs font-mono text-slate-500 dark:text-slate-400 mt-1">
            {traceId}
          </p>
          <div className="flex items-center gap-4 mt-3">
            <div className="flex items-center gap-2 bg-slate-50 dark:bg-slate-800 px-3 py-1.5 rounded-lg">
              <div className="w-2 h-2 rounded-full bg-indigo-500" />
              <span className="text-xs text-slate-600 dark:text-slate-300">
                {q.data?.nodes?.length || 0} nodes
              </span>
            </div>
            <div className="flex items-center gap-2 bg-slate-50 dark:bg-slate-800 px-3 py-1.5 rounded-lg">
              <div
                className={`w-2 h-2 rounded-full ${
                  errorCount > 0 ? "bg-red-500" : "bg-emerald-500"
                }`}
              />
              <span className="text-xs text-slate-600 dark:text-slate-300">
                {errorCount > 0 ? `${errorCount} error(s)` : "All OK"}
              </span>
            </div>
            {totalDuration > 0 && (
              <div className="flex items-center gap-2 bg-slate-50 dark:bg-slate-800 px-3 py-1.5 rounded-lg">
                <span className="text-xs text-slate-600 dark:text-slate-300">
                  {(totalDuration / 1000).toFixed(2)}s total
                </span>
              </div>
            )}
          </div>
        </div>

        <div className="flex gap-3">
          <div className="flex bg-white dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700 focus-within:ring-2 ring-indigo-500 overflow-hidden text-sm">
            <input
              className="px-3 py-2 w-64 outline-none bg-transparent text-slate-900 dark:text-white"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Buscar nós..."
            />
            {searchQ.isFetching && (
              <span className="text-slate-400 py-2 pr-3">...</span>
            )}
          </div>
          <div className="flex gap-2 text-sm">
            <input
              className="border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 rounded-lg px-3 py-2 text-slate-900 dark:text-white outline-none focus:ring-2 focus:ring-indigo-500"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Add note"
            />
            <button
              onClick={addAnnotation}
              className="px-4 py-2 rounded-lg bg-indigo-600 text-white font-medium hover:bg-indigo-700 transition-colors"
            >
              Salvar
            </button>
          </div>
        </div>
      </header>

      {/* Graph */}
      <div className="h-[70vh] bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodeClick={onNodeClick}
          fitView
          fitViewOptions={{ padding: 0.2 }}
          defaultEdgeOptions={{
            type: "smoothstep",
          }}
        >
          <Background color="#94a3b8" gap={20} size={1.5} />
          <MiniMap
            nodeStrokeColor="#6366f1"
            nodeColor="#e2e8f0"
            maskColor="rgba(15, 23, 42, 0.1)"
            color="#f8fafc"
          />
          <Controls className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700" />
        </ReactFlow>
      </div>

      {/* Node Details Panel */}
      {selectedNode && (
        <NodeDetails
          data={selectedNode}
          onClose={() => setSelectedNode(null)}
        />
      )}
    </section>
  );
}
