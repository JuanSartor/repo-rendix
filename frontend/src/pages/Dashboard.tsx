import { useEffect, useState, useCallback } from 'react'
import { api } from '../hooks/useApi'
import { ResumenCartera } from '../types'
import TarjetaResumen from '../components/TarjetaResumen'
import TablaCartera from '../components/TablaCartera'
import ModalOperacion from '../components/ModalOperacion'
import Historial from '../components/Historial'

type Vista = 'cartera' | 'historial'

export default function Dashboard() {
  const [cartera, setCartera] = useState<ResumenCartera | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [vista, setVista] = useState<Vista>('cartera')
  const [modalAbierto, setModalAbierto] = useState<'compra' | 'venta' | null>(null)

  const cargarCartera = useCallback(async () => {
    try {
      setLoading(true)
      setError('')
      const data = await api.getCartera()
      setCartera(data)
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    cargarCartera()
    // Refresh cada 60 segundos
    const interval = setInterval(cargarCartera, 60_000)
    return () => clearInterval(interval)
  }, [cargarCartera])

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="border-b border-gray-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-2xl">📊</span>
          <h1 className="text-xl font-bold tracking-tight">Portfolio Tracker</h1>
          {cartera && (
            <span className="text-xs text-gray-400 ml-2">
              CCL: <span className="text-green-400 font-mono">ARS {cartera.dolar_ccl.toLocaleString('es-AR', { minimumFractionDigits: 2 })}</span>
            </span>
          )}
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setModalAbierto('compra')}
            className="bg-green-600 hover:bg-green-500 px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            + Compra
          </button>
          <button
            onClick={() => setModalAbierto('venta')}
            className="bg-red-700 hover:bg-red-600 px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            − Venta
          </button>
          <button
            onClick={cargarCartera}
            className="bg-gray-800 hover:bg-gray-700 px-3 py-2 rounded-lg text-sm transition-colors"
            title="Actualizar"
          >
            🔄
          </button>
        </div>
      </header>

      {/* Nav tabs */}
      <nav className="border-b border-gray-800 px-6 flex gap-6">
        {(['cartera', 'historial'] as Vista[]).map((v) => (
          <button
            key={v}
            onClick={() => setVista(v)}
            className={`py-3 text-sm font-medium border-b-2 transition-colors capitalize ${
              vista === v
                ? 'border-blue-500 text-blue-400'
                : 'border-transparent text-gray-400 hover:text-white'
            }`}
          >
            {v === 'cartera' ? '💼 Cartera' : '📋 Historial'}
          </button>
        ))}
      </nav>

      {/* Content */}
      <main className="p-6">
        {loading && (
          <div className="flex justify-center py-20 text-gray-400">
            <span className="animate-pulse">Cargando cartera...</span>
          </div>
        )}

        {error && (
          <div className="bg-red-900/40 border border-red-700 rounded-lg p-4 text-red-300 text-sm">
            ⚠️ {error}
          </div>
        )}

        {!loading && !error && cartera && vista === 'cartera' && (
          <>
            {/* Tarjetas resumen */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
              <TarjetaResumen
                label="Invertido"
                valor={`ARS ${cartera.total_invertido_ars.toLocaleString('es-AR', { maximumFractionDigits: 0 })}`}
                sub=""
                color="neutral"
              />
              <TarjetaResumen
                label="Valor actual"
                valor={`ARS ${cartera.total_actual_ars.toLocaleString('es-AR', { maximumFractionDigits: 0 })}`}
                sub={`USD ${cartera.total_usd.toLocaleString('en-US', { maximumFractionDigits: 0 })}`}
                color="neutral"
              />
              <TarjetaResumen
                label="P&L Total"
                valor={`ARS ${cartera.total_pnl_ars.toLocaleString('es-AR', { maximumFractionDigits: 0 })}`}
                sub={`${cartera.total_pnl_pct >= 0 ? '+' : ''}${cartera.total_pnl_pct.toFixed(2)}%`}
                color={cartera.total_pnl_ars >= 0 ? 'green' : 'red'}
              />
              <TarjetaResumen
                label="Posiciones"
                valor={`${cartera.posiciones.length}`}
                sub="activas"
                color="neutral"
              />
            </div>

            {/* Tabla */}
            <TablaCartera posiciones={cartera.posiciones} />
          </>
        )}

        {!loading && vista === 'historial' && (
          <Historial />
        )}
      </main>

      {/* Modal compra/venta */}
      {modalAbierto && (
        <ModalOperacion
          tipo={modalAbierto}
          onClose={() => setModalAbierto(null)}
          onSuccess={() => {
            setModalAbierto(null)
            cargarCartera()
          }}
        />
      )}
    </div>
  )
}
