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
      <div className="text-center py-16 text-gray-500">
        <p className="text-4xl mb-3">📭</p>
        <p>No tenés posiciones abiertas.</p>
        <p className="text-sm mt-1">Registrá tu primera compra.</p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto rounded-xl border border-gray-800">
      <table className="w-full text-sm">
        <thead>
          <tr className="bg-gray-900 text-gray-400 text-xs uppercase tracking-wide">
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
        <tbody className="divide-y divide-gray-800">
          {posiciones.map((p) => {
            const positivo = p.pnl_ars >= 0
            return (
              <tr key={p.ticker} className="bg-gray-950 hover:bg-gray-900 transition-colors">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-white">{p.ticker}</span>
                    {p.es_cedear && (
                      <span className="text-xs bg-blue-900 text-blue-300 px-1.5 py-0.5 rounded">CEDEAR</span>
                    )}
                  </div>
                </td>
                <td className="px-4 py-3 text-right font-mono text-gray-300">{fmt(p.cantidad)}</td>
                <td className="px-4 py-3 text-right font-mono text-gray-300">{fmt(p.precio_prom_ars)}</td>
                <td className="px-4 py-3 text-right font-mono text-white">{fmt(p.precio_actual_ars)}</td>
                <td className="px-4 py-3 text-right font-mono text-gray-400">{fmt(p.cantidad * p.precio_prom_ars, 0)}</td>
                <td className="px-4 py-3 text-right font-mono text-white">{fmt(p.valor_total_ars, 0)}</td>
                <td className={`px-4 py-3 text-right font-mono font-medium ${positivo ? 'text-green-400' : 'text-red-400'}`}>
                  {positivo ? '+' : ''}{fmt(p.pnl_ars, 0)}
                </td>
                <td className={`px-4 py-3 text-right font-mono text-sm ${positivo ? 'text-green-400' : 'text-red-400'}`}>
                  <span className={`px-2 py-0.5 rounded text-xs ${positivo ? 'bg-green-900/40' : 'bg-red-900/40'}`}>
                    {positivo ? '+' : ''}{p.pnl_pct.toFixed(2)}%
                  </span>
                </td>
                <td className="px-4 py-3 text-center">
                  <span className="text-xs text-gray-400">{p.broker}</span>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
