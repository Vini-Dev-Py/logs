import { MagnifyingGlassIcon } from '@heroicons/react/24/outline'
import { useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import { api } from '../api/client'

const rowVariants = {
  hidden: { opacity: 0, y: 8 },
  visible: (i: number) => ({ opacity: 1, y: 0, transition: { delay: i * 0.04, duration: 0.2 } }),
}

export function TracesPage() {
  const { t } = useTranslation()
  const [status, setStatus] = useState('')
  const [service, setService] = useState('')
  const [searchQuery, setSearchQuery] = useState('')

  const timeRange = useMemo(() => {
    const now = new Date()
    return {
      from: new Date(now.getTime() - 24 * 3600 * 1000).toISOString(),
      to: now.toISOString(),
    }
  }, [])

  const q = useQuery({
    queryKey: ['traces', status, service],
    queryFn: async () => {
      const params = new URLSearchParams({ from: timeRange.from, to: timeRange.to, status, service })
      return (await api.get(`/traces?${params.toString()}`)).data.items || []
    },
  })

  const searchQ = useQuery({
    queryKey: ['search', searchQuery],
    queryFn: async () => {
      if (!searchQuery) return null
      const params = new URLSearchParams({ query: searchQuery })
      return (await api.get(`/search?${params.toString()}`)).data.items || []
    },
    enabled: searchQuery.length > 2,
  })

  const isSearching = searchQuery.length > 2

  return (
    <section className="space-y-4">
      <header className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold text-slate-900 dark:text-white">{t('traces.title')}</h2>
      </header>

      {/* Filters */}
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-4 flex flex-col md:flex-row gap-3">
        <div className="flex-1 relative">
          <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-2.5 text-slate-400" />
          <input
            className="w-full pl-10 pr-4 py-2 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 outline-none transition"
            placeholder={t('traces.search_placeholder')}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <select
          title={t('traces.filter_status')}
          className="border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-indigo-500 transition"
          value={status}
          onChange={(e) => setStatus(e.target.value)}
        >
          <option value="">{t('traces.filter_status')}</option>
          <option value="OK">OK</option>
          <option value="ERROR">ERROR</option>
        </select>
        <input
          className="border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-indigo-500 transition"
          placeholder={t('traces.filter_service_placeholder')}
          value={service}
          onChange={(e) => setService(e.target.value)}
        />
      </div>

      {/* Table */}
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
            <tr>
              <th className="text-left p-4 font-medium">{t('traces.col_trace')}</th>
              {isSearching ? (
                <th className="text-left p-4 font-medium">{t('traces.col_content')}</th>
              ) : (
                <>
                  <th className="text-left p-4 font-medium">{t('traces.col_datetime')}</th>
                  <th className="text-left p-4 font-medium">{t('traces.col_endpoint')}</th>
                  <th className="text-left p-4 font-medium">{t('traces.col_status_duration')}</th>
                  <th className="text-left p-4 font-medium">{t('traces.col_service')}</th>
                </>
              )}
            </tr>
          </thead>
          <tbody>
            {!isSearching
              ? q.data?.map((tr: any, i: number) => (
                  <motion.tr
                    key={tr.traceId}
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
                        {tr.traceId.split('-')[0]}...
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
                        <span
                          className={`px-2.5 py-1 rounded-full text-xs font-semibold ${
                            tr.status === 'ERROR'
                              ? 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400'
                              : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400'
                          }`}
                        >
                          {tr.status}
                        </span>
                        <span className="text-slate-500 dark:text-slate-400">{tr.durationMs}ms</span>
                      </div>
                    </td>
                    <td className="p-4 text-slate-600 dark:text-slate-400 font-medium">
                      {tr.serviceName}
                    </td>
                  </motion.tr>
                ))
              : searchQ.data?.map((hit: any, i: number) => (
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
                          Trace: {hit.traceId.split('-')[0]}
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

            {((!isSearching && q.isLoading) || (isSearching && searchQ.isLoading)) && (
              <tr>
                <td colSpan={5} className="p-8 text-center text-slate-400">{t('traces.loading')}</td>
              </tr>
            )}

            {((!isSearching && !q.isLoading && !q.data?.length) ||
              (isSearching && !searchQ.isLoading && !searchQ.data?.length)) && (
              <tr>
                <td colSpan={5} className="p-8 text-center text-slate-500 dark:text-slate-400">
                  {t('traces.empty')}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </section>
  )
}
