import { motion } from 'framer-motion'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { api } from '../api/client'
import { useAuthStore } from '../store/auth'

export function SettingsUsersPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const user = useAuthStore(s => s.user)

  const qUsers = useQuery({
    queryKey: ['users'],
    queryFn: async () => (await api.get('/users')).data.items || [],
  })
  const qRoles = useQuery({
    queryKey: ['roles'],
    queryFn: async () => (await api.get('/roles')).data.items || [],
  })

  const updateRole = useMutation({
    mutationFn: async ({ id, roleId }: { id: string; roleId: string }) =>
      api.put(`/users/${id}/role`, { roleId }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })

  if (!user?.permissions?.includes('users:manage')) {
    return (
      <div className="p-6 text-red-500">{t('settings.access_denied')}</div>
    )
  }

  return (
    <div className="space-y-6 max-w-4xl">
      <header>
        <h2 className="text-2xl font-semibold text-slate-900 dark:text-white">{t('settings.title')}</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">{t('settings.subtitle')}</p>
      </header>

      <motion.div
        initial={{ y: 10, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ duration: 0.25 }}
        className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm"
      >
        <table className="w-full text-sm">
          <thead className="bg-slate-50 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
            <tr>
              <th className="text-left p-4 font-medium">{t('settings.col_user')}</th>
              <th className="text-left p-4 font-medium">{t('settings.col_role')}</th>
            </tr>
          </thead>
          <tbody>
            {(qUsers.data || []).map((u: any) => (
              <tr
                key={u.id}
                className="border-t border-slate-100 dark:border-slate-800 hover:bg-slate-50/80 dark:hover:bg-slate-800/60 transition-colors"
              >
                <td className="p-4">
                  <div className="font-medium text-slate-700 dark:text-slate-200">{u.name}</div>
                  <div className="text-slate-500 dark:text-slate-400 text-xs">{u.email}</div>
                </td>
                <td className="p-4">
                  <select
                    title={t('settings.col_role')}
                    className="border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg p-1.5 min-w-[200px] text-sm outline-none focus:ring-2 focus:ring-indigo-500 transition"
                    value={qRoles.data?.find((r: any) => r.name === u.role)?.id || ''}
                    onChange={(e) => updateRole.mutate({ id: u.id, roleId: e.target.value })}
                    disabled={updateRole.isPending}
                  >
                    <option value="" disabled>
                      {t('settings.no_role')}
                    </option>
                    {(qRoles.data || []).map((r: any) => (
                      <option key={r.id} value={r.id}>
                        {r.name}
                      </option>
                    ))}
                  </select>
                </td>
              </tr>
            ))}
            {qUsers.isLoading && (
              <tr>
                <td colSpan={2} className="p-6 text-center text-slate-400">
                  {t('settings.loading')}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </motion.div>
    </div>
  )
}
