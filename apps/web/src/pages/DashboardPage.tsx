import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import {
    Activity,
    AlertTriangle,
    ArrowUpRight,
    Server,
    TrendingUp,
} from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
    Bar,
    BarChart,
    CartesianGrid,
    Cell,
    Line,
    LineChart,
    Pie,
    PieChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";
import { api } from "../api/client";
import { Pagination } from "../components/Pagination";
import { Badge } from "../components/ui/Badge";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "../components/ui/Card";
import { SkeletonTable } from "../components/ui/Skeleton";

const COLORS = [
  "#6366f1",
  "#8b5cf6",
  "#a78bfa",
  "#c4b5fd",
  "#818cf8",
  "#7c3aed",
];

const CHART_COLORS = [
  "#6366f1",
  "#8b5cf6",
  "#ec4899",
  "#f59e0b",
  "#10b981",
  "#3b82f6",
  "#ef4444",
  "#14b8a6",
];

const cardVariants = {
  hidden: { y: 12, opacity: 0 },
  visible: (i: number) => ({
    y: 0,
    opacity: 1,
    transition: { delay: i * 0.08, duration: 0.3, ease: "easeOut" as const },
  }),
};

export function DashboardPage() {
  const { t } = useTranslation();
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const q = useQuery({
    queryKey: ["endpoints", "metrics"],
    queryFn: async () => {
      const params = new URLSearchParams({
        from: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)
          .toISOString()
          .split("T")[0],
        to: new Date().toISOString().split("T")[0],
      });
      const resp = await api.get(`/metrics/endpoints?${params.toString()}`);
      console.log("Dashboard response:", resp.data);
      return resp.data?.items || [];
    },
    retry: 1,
  });

  const endpoints = q.data || [];
  const sortedEndpoints = [...endpoints].sort(
    (a: any, b: any) => b.calls - a.calls,
  );
  const total = sortedEndpoints.length;
  const paginatedEndpoints = sortedEndpoints.slice(
    (page - 1) * pageSize,
    page * pageSize,
  );

  // Calculate metrics
  const totalCalls = endpoints.reduce(
    (sum: number, e: any) => sum + e.calls,
    0,
  );
  const uniqueServices = [...new Set(endpoints.map((e: any) => e.serviceName))]
    .length;
  const errorEndpoints = endpoints.filter(
    (e: any) => e.httpMethod === "ERROR",
  ).length;
  const avgCallsPerEndpoint = total > 0 ? Math.round(totalCalls / total) : 0;

  // Prepare chart data
  const topEndpointsChartData = sortedEndpoints.slice(0, 10).map((e: any) => ({
    name: `${e.httpMethod} ${e.httpPath}`,
    calls: e.calls,
    service: e.serviceName,
  }));

  const servicesChartData = endpoints
    .reduce((acc: any[], e: any) => {
      const existing = acc.find((s) => s.name === e.serviceName);
      if (existing) {
        existing.calls += e.calls;
      } else {
        acc.push({ name: e.serviceName, calls: e.calls });
      }
      return acc;
    }, [])
    .sort((a: any, b: any) => b.calls - a.calls)
    .slice(0, 8);

  const methodDistribution = endpoints.reduce(
    (acc: any, e: any) => {
      acc[e.httpMethod] = (acc[e.httpMethod] || 0) + e.calls;
      return acc;
    },
    {} as Record<string, number>,
  );

  const methodChartData = Object.entries(methodDistribution).map(
    ([name, value]) => ({
      name,
      calls: value,
    }),
  );

  if (q.isLoading) {
    return (
      <section className="space-y-6">
        <div>
          <div className="h-8 w-48 bg-slate-200 dark:bg-slate-800 rounded animate-pulse" />
          <div className="h-4 w-72 bg-slate-200 dark:bg-slate-800 rounded animate-pulse mt-2" />
        </div>

        {/* Skeleton metric cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div
              key={i}
              className="rounded-xl border border-slate-200 dark:border-slate-800 p-6 space-y-3"
            >
              <div className="h-4 w-24 bg-slate-200 dark:bg-slate-800 rounded animate-pulse" />
              <div className="h-8 w-16 bg-slate-200 dark:bg-slate-800 rounded animate-pulse" />
            </div>
          ))}
        </div>

        <SkeletonTable rows={8} />
      </section>
    );
  }

  if (q.isError) {
    return (
      <section className="space-y-6">
        <Card className="border-red-200 dark:border-red-800">
          <CardContent className="p-6">
            <div className="flex items-center gap-3 text-red-600 dark:text-red-400">
              <AlertTriangle className="h-5 w-5" />
              <div className="flex-1">
                <p className="font-semibold">Erro ao carregar métricas</p>
                <p className="text-sm text-red-500 dark:text-red-400">
                  {(q.error as Error)?.message || "Erro desconhecido"}
                </p>
              </div>
              <button
                onClick={() => q.refetch()}
                className="flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-red-700 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition-colors"
              >
                Tentar novamente
              </button>
            </div>
          </CardContent>
        </Card>
      </section>
    );
  }

  if (endpoints.length === 0) {
    return (
      <section className="space-y-6">
        <header>
          <h2 className="text-2xl font-semibold text-slate-900 dark:text-white">
            {t("dashboard.title")}
          </h2>
          <p className="text-slate-500 dark:text-slate-400 text-sm mt-0.5">
            {t("dashboard.subtitle")}
          </p>
        </header>

        <Card className="border-dashed">
          <CardContent className="p-12 text-center">
            <Server className="h-12 w-12 mx-auto text-slate-300 dark:text-slate-700 mb-4" />
            <h3 className="text-lg font-semibold text-slate-700 dark:text-slate-300 mb-2">
              Nenhuma métrica disponível
            </h3>
            <p className="text-sm text-slate-500 dark:text-slate-400 max-w-md mx-auto mb-4">
              Execute o simulador para gerar dados de teste e populá o
              dashboard.
            </p>
            <button
              onClick={() => q.refetch()}
              className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium"
            >
              Atualizar dados
            </button>
          </CardContent>
        </Card>
      </section>
    );
  }

  return (
    <section className="space-y-6">
      <header>
        <h2 className="text-2xl font-semibold text-slate-900 dark:text-white">
          {t("dashboard.title")}
        </h2>
        <p className="text-slate-500 dark:text-slate-400 text-sm mt-0.5">
          {t("dashboard.subtitle")}
        </p>
      </header>

      {/* Metric Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <motion.div
          custom={0}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-sm font-medium text-slate-600 dark:text-slate-400">
                Total de Chamadas
              </CardTitle>
              <Activity className="h-4 w-4 text-indigo-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-slate-900 dark:text-white">
                {totalCalls.toLocaleString()}
              </div>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                {total} endpoints monitorados
              </p>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div
          custom={1}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-sm font-medium text-slate-600 dark:text-slate-400">
                Serviços Ativos
              </CardTitle>
              <Server className="h-4 w-4 text-emerald-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-slate-900 dark:text-white">
                {uniqueServices}
              </div>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                serviços únicos
              </p>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div
          custom={2}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-sm font-medium text-slate-600 dark:text-slate-400">
                Média por Endpoint
              </CardTitle>
              <TrendingUp className="h-4 w-4 text-amber-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-slate-900 dark:text-white">
                {avgCallsPerEndpoint.toLocaleString()}
              </div>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                chamadas/endpoint
              </p>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div
          custom={3}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-sm font-medium text-slate-600 dark:text-slate-400">
                Top Serviço
              </CardTitle>
              <ArrowUpRight className="h-4 w-4 text-violet-500" />
            </CardHeader>
            <CardContent>
              <div className="text-lg font-bold text-slate-900 dark:text-white truncate">
                {servicesChartData[0]?.name || "—"}
              </div>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                {servicesChartData[0]?.calls.toLocaleString() || 0} chamadas
              </p>
            </CardContent>
          </Card>
        </motion.div>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <motion.div
          custom={4}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <Card className="h-full">
            <CardHeader>
              <CardTitle>Top Endpoints</CardTitle>
              <CardDescription>10 endpoints com mais chamadas</CardDescription>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={topEndpointsChartData}>
                  <CartesianGrid
                    strokeDasharray="3 3"
                    className="stroke-slate-200 dark:stroke-slate-800"
                  />
                  <XAxis
                    dataKey="name"
                    className="text-xs"
                    tick={{ fill: "currentColor", fontSize: 11 }}
                    angle={-45}
                    textAnchor="end"
                    height={80}
                  />
                  <YAxis
                    className="text-xs"
                    tick={{ fill: "currentColor", fontSize: 11 }}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "hsl(0 0% 100%)",
                      border: "1px solid hsl(220 13% 91%)",
                      borderRadius: "8px",
                      fontSize: "12px",
                    }}
                    wrapperClassName="dark:[&_*]:!bg-slate-900 dark:[&_*]:!border-slate-800"
                  />
                  <Bar dataKey="calls" fill="#6366f1" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div
          custom={5}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <Card className="h-full">
            <CardHeader>
              <CardTitle>Distribuição por Serviço</CardTitle>
              <CardDescription>Chamadas por serviço</CardDescription>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={servicesChartData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percent }: any) =>
                      `${name} (${(percent * 100).toFixed(0)}%)`
                    }
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="calls"
                  >
                    {servicesChartData.map((_entry: any, index: number) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={CHART_COLORS[index % CHART_COLORS.length]}
                      />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </motion.div>
      </div>

      {/* Method Distribution */}
      <motion.div
        custom={6}
        initial="hidden"
        animate="visible"
        variants={cardVariants}
      >
        <Card>
          <CardHeader>
            <CardTitle>Distribuição por Método HTTP</CardTitle>
            <CardDescription>Chamadas agrupadas por método</CardDescription>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={methodChartData}>
                <CartesianGrid
                  strokeDasharray="3 3"
                  className="stroke-slate-200 dark:stroke-slate-800"
                />
                <XAxis
                  dataKey="name"
                  className="text-xs"
                  tick={{ fill: "currentColor", fontSize: 12 }}
                />
                <YAxis
                  className="text-xs"
                  tick={{ fill: "currentColor", fontSize: 12 }}
                />
                <Tooltip />
                <Line
                  type="monotone"
                  dataKey="calls"
                  stroke="#6366f1"
                  strokeWidth={2}
                  dot={{ fill: "#6366f1", r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </motion.div>

      {/* Endpoints Table */}
      <motion.div
        custom={7}
        initial="hidden"
        animate="visible"
        variants={cardVariants}
      >
        <Card>
          <CardHeader>
            <CardTitle>{t("dashboard.top_endpoints")}</CardTitle>
            <CardDescription>
              Detalhamento de todos os endpoints
            </CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 dark:bg-slate-800 text-slate-500 dark:text-slate-400">
                <tr>
                  <th className="text-left p-4 font-medium">
                    {t("dashboard.col_service")}
                  </th>
                  <th className="text-left p-4 font-medium">
                    {t("dashboard.col_endpoint")}
                  </th>
                  <th className="text-right p-4 font-medium">
                    {t("dashboard.col_calls")}
                  </th>
                </tr>
              </thead>
              <tbody>
                {paginatedEndpoints.map((item: any, i: number) => (
                  <motion.tr
                    key={`${item.serviceName}-${item.httpMethod}-${item.httpPath}-${i}`}
                    custom={i}
                    initial="hidden"
                    animate="visible"
                    variants={{
                      hidden: { opacity: 0, y: 8 },
                      visible: {
                        opacity: 1,
                        y: 0,
                        transition: { delay: i * 0.03, duration: 0.15 },
                      },
                    }}
                    className="border-t border-slate-100 dark:border-slate-800 hover:bg-slate-50/80 dark:hover:bg-slate-800/60 transition-colors"
                  >
                    <td className="p-4 text-slate-700 dark:text-slate-300 font-medium">
                      {item.serviceName}
                    </td>
                    <td className="p-4">
                      <div className="flex items-center gap-2">
                        <Badge
                          variant={
                            item.httpMethod === "GET"
                              ? "info"
                              : item.httpMethod === "POST"
                                ? "success"
                                : item.httpMethod === "PUT"
                                  ? "warning"
                                  : item.httpMethod === "DELETE"
                                    ? "destructive"
                                    : "outline"
                          }
                        >
                          {item.httpMethod}
                        </Badge>
                        <code className="font-mono text-xs text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-1.5 py-0.5 rounded">
                          {item.httpPath}
                        </code>
                      </div>
                    </td>
                    <td className="p-4 text-right font-mono font-semibold text-indigo-600 dark:text-indigo-400">
                      {item.calls.toLocaleString()}
                    </td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      </motion.div>

      {/* Pagination */}
      {total > pageSize && (
        <Pagination
          page={page}
          pageSize={pageSize}
          total={total}
          onPageChange={setPage}
        />
      )}
    </section>
  );
}
