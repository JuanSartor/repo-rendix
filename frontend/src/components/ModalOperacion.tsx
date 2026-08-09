import { useState } from 'react'
import { api } from '../hooks/useApi'

interface Props {
  tipo: 'compra' | 'venta'
  onClose: () => void
  onSuccess: () => void
}

const BROKERS = ['Cocos', 'Bull Market']

export default function ModalOperacion({ tipo, onClose, onSuccess }: Props) {
  const [ticker, setTicker] = useState('')
  const [cantidad, setCantidad] = useState('')
  const [precioARS, setPrecioARS] = useState('')
  const [esCedear, setEsCedear] = useState(true)
  const [broker, setBroker] = useState('Cocos')
  const [notas, setNotas] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const esCompra = tipo === 'compra'
  const total = (parseFloat(cantidad) || 0) * (parseFloat(precioARS) || 0)

  const handleSubmit = async () => {
    if (!ticker || !cantidad || !precioARS) {
      setError('Completá ticker, cantidad y precio.')
      return
    }
    setLoading(true)
    setError('')
    try {
      if (esCompra) {
        await api.postCompra({
          ticker: ticker.toUpperCase(),
          cantidad: parseFloat(cantidad),
          precio_ars: parseFloat(precioARS),
          es_cedear: esCedear,
          broker,
          notas,
        })
      } else {
        await api.postVenta({
          ticker: ticker.toUpperCase(),
          cantidad: parseFloat(cantidad),
          precio_ars: parseFloat(precioARS),
          broker,
          notas,
        })
      }
      onSuccess()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-900 border border-gray-700 rounded-2xl w-full max-w-md shadow-2xl">
        {/* Header */}
        <div className={`px-6 py-4 border-b border-gray-800 flex items-center justify-between rounded-t-2xl ${
          esCompra ? 'bg-green-900/20' : 'bg-red-900/20'
        }`}>
          <h2 className="font-bold text-lg">
            {esCompra ? '🟢 Registrar Compra' : '🔴 Registrar Venta'}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white text-xl">✕</button>
        </div>

        {/* Form */}
        <div className="px-6 py-5 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-xs text-gray-400 mb-1 block">Ticker *</label>
              <input
                value={ticker}
                onChange={(e) => setTicker(e.target.value.toUpperCase())}
                placeholder="AAPL, GGAL..."
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500 font-mono uppercase"
              />
            </div>
            <div>
              <label className="text-xs text-gray-400 mb-1 block">Cantidad *</label>
              <input
                type="number"
                value={cantidad}
                onChange={(e) => setCantidad(e.target.value)}
                placeholder="0"
                min="0"
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
          </div>

          <div>
            <label className="text-xs text-gray-400 mb-1 block">Precio ARS por unidad *</label>
            <input
              type="number"
              value={precioARS}
              onChange={(e) => setPrecioARS(e.target.value)}
              placeholder="0.00"
              min="0"
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>

          {total > 0 && (
            <div className="bg-gray-800 rounded-lg px-4 py-2 text-sm text-gray-300">
              Total operación:{' '}
              <span className="text-white font-mono font-bold">
                ARS {total.toLocaleString('es-AR', { maximumFractionDigits: 2 })}
              </span>
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-xs text-gray-400 mb-1 block">Broker</label>
              <select
                value={broker}
                onChange={(e) => setBroker(e.target.value)}
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
              >
                {BROKERS.map((b) => (
                  <option key={b}>{b}</option>
                ))}
              </select>
            </div>
            {esCompra && (
              <div className="flex items-end pb-2">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={esCedear}
                    onChange={(e) => setEsCedear(e.target.checked)}
                    className="w-4 h-4 accent-blue-500"
                  />
                  <span className="text-sm text-gray-300">Es CEDEAR</span>
                </label>
              </div>
            )}
          </div>

          <div>
            <label className="text-xs text-gray-400 mb-1 block">Notas (opcional)</label>
            <input
              value={notas}
              onChange={(e) => setNotas(e.target.value)}
              placeholder="DCA, oportunidad, etc."
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
            />
          </div>

          {error && (
            <p className="text-red-400 text-sm bg-red-900/20 border border-red-800 rounded-lg px-3 py-2">
              ⚠️ {error}
            </p>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 pb-5 flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 py-2.5 rounded-xl border border-gray-700 text-gray-300 hover:bg-gray-800 transition-colors text-sm"
          >
            Cancelar
          </button>
          <button
            onClick={handleSubmit}
            disabled={loading}
            className={`flex-1 py-2.5 rounded-xl font-medium text-sm transition-colors ${
              esCompra
                ? 'bg-green-600 hover:bg-green-500 text-white'
                : 'bg-red-700 hover:bg-red-600 text-white'
            } disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {loading ? 'Guardando...' : esCompra ? 'Registrar Compra' : 'Registrar Venta'}
          </button>
        </div>
      </div>
    </div>
  )
}
