import { useState } from 'react'
import { api } from '../hooks/useApi'

interface Props {
  tipo: 'compra' | 'venta'
  onClose: () => void
  onSuccess: () => void
}

const BROKERS = ['Cocos', 'Bull Market']

function hoyISO() {
  return new Date().toISOString().slice(0, 10)
}

export default function ModalOperacion({ tipo, onClose, onSuccess }: Props) {
  const [ticker, setTicker] = useState('')
  const [cantidad, setCantidad] = useState('')
  const [precioARS, setPrecioARS] = useState('')
  const [comisionARS, setComisionARS] = useState('')
  const [esCedear, setEsCedear] = useState(true)
  const [broker, setBroker] = useState('Cocos')
  const [notas, setNotas] = useState('')
  const [fecha, setFecha] = useState(hoyISO)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const esCompra = tipo === 'compra'
  const comision = parseFloat(comisionARS) || 0
  const subtotal = (parseFloat(cantidad) || 0) * (parseFloat(precioARS) || 0)
  const total = esCompra ? subtotal + comision : subtotal - comision

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
          comision_ars: comision,
          es_cedear: esCedear,
          broker,
          notas,
          fecha_opera: fecha,
        })
      } else {
        await api.postVenta({
          ticker: ticker.toUpperCase(),
          cantidad: parseFloat(cantidad),
          precio_ars: parseFloat(precioARS),
          comision_ars: comision,
          broker,
          notas,
          fecha_opera: fecha,
        })
      }
      onSuccess()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  const inputClass =
    'w-full bg-white border border-gray-300 text-gray-900 dark:bg-gray-800 dark:border-gray-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-blue-500 font-mono'
  const labelClass = 'text-xs text-gray-500 dark:text-gray-400 mb-1 block'

  return (
    <div className="fixed inset-0 bg-black/50 dark:bg-black/70 flex items-center justify-center z-50 p-4 overflow-y-auto">
      <div className="bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-700 rounded-2xl w-full max-w-md shadow-2xl my-8">
        {/* Header */}
        <div className={`px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between rounded-t-2xl ${
          esCompra ? 'bg-green-50 dark:bg-green-900/20' : 'bg-red-50 dark:bg-red-900/20'
        }`}>
          <h2 className="font-bold text-lg text-gray-900 dark:text-white">
            {esCompra ? '🟢 Registrar Compra' : '🔴 Registrar Venta'}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-700 dark:hover:text-white text-xl">✕</button>
        </div>

        {/* Form */}
        <div className="px-6 py-5 space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Ticker *</label>
              <input
                value={ticker}
                onChange={(e) => setTicker(e.target.value.toUpperCase())}
                placeholder="AAPL, GGAL..."
                className={`${inputClass} uppercase`}
              />
            </div>
            <div>
              <label className={labelClass}>Cantidad *</label>
              <input
                type="number"
                value={cantidad}
                onChange={(e) => setCantidad(e.target.value)}
                placeholder="0"
                min="0"
                className={inputClass}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Precio ARS por unidad *</label>
              <input
                type="number"
                value={precioARS}
                onChange={(e) => setPrecioARS(e.target.value)}
                placeholder="0.00"
                min="0"
                className={inputClass}
              />
            </div>
            <div>
              <label className={labelClass}>Comisión ARS (opcional)</label>
              <input
                type="number"
                value={comisionARS}
                onChange={(e) => setComisionARS(e.target.value)}
                placeholder="0.00"
                min="0"
                className={inputClass}
              />
            </div>
          </div>

          {total > 0 && (
            <div className="bg-gray-100 dark:bg-gray-800 rounded-lg px-4 py-2 text-sm text-gray-600 dark:text-gray-300">
              {esCompra ? 'Total a pagar (con comisión)' : 'Total a cobrar (neto de comisión)'}:{' '}
              <span className="text-gray-900 dark:text-white font-mono font-bold">
                ARS {total.toLocaleString('es-AR', { maximumFractionDigits: 2 })}
              </span>
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Fecha</label>
              <input
                type="date"
                value={fecha}
                onChange={(e) => setFecha(e.target.value)}
                max={hoyISO()}
                className={inputClass}
              />
            </div>
            <div>
              <label className={labelClass}>Broker</label>
              <select
                value={broker}
                onChange={(e) => setBroker(e.target.value)}
                className={inputClass}
              >
                {BROKERS.map((b) => (
                  <option key={b}>{b}</option>
                ))}
              </select>
            </div>
          </div>

          {esCompra && fecha !== hoyISO() && (
            <p className="text-xs text-gray-400 dark:text-gray-500 -mt-2">
              Fecha retroactiva: el retorno real en USD para esta posición no va a estar disponible (no tenemos el dólar CCL histórico de esa fecha). El retorno real en ARS sí se calcula bien.
            </p>
          )}

          {esCompra && (
            <label className="flex items-center gap-2 cursor-pointer w-fit">
              <input
                type="checkbox"
                checked={esCedear}
                onChange={(e) => setEsCedear(e.target.checked)}
                className="w-4 h-4 accent-blue-500"
              />
              <span className="text-sm text-gray-600 dark:text-gray-300">Es CEDEAR</span>
            </label>
          )}

          <div>
            <label className={labelClass}>Notas (opcional)</label>
            <input
              value={notas}
              onChange={(e) => setNotas(e.target.value)}
              placeholder="DCA, oportunidad, etc."
              className={inputClass}
            />
          </div>

          {error && (
            <p className="text-red-700 dark:text-red-400 text-sm bg-red-50 border border-red-200 dark:bg-red-900/20 dark:border-red-800 rounded-lg px-3 py-2">
              ⚠️ {error}
            </p>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 pb-5 flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 py-2.5 rounded-xl border border-gray-300 text-gray-600 hover:bg-gray-100 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800 transition-colors text-sm"
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
