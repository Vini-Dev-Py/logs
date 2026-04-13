import { MagnifyingGlassIcon } from "@heroicons/react/24/outline";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { AlertTriangle, RefreshCw } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { Pagination } from "../components/Pagination";
import { Badge } from "../components/ui/Badge";
import { SkeletonTable } from "../components/ui/Skeleton";

const rowVariants = {
  hidden: { opacity: 0, y: 8 },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: i * 0.04, duration: 0.2 },
  }),
};

const PAGE_SIZE_KEY = "traces_page_size";

function getInitialPageSize(): number {
  try {
    const saved = localStorage.getItem(PAGE_SIZE_KEY);
    return saved ? parseInt(saved, 10) : 50;
  } catch {
    return 50;
  }
}

export function TracesPage() {
  const { t } = useTranslation();
  const [status, setStatus] = useState("");
  const [service, setService] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(getInitialPageSize);
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date();
    d.setDate(1); // primeiro dia do mês
    d.setHours(0, 0, 0, 0);
    return d.toISOString().slice(0, 16);
  });
  const [dateTo, setDateTo] = useState(() =>
    new Date().toISOString().slice(0, 16),
  );
  const queryClient = useQueryClient();
  const isMountedRef = useRef(true);

  // Cleanup on unmount
  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
      // React Query gerencia cache automaticamente - não cancelar queries no cleanup
    };
  }, []);

  // Reset page when filters change
  useEffect(() => {
    setPage(1);
  }, [status, service, searchQuery]);

  // Save page size to localStorage
  useEffect(() => {
    try {
      localStorage.setItem(PAGE_SIZE_KEY, pageSize.toString());
    } catch {}
  }, [pageSize]);

  const q = useQuery({
    queryKey: ["traces", status, service, page, pageSize, dateFrom, dateTo],
    queryFn: async () => {
      const params = new URLSearchParams({
        from: new Date(dateFrom).toISOString(),
        to: new Date(dateTo).toISOString(),
        status,
        service,
        page: page.toString(),
        pageSize: pageSize.toString(),
      });
      return (await api.get(`/traces?${params.toString()}`)).data;
    },
  });

  const searchQ = useQuery({
    queryKey: ["search", searchQuery, page, pageSize],
    queryFn: async () => {
      if (!searchQuery) return null;
      const params = new URLSearchParams({
        query: searchQuery,
        page: page.toString(),
        pageSize: pageSize.toString(),
      });
      return (await api.get(`/search?${params.toString()}`)).data;
    },
    enabled: searchQuery.length > 2,
  });

  const isSearching = searchQuery.length > 2;
  const tracesData = isSearching ? searchQ.data : q.data;
  const items = tracesData?.items || [];
  const total = tracesData?.total || 0;

  const handlePageSizeChange = (newPageSize: number) => {
    setPageSize(newPageSize);
    setPage(1);
  };

  return (
    <section className="space-y-4">
      <header className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold text-slate-900 dark:text-white">
          {t("traces.title")}
        </h2>
      </header>

      {/* Filters */}
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-4 space-y-3">
        {/* Date range */}
        <div className="flex flex-wrap items-center gap-3 text-sm">
          <span className="text-slate-600 dark:text-slate-400 font-medium">
            Período:
          </span>
          <input
            type="datetime-local"
            className="border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-indigo-500 transition"
            value={dateFrom}
            onChange={(e) => {
              setDateFrom(e.target.value);
              setPage(1);
            }}
          />
          <span className="text-slate-400">→</span>
          <input
            type="datetime-local"
            className="border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-indigo-500 transition"
            value={dateTo}
            onChange={(e) => {
              setDateTo(e.target.value);
              setPage(1);
            }}
          />
          <button
            className="ml-auto text-xs text-indigo-600 dark:text-indigo-400 hover:underline"
            onClick={() => {
              const now = new Date();
              setDateFrom(
                new Date(now.getFullYear(), now.getMonth(), 1)
                  .toISOString()
                  .slice(0, 16),
              );
              setDateTo(now.toISOString().slice(0, 16));
              setPage(1);
            }}
          >
            Este mês
          </button>
          <button
            className="text-xs text-indigo-600 dark:text-indigo-400 hover:underline"
            onClick={() => {
              const now = new Date();
              setDateFrom(
                new Date(now.getTime() - 24 * 3600 * 1000)
                  .toISOString()
                  .slice(0, 16),
              );
              setDateTo(now.toISOString().slice(0, 16));
              setPage(1);
            }}
          >
            Últimas 24h
          </button>
          <button
            className="text-xs text-indigo-600 dark:text-indigo-400 hover:underline"
            onClick={() => {
              const now = new Date();
              setDateFrom(
                new Date(now.getTime() - 7 * 24 * 3600 * 1000)
                  .toISOString()
                  .slice(0, 16),
              );
              setDateTo(now.toISOString().slice(0, 16));
              setPage(1);
            }}
          >
            Últimos 7 dias
          </button>
        </div>

        {/* Search + filters row */}
        <div className="flex flex-col md:flex-row gap-3">
          <div className="flex-1 relative">
            <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-2.5 text-slate-400" />
            <input
              className="w-full pl-10 pr-4 py-2 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 outline-none transition"
              placeholder={t("traces.search_placeholder")}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
          <select
            title={t("traces.filter_status")}
            className="border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-indigo-500 transition"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
          >
            <option value="">{t("traces.filter_status")}</option>
            <option value="OK">OK</option>
            <option value="ERROR">ERROR</option>
          </select>
          <input
            className="border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-indigo-500 transition"
            placeholder={t("traces.filter_service_placeholder")}
            value={service}
            onChange={(e) => setService(e.target.value)}
          />
        </div>
      </div>

      {/* Table */}
      {q.isLoading && !isSearching ? (
        <SkeletonTable rows={8} />
      ) : q.isError ? (
        <div className="bg-white dark:bg-slate-900 border border-red-200 dark:border-red-800 rounded-xl p-6">
          <div className="flex items-center gap-3 text-red-600 dark:text-red-400">
            <AlertTriangle className="h-5 w-5" />
            <div className="flex-1">
              <p className="font-semibold">Erro ao carregar traces</p>
              <p className="text-sm text-red-500 dark:text-red-400">
                Não foi possível conectar ao servidor. Tente novamente.
              </p>
            </div>
            <button
              onClick={() => q.refetch()}
              className="flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-red-700 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition-colors"
            >
              <RefreshCw className="h-3 w-3" />
              Tentar novamente
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm">
          <table className="w-full text-sm">
            <thead className="bg-slate-50 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
              <tr>
                <th className="text-left p-4 font-medium">
                  {t("traces.col_trace")}
                </th>
                {isSearching ? (
                  <th className="text-left p-4 font-medium">
                    {t("traces.col_content")}
                  </th>
                ) : (
                  <>
                    <th className="text-left p-4 font-medium">
                      {t("traces.col_datetime")}
                    </th>
                    <th className="text-left p-4 font-medium">
                      {t("traces.col_endpoint")}
                    </th>
                    <th className="text-left p-4 font-medium">
                      {t("traces.col_status_duration")}
                    </th>
                    <th className="text-left p-4 font-medium">
                      {t("traces.col_service")}
                    </th>
                  </>
                )}
              </tr>
            </thead>
            <tbody>
              {!isSearching
                ? items.map((tr: any, i: number) => (
                    <motion.tr
                      key={`${tr.traceId}-${i}`}
                      custom={i}
                      initial="hidden"
                      animate="visible"
                      variants={rowVariants}
                      className="border-t border-slate-100 dark:border-slate-800 hover:bg-slate-50/80 dark:hover:bg-slate-800/60 transition-colors"
                    >
                      <td className="p-4 font-mono text-xs text-slate-500 dark:text-slate-400">
                        <Link
                          className="text-indigo-600 dark:text-indigo-400 hover:underline"
                          to={`/traces/${tr.traceId}`}
                        >
                          {tr.traceId.split("-")[0]}...
                        </Link>
                      </td>
                      <td className="p-4 text-slate-600 dark:text-slate-400">
                        {new Date(tr.startedAt).toLocaleString()}
                      </td>
                      <td className="p-4 font-medium text-slate-800 dark:text-slate-200">
                        {tr.httpMethod} {tr.httpPath}
                      </td>
                      <td className="p-4">
                        <div className="flex items-center gap-3">
                          <Badge
                            variant={
                              tr.status === "ERROR" ? "destructive" : "success"
                            }
                          >
                            {tr.status}
                          </Badge>
                          <span className="text-slate-500 dark:text-slate-400">
                            {tr.durationMs}ms
                          </span>
                        </div>
                      </td>
                      <td className="p-4 text-slate-600 dark:text-slate-400 font-medium">
                        {tr.serviceName}
                      </td>
                    </motion.tr>
                  ))
                : items.map((hit: any, i: number) => (
                    <motion.tr
                      key={`${hit.traceId}-${i}`}
                      custom={i}
                      initial="hidden"
                      animate="visible"
                      variants={rowVariants}
                      className="border-t border-slate-100 dark:border-slate-800 hover:bg-slate-50/80 dark:hover:bg-slate-800/60 transition-colors"
                    >
                      <td className="p-4">
                        <div className="flex flex-col gap-1">
                          <Link
                            className="font-mono text-xs text-indigo-600 dark:text-indigo-400 hover:underline"
                            to={`/traces/${hit.traceId}`}
                          >
                            Trace: {hit.traceId.split("-")[0]}
                          </Link>
                          <span className="text-xs text-slate-500 dark:text-slate-400">
                            Nó: {hit.name} ({hit.type})
                          </span>
                        </div>
                      </td>
                      <td className="p-4 text-slate-700 dark:text-slate-300">
                        {hit.dbQuery && (
                          <code className="block bg-slate-100 dark:bg-slate-800 p-2 rounded text-xs text-slate-600 dark:text-slate-400 font-mono whitespace-pre-wrap">
                            {hit.dbQuery}
                          </code>
                        )}
                        {hit.metadata && (
                          <span className="text-xs text-slate-500 font-mono truncate max-w-96 block">
                            {hit.metadata}
                          </span>
                        )}
                      </td>
                    </motion.tr>
                  ))}

              {searchQ.isLoading && (
                <tr>
                  <td colSpan={isSearching ? 2 : 5} className="p-8 text-center">
                    <div className="flex items-center justify-center gap-2 text-slate-400">
                      <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                        <circle
                          className="opacity-25"
                          cx="12"
                          cy="12"
                          r="10"
                          stroke="currentColor"
                          strokeWidth="4"
                          fill="none"
                        />
                        <path
                          className="opacity-75"
                          fill="currentColor"
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                        />
                      </svg>
                      {t("traces.loading")}
                    </div>
                  </td>
                </tr>
              )}

              {!items.length && !q.isLoading && !searchQ.isLoading && (
                <tr>
                  <td
                    colSpan={isSearching ? 2 : 5}
                    className="p-8 text-center text-slate-500 dark:text-slate-400"
                  >
                    {t("traces.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination */}
      {total > 0 && (
        <Pagination
          page={page}
          pageSize={pageSize}
          total={total}
          onPageChange={setPage}
          onPageSizeChange={handlePageSizeChange}
        />
      )}
    </section>
  );
}
