import { useEffect, useState } from 'react'
import { api } from '../hooks/useApi'
import { Operacion } from '../types'

export default function Historial() {
  const [ops, setOps] = useState<Operacion[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filtroTicker, setFiltroTicker] = useState('')

  const cargar = async (ticker?: string) => {
    try {
      setLoading(true)
      const data = await api.getHistorial(ticker)
      setOps(data || [])
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { cargar() }, [])

  const handleBuscar = () => cargar(filtroTicker || undefined)

  return (
    <div>
      {/* Filtro */}
      <div className="flex gap-3 mb-4">
        <input
          value={filtroTicker}
          onChange={(e) => setFiltroTicker(e.target.value.toUpperCase())}
          placeholder="Filtrar por ticker (AAPL, GGAL...)"
          className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500 w-64 font-mono"
        />
        <button
          onClick={handleBuscar}
          className="bg-gray-700 hover:bg-gray-600 px-4 py-2 rounded-lg text-sm transition-colors"
        >
          Filtrar
        </button>
        {filtroTicker && (
          <button
            onClick={() => { setFiltroTicker(''); cargar() }}
            className="text-gray-400 hover:text-white text-sm"
          >
            Limpiar
          </button>
        )}
      </div>

      {loading && <p className="text-gray-400 text-sm animate-pulse">Cargando historial...</p>}
      {error && <p className="text-red-400 text-sm">{error}</p>}

      {!loading && ops.length === 0 && (
        <div className="text-center py-16 text-gray-500">
          <p className="text-4xl mb-3">📋</p>
          <p>Sin operaciones registradas.</p>
        </div>
      )}

      {!loading && ops.length > 0 && (
        <div className="overflow-x-auto rounded-xl border border-gray-800">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-900 text-gray-400 text-xs uppercase tracking-wide">
                <th className="px-4 py-3 text-left">#</th>
                <th className="px-4 py-3 text-left">Fecha</th>
                <th className="px-4 py-3 text-left">Tipo</th>
                <th className="px-4 py-3 text-left">Ticker</th>
                <th className="px-4 py-3 text-right">Cantidad</th>
                <th className="px-4 py-3 text-right">Precio ARS</th>
                <th className="px-4 py-3 text-right">Comisión</th>
                <th className="px-4 py-3 text-right">Total ARS</th>
                <th className="px-4 py-3 text-center">Broker</th>
                <th className="px-4 py-3 text-left">Notas</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800">
              {ops.map((op) => (
                <tr key={op.id} className="bg-gray-950 hover:bg-gray-900 transition-colors">
                  <td className="px-4 py-3 text-gray-500 font-mono text-xs">{op.id}</td>
                  <td className="px-4 py-3 text-gray-400 text-xs">
                    {new Date(op.fecha_opera).toLocaleDateString('es-AR')}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`text-xs font-medium px-2 py-0.5 rounded ${
                      op.tipo === 'COMPRA'
                        ? 'bg-green-900/40 text-green-400'
                        : 'bg-red-900/40 text-red-400'
                    }`}>
                      {op.tipo}
                    </span>
                  </td>
                  <td className="px-4 py-3 font-bold text-white font-mono">
                    {op.ticker}
                    {op.es_cedear && (
                      <span className="ml-1 text-xs bg-blue-900 text-blue-300 px-1 py-0.5 rounded">C</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-right font-mono text-gray-300">
                    {op.cantidad.toLocaleString('es-AR', { maximumFractionDigits: 2 })}
                  </td>
                  <td className="px-4 py-3 text-right font-mono text-gray-300">
                    {op.precio_ars.toLocaleString('es-AR', { maximumFractionDigits: 2 })}
                  </td>
                  <td className="px-4 py-3 text-right font-mono text-gray-500 text-xs">
                    {op.comision_ars > 0 ? op.comision_ars.toLocaleString('es-AR', { maximumFractionDigits: 0 }) : '—'}
                  </td>
                  <td className="px-4 py-3 text-right font-mono text-white font-medium">
                    {op.total_ars.toLocaleString('es-AR', { maximumFractionDigits: 0 })}
                  </td>
                  <td className="px-4 py-3 text-center text-xs text-gray-400">{op.broker}</td>
                  <td className="px-4 py-3 text-xs text-gray-500 italic">{op.notas}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
