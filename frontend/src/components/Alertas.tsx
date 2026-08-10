import { useEffect, useState } from 'react'
import { api } from '../hooks/useApi'
import { Alerta } from '../types'

export default function Alertas() {
  const [alertas, setAlertas] = useState<Alerta[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [ticker, setTicker] = useState('')
  const [precioObjetivo, setPrecioObjetivo] = useState('')
  const [direccion, setDireccion] = useState<'ARRIBA' | 'ABAJO'>('ARRIBA')
  const [esCedear, setEsCedear] = useState(true)
  const [guardando, setGuardando] = useState(false)
  const [formError, setFormError] = useState('')

  const cargar = async () => {
    try {
      setLoading(true)
      const data = await api.getAlertas()
      setAlertas(data)
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { cargar() }, [])

  const handleCrear = async () => {
    if (!ticker || !precioObjetivo) {
      setFormError('Completá ticker y precio objetivo.')
      return
    }
    setGuardando(true)
    setFormError('')
    try {
      await api.postAlerta({
        ticker: ticker.toUpperCase(),
        precio_objetivo: parseFloat(precioObjetivo),
        direccion,
        es_cedear: esCedear,
      })
      setTicker('')
      setPrecioObjetivo('')
      await cargar()
    } catch (e: any) {
      setFormError(e.message)
    } finally {
      setGuardando(false)
    }
  }

  const handleEliminar = async (id: number) => {
    try {
      await api.deleteAlerta(id)
      await cargar()
    } catch (e: any) {
      setError(e.message)
    }
  }

  const activas = alertas.filter((a) => a.activa)
  const disparadas = alertas.filter((a) => !a.activa)

  const inputClass =
    'w-full bg-white border border-gray-300 text-gray-900 dark:bg-gray-800 dark:border-gray-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-blue-500 font-mono'
  const labelClass = 'text-xs text-gray-500 dark:text-gray-400 mb-1 block'

  return (
    <div>
      {/* Formulario para crear alerta */}
      <div className="bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-800 rounded-xl p-5 mb-6">
        <h3 className="font-bold text-gray-900 dark:text-white mb-4">🔔 Nueva alerta de precio</h3>
        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4 mb-4">
          <div>
            <label className={labelClass}>Ticker *</label>
            <input
              value={ticker}
              onChange={(e) => setTicker(e.target.value.toUpperCase())}
              placeholder="TSLA, YPF..."
              className={`${inputClass} uppercase`}
            />
          </div>
          <div>
            <label className={labelClass}>Precio objetivo ARS *</label>
            <input
              type="number"
              value={precioObjetivo}
              onChange={(e) => setPrecioObjetivo(e.target.value)}
              placeholder="0.00"
              min="0"
              className={inputClass}
            />
          </div>
          <div>
            <label className={labelClass}>Dirección</label>
            <select value={direccion} onChange={(e) => setDireccion(e.target.value as 'ARRIBA' | 'ABAJO')} className={inputClass}>
              <option value="ARRIBA">Cuando suba a...</option>
              <option value="ABAJO">Cuando baje a...</option>
            </select>
          </div>
          <div className="flex items-end pb-2">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={esCedear}
                onChange={(e) => setEsCedear(e.target.checked)}
                className="w-4 h-4 accent-blue-500"
              />
              <span className="text-sm text-gray-600 dark:text-gray-300">Es CEDEAR</span>
            </label>
          </div>
        </div>

        {formError && (
          <p className="text-red-700 dark:text-red-400 text-sm bg-red-50 border border-red-200 dark:bg-red-900/20 dark:border-red-800 rounded-lg px-3 py-2 mb-4">
            ⚠️ {formError}
          </p>
        )}

        <button
          onClick={handleCrear}
          disabled={guardando}
          className="bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
        >
          {guardando ? 'Creando...' : '+ Crear alerta'}
        </button>
      </div>

      {loading && <p className="text-gray-500 dark:text-gray-400 text-sm animate-pulse">Cargando alertas...</p>}
      {error && <p className="text-red-600 dark:text-red-400 text-sm">{error}</p>}

      {!loading && alertas.length === 0 && (
        <div className="text-center py-16 text-gray-500">
          <p className="text-4xl mb-3">🔕</p>
          <p>No tenés alertas configuradas.</p>
        </div>
      )}

      {!loading && activas.length > 0 && (
        <div className="mb-6">
          <h4 className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-2">Activas</h4>
          <div className="space-y-2">
            {activas.map((a) => (
              <div key={a.id} className="flex items-center justify-between bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-800 rounded-lg px-4 py-3">
                <div className="flex items-center gap-3">
                  <span className="font-bold text-gray-900 dark:text-white">{a.ticker}</span>
                  {a.es_cedear && <span className="text-xs bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300 px-1.5 py-0.5 rounded">CEDEAR</span>}
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    {a.direccion === 'ARRIBA' ? '📈 sube a' : '📉 baja a'} ARS {a.precio_objetivo.toLocaleString('es-AR', { maximumFractionDigits: 2 })}
                  </span>
                </div>
                <button
                  onClick={() => handleEliminar(a.id)}
                  className="text-xs text-gray-400 hover:text-red-500 transition-colors"
                >
                  Cancelar
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {!loading && disparadas.length > 0 && (
        <div>
          <h4 className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-2">Disparadas</h4>
          <div className="space-y-2">
            {disparadas.map((a) => (
              <div key={a.id} className="flex items-center justify-between bg-gray-50 border border-gray-200 dark:bg-gray-950 dark:border-gray-800 rounded-lg px-4 py-3 opacity-70">
                <div className="flex items-center gap-3">
                  <span className="font-bold text-gray-700 dark:text-gray-300">{a.ticker}</span>
                  <span className="text-sm text-gray-500 dark:text-gray-500">
                    {a.direccion === 'ARRIBA' ? '📈 subió a' : '📉 bajó a'} ARS {a.precio_objetivo.toLocaleString('es-AR', { maximumFractionDigits: 2 })}
                  </span>
                  {a.disparada_en && (
                    <span className="text-xs text-gray-400 dark:text-gray-600">
                      {new Date(a.disparada_en).toLocaleDateString('es-AR')}
                    </span>
                  )}
                </div>
                <button
                  onClick={() => handleEliminar(a.id)}
                  className="text-xs text-gray-400 hover:text-red-500 transition-colors"
                >
                  Eliminar
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
