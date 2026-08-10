import { useEffect, useState } from 'react'
import { api } from '../hooks/useApi'
import { RendimientoReal } from '../types'

function fmtPct(n: number) {
  return `${n >= 0 ? '+' : ''}${n.toFixed(2)}%`
}

function fmtARS(n: number) {
  return `ARS ${n.toLocaleString('es-AR', { maximumFractionDigits: 0 })}`
}

export default function RetornoReal() {
  const [data, setData] = useState<RendimientoReal | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .getRendimientoReal()
      .then(setData)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return <p className="text-gray-500 dark:text-gray-400 text-sm animate-pulse py-16 text-center">Calculando retorno real...</p>
  }
  if (error) {
    return (
      <div className="bg-red-100 border border-red-300 text-red-700 dark:bg-red-900/40 dark:border-red-700 dark:text-red-300 rounded-lg p-4 text-sm">
        ⚠️ {error}
      </div>
    )
  }
  if (!data || data.posiciones.length === 0) {
    return (
      <div className="text-center py-16 text-gray-500">
        <p className="text-4xl mb-3">📉</p>
        <p>No tenés posiciones abiertas.</p>
      </div>
    )
  }

  const nominalPositivo = data.total_retorno_nominal_pct >= 0
  const realPositivo = data.total_retorno_real_pct >= 0

  return (
    <div>
      {data.posiciones_sin_cotizar && data.posiciones_sin_cotizar.length > 0 && (
        <div className="bg-yellow-100 border border-yellow-300 text-yellow-800 dark:bg-yellow-900/20 dark:border-yellow-700/50 dark:text-yellow-400 rounded-lg p-3 mb-4 text-sm">
          ⚠️ No se pudo obtener el precio actual de {data.posiciones_sin_cotizar.join(', ')} — se excluyeron de este cálculo.
        </div>
      )}

      {/* Resumen en lenguaje natural */}
      <div className="bg-gradient-to-r from-blue-50 to-white dark:from-blue-950/60 dark:to-gray-900 border border-blue-200 dark:border-blue-900/50 rounded-xl p-5 mb-6">
        <p className="text-xs text-blue-700 dark:text-blue-300 uppercase tracking-wide mb-1">Retorno real vs inflación (INDEC, IPC Nacional)</p>
        <p className="text-lg font-medium text-gray-900 dark:text-white">{data.resumen}</p>
        <p className="text-xs text-gray-500 dark:text-gray-500 mt-2">Último dato de inflación disponible: {data.ipc_fecha_actual}</p>
      </div>

      {/* Tarjetas nominal vs real */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 sm:gap-4 mb-6">
        <div className="bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-800 rounded-xl p-4">
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide">Retorno nominal</p>
          <p className={`text-lg font-bold font-mono ${nominalPositivo ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
            {fmtPct(data.total_retorno_nominal_pct)}
          </p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{fmtARS(data.total_retorno_nominal_ars)}</p>
        </div>
        <div className="bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-800 rounded-xl p-4">
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide">Retorno real</p>
          <p className={`text-lg font-bold font-mono ${realPositivo ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
            {fmtPct(data.total_retorno_real_pct)}
          </p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{fmtARS(data.total_retorno_real_ars)}</p>
        </div>
        <div className="bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-800 rounded-xl p-4">
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide">Inflación acumulada</p>
          <p className="text-lg font-bold font-mono text-orange-600 dark:text-orange-400">
            {fmtPct(data.total_inflacion_ars_pct)}
          </p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">desde apertura de posiciones</p>
        </div>
        <div className="bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-800 rounded-xl p-4">
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide">Invertido</p>
          <p className="text-lg font-bold font-mono text-gray-900 dark:text-white">{fmtARS(data.total_invertido_ars)}</p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{fmtARS(data.total_actual_ars)} actual</p>
        </div>
      </div>

      {/* Tabla por posición */}
      <div className="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-gray-100 text-gray-500 dark:bg-gray-900 dark:text-gray-400 text-xs uppercase tracking-wide">
              <th className="px-4 py-3 text-left">Ticker</th>
              <th className="px-4 py-3 text-right">Invertido</th>
              <th className="px-4 py-3 text-right">Actual</th>
              <th className="px-4 py-3 text-right">Inflación ARS</th>
              <th className="px-4 py-3 text-right">Retorno nominal</th>
              <th className="px-4 py-3 text-right">Retorno real</th>
              <th className="px-4 py-3 text-right">Retorno real USD</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
            {data.posiciones.map((p) => (
              <tr key={p.ticker} className="bg-white hover:bg-gray-50 dark:bg-gray-950 dark:hover:bg-gray-900 transition-colors">
                <td className="px-4 py-3 font-bold text-gray-900 dark:text-white">{p.ticker}</td>
                <td className="px-4 py-3 text-right font-mono text-gray-600 dark:text-gray-300">{fmtARS(p.invertido_ars)}</td>
                <td className="px-4 py-3 text-right font-mono text-gray-900 dark:text-white">{fmtARS(p.valor_actual_ars)}</td>
                <td className="px-4 py-3 text-right font-mono text-orange-600 dark:text-orange-400">{fmtPct(p.inflacion_ars_pct)}</td>
                <td className={`px-4 py-3 text-right font-mono ${p.retorno_nominal_pct >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                  {fmtPct(p.retorno_nominal_pct)}
                </td>
                <td className={`px-4 py-3 text-right font-mono font-medium ${p.retorno_real_pct >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                  {fmtPct(p.retorno_real_pct)}
                </td>
                <td className="px-4 py-3 text-right font-mono text-gray-400 dark:text-gray-500 text-xs">
                  {p.retorno_real_usd_pct !== undefined
                    ? fmtPct(p.retorno_real_usd_pct)
                    : <span title="Sin CCL de apertura registrado o falta FRED_API_KEY">—</span>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
