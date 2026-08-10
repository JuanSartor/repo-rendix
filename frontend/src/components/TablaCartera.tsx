import { Posicion } from '../types'

interface Props {
  posiciones: Posicion[]
}

function fmt(n: number, decimals = 2) {
  return n.toLocaleString('es-AR', { minimumFractionDigits: decimals, maximumFractionDigits: decimals })
}

export default function TablaCartera({ posiciones }: Props) {
  if (posiciones.length === 0) {
    return (
      <div className="text-center py-16 text-gray-500 dark:text-gray-500">
        <p className="text-4xl mb-3">📭</p>
        <p>No tenés posiciones abiertas.</p>
        <p className="text-sm mt-1">Registrá tu primera compra.</p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800">
      <table className="w-full text-sm">
        <thead>
          <tr className="bg-gray-100 text-gray-500 dark:bg-gray-900 dark:text-gray-400 text-xs uppercase tracking-wide">
            <th className="px-4 py-3 text-left">Ticker</th>
            <th className="px-4 py-3 text-right">Cantidad</th>
            <th className="px-4 py-3 text-right">P. Prom ARS</th>
            <th className="px-4 py-3 text-right">P. Actual ARS</th>
            <th className="px-4 py-3 text-right">Invertido</th>
            <th className="px-4 py-3 text-right">Actual</th>
            <th className="px-4 py-3 text-right">P&L</th>
            <th className="px-4 py-3 text-right">%</th>
            <th className="px-4 py-3 text-center">Broker</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
          {posiciones.map((p) => {
            if (!p.precio_disponible) {
              return (
                <tr key={p.ticker} className="bg-white hover:bg-gray-50 dark:bg-gray-950 dark:hover:bg-gray-900 transition-colors opacity-60">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-gray-900 dark:text-white">{p.ticker}</span>
                      {p.es_cedear && (
                        <span className="text-xs bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300 px-1.5 py-0.5 rounded">CEDEAR</span>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-right font-mono text-gray-600 dark:text-gray-300">{fmt(p.cantidad)}</td>
                  <td className="px-4 py-3 text-right font-mono text-gray-600 dark:text-gray-300">{fmt(p.precio_prom_ars)}</td>
                  <td colSpan={5} className="px-4 py-3 text-center text-xs text-yellow-600 dark:text-yellow-500" title="No se pudo obtener el precio actual en este momento">
                    ⚠️ sin cotización — reintentando
                  </td>
                  <td className="px-4 py-3 text-center">
                    <span className="text-xs text-gray-500 dark:text-gray-400">{p.broker}</span>
                  </td>
                </tr>
              )
            }
            const positivo = p.pnl_ars >= 0
            return (
              <tr key={p.ticker} className="bg-white hover:bg-gray-50 dark:bg-gray-950 dark:hover:bg-gray-900 transition-colors">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-gray-900 dark:text-white">{p.ticker}</span>
                    {p.es_cedear && (
                      <span className="text-xs bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300 px-1.5 py-0.5 rounded">CEDEAR</span>
                    )}
                  </div>
                </td>
                <td className="px-4 py-3 text-right font-mono text-gray-600 dark:text-gray-300">{fmt(p.cantidad)}</td>
                <td className="px-4 py-3 text-right font-mono text-gray-600 dark:text-gray-300">{fmt(p.precio_prom_ars)}</td>
                <td className="px-4 py-3 text-right font-mono text-gray-900 dark:text-white">{fmt(p.precio_actual_ars)}</td>
                <td className="px-4 py-3 text-right font-mono text-gray-500 dark:text-gray-400">{fmt(p.cantidad * p.precio_prom_ars, 0)}</td>
                <td className="px-4 py-3 text-right font-mono text-gray-900 dark:text-white">{fmt(p.valor_total_ars, 0)}</td>
                <td className={`px-4 py-3 text-right font-mono font-medium ${positivo ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                  {positivo ? '+' : ''}{fmt(p.pnl_ars, 0)}
                </td>
                <td className={`px-4 py-3 text-right font-mono text-sm ${positivo ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                  <span className={`px-2 py-0.5 rounded text-xs ${positivo ? 'bg-green-100 dark:bg-green-900/40' : 'bg-red-100 dark:bg-red-900/40'}`}>
                    {positivo ? '+' : ''}{p.pnl_pct.toFixed(2)}%
                  </span>
                </td>
                <td className="px-4 py-3 text-center">
                  <span className="text-xs text-gray-500 dark:text-gray-400">{p.broker}</span>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
