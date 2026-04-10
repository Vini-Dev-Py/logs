import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import { api } from '../api/client'

const rowVariants = {
  hidden: { opacity: 0, y: 8 },
  visible: (i: number) => ({ opacity: 1, y: 0, transition: { delay: i * 0.05, duration: 0.2 } }),
}

export function DashboardPage() {
  const { t } = useTranslation()
  const q = useQuery({
    queryKey: ['endpoints', 'metrics'],
    queryFn: async () => (await api.get('/metrics/endpoints')).data.items || [],
  })

  const topEndpoints = (q.data || []).sort((a: any, b: any) => b.calls - a.calls).slice(0, 10)

  return (
    <section className="space-y-6">
      <header>
        <h2 className="text-2xl font-semibold text-slate-900 dark:text-white">{t('dashboard.title')}</h2>
        <p className="text-slate-500 dark:text-slate-400 text-sm mt-0.5">{t('dashboard.subtitle')}</p>
      </header>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <motion.div
          initial={{ y: 12, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.3, ease: 'easeOut' }}
          whileHover={{ scale: 1.005 }}
          className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm"
        >
          <div className="p-5 border-b border-slate-100 dark:border-slate-800">
            <h3 className="font-semibold text-slate-800 dark:text-slate-200 text-sm">
              {t('dashboard.top_endpoints')}
            </h3>
          </div>

          <table className="w-full text-sm">
            <thead className="bg-slate-50 dark:bg-slate-800 text-slate-500 dark:text-slate-400">
              <tr>
                <th className="text-left p-4 font-medium">{t('dashboard.col_service')}</th>
                <th className="text-left p-4 font-medium">{t('dashboard.col_endpoint')}</th>
                <th className="text-right p-4 font-medium">{t('dashboard.col_calls')}</th>
              </tr>
            </thead>
            <tbody>
              {topEndpoints.map((item: any, i: number) => (
                <motion.tr
                  key={`${item.serviceName}-${item.httpMethod}-${item.httpPath}`}
                  custom={i}
                  initial="hidden"
                  animate="visible"
                  variants={rowVariants}
                  className="border-t border-slate-100 dark:border-slate-800 hover:bg-slate-50/80 dark:hover:bg-slate-800/60 transition-colors"
                >
                  <td className="p-4 text-slate-700 dark:text-slate-300 font-medium">{item.serviceName}</td>
                  <td className="p-4">
                    <div className="flex items-center gap-2">
                      <span
                        className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                          item.httpMethod === 'GET'
                            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400'
                            : item.httpMethod === 'POST'
                            ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400'
                            : item.httpMethod === 'PUT'
                            ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400'
                            : item.httpMethod === 'DELETE'
                            ? 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400'
                            : 'bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-300'
                        }`}
                      >
                        {item.httpMethod}
                      </span>
                      <span className="font-mono text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-1 py-0.5 rounded text-xs">
                        {item.httpPath}
                      </span>
                    </div>
                  </td>
                  <td className="p-4 text-right font-mono font-semibold text-indigo-600 dark:text-indigo-400">
                    {item.calls}
                  </td>
                </motion.tr>
              ))}

              {q.isLoading && (
                <tr>
                  <td colSpan={3} className="p-6 text-center text-slate-400">{t('dashboard.loading')}</td>
                </tr>
              )}
              {!q.isLoading && topEndpoints.length === 0 && (
                <tr>
                  <td colSpan={3} className="p-6 text-center text-slate-400">{t('dashboard.empty')}</td>
                </tr>
              )}
            </tbody>
          </table>
        </motion.div>
      </div>
    </section>
  )
}
