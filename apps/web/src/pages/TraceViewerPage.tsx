import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState } from "react";
import ReactFlow from "react-flow-renderer";
import { useParams } from "react-router-dom";
import { api } from "../api/client";

export function TraceViewerPage() {
  const { traceId = "" } = useParams();
  const [note, setNote] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
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

  // Cleanup on unmount to prevent memory leaks
  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
      // Cancel ongoing queries when component unmounts
      queryClient.cancelQueries({ queryKey: ["trace", traceId] });
      queryClient.cancelQueries({ queryKey: ["search", searchQuery] });
    };
  }, [traceId, searchQuery, queryClient]);

  const { nodes, edges } = useMemo(() => {
    const matchedNodeIds = new Set(
      (searchQ.data || [])
        .filter((r: any) => r.traceId === traceId)
        .map((r: any) => r.nodeId),
    );
    const isSearching = searchQuery.length > 2;

    // Build nodes with unique IDs
    const baseNodes = (q.data?.nodes || []).map((n: any, i: number) => {
      let isMatch = false;
      if (isSearching) {
        if (matchedNodeIds.has(n.id)) {
          isMatch = true;
        }
      }

      let borderStyle =
        n.data.status === "ERROR"
          ? "2px solid #dc2626"
          : n.data.durationMs > 1000
            ? "2px solid #f59e0b"
            : "1px solid #cbd5e1";
      let opacity = 1;
      let boxShadow = "none";

      if (isSearching) {
        if (isMatch) {
          borderStyle = "3px solid #6366f1"; // Indigo highlight
          boxShadow = "0 0 15px rgba(99, 102, 241, 0.6)";
        } else {
          opacity = 0.4;
        }
      }

      return {
        id: `node-${n.id}-${i}`,
        originalId: n.id,
        data: { label: `${n.data.label} • ${n.data.type}` },
        position: { x: (i % 4) * 260, y: Math.floor(i / 4) * 120 },
        style: {
          border: borderStyle,
          borderRadius: 12,
          padding: 8,
          background: "white",
          opacity,
          boxShadow,
        },
      };
    });

    const annotationNodes = (q.data?.annotations || []).map((a: any) => ({
      id: `ann-${a.id}`,
      data: { label: `📝 ${a.text}` },
      position: { x: a.x, y: a.y },
      style: {
        background: "#fffbeb",
        border: "1px solid #f59e0b",
        borderRadius: 10,
      },
    }));

    // Build edges mapping original IDs to new IDs
    const nodeIdMap = new Map<string, string>();
    baseNodes.forEach((n: any) => {
      nodeIdMap.set(n.originalId, n.id);
    });

    const edges = (q.data?.edges || [])
      .map((e: any) => {
        const sourceId = nodeIdMap.get(e.source);
        const targetId = nodeIdMap.get(e.target);
        if (sourceId && targetId) {
          return {
            id: `edge-${sourceId}-${targetId}`,
            source: sourceId,
            target: targetId,
            type: "smoothstep",
            style: { stroke: "#6366f1", strokeWidth: 2 },
            animated: false,
          };
        }
        return null;
      })
      .filter(Boolean);

    return { nodes: [...baseNodes, ...annotationNodes], edges };
  }, [q.data, searchQ.data, searchQuery]);

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
      // Only refetch if component is still mounted
      if (isMountedRef.current) {
        q.refetch();
      }
    } catch (error) {
      // Ignore error if component is unmounted
      if (!isMountedRef.current) return;
      throw error;
    }
  }

  return (
    <section className="space-y-4">
      <header className="flex items-center justify-between flex-wrap gap-4">
        <h3 className="text-xl font-semibold">Trace {traceId}</h3>
        <div className="flex gap-4">
          <div className="flex bg-white rounded-lg border focus-within:ring-2 ring-indigo-500 overflow-hidden text-sm">
            <input
              className="px-3 py-2 w-64 outline-none"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Buscar nós (Ctrl+F)"
            />
            {searchQ.isFetching && (
              <span className="text-slate-400 py-2 pr-3">...</span>
            )}
          </div>
          <div className="flex gap-2 text-sm">
            <input
              className="border rounded-lg px-3 py-2"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Add note"
            />
            <button
              onClick={addAnnotation}
              className="px-3 py-2 rounded-lg bg-indigo-600 text-white font-medium hover:bg-indigo-700"
            >
              Salvar nota
            </button>
          </div>
        </div>
      </header>
      <div className="h-[78vh] bg-white border border-slate-200 rounded-xl overflow-hidden">
        <ReactFlow nodes={nodes} edges={edges} fitView />
      </div>
    </section>
  );
}
