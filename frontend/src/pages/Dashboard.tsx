import { useEffect, useState, useCallback } from 'react'
import { api } from '../hooks/useApi'
import { useTheme } from '../hooks/useTheme'
import { ResumenCartera } from '../types'
import TarjetaResumen from '../components/TarjetaResumen'
import TablaCartera from '../components/TablaCartera'
import ModalOperacion from '../components/ModalOperacion'
import Historial from '../components/Historial'
import RetornoReal from '../components/RetornoReal'
import Alertas from '../components/Alertas'
import ImportarCSV from '../components/ImportarCSV'

type Vista = 'cartera' | 'retorno-real' | 'alertas' | 'historial'

export default function Dashboard() {
  const [cartera, setCartera] = useState<ResumenCartera | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [vista, setVista] = useState<Vista>('cartera')
  const [modalAbierto, setModalAbierto] = useState<'compra' | 'venta' | null>(null)
  const [importarAbierto, setImportarAbierto] = useState(false)
  const { tema, toggleTema } = useTheme()

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
    <div className="min-h-screen bg-gray-50 text-gray-900 dark:bg-gray-950 dark:text-white">
      {/* Header */}
      <header className="border-b border-gray-200 dark:border-gray-800 px-4 sm:px-6 py-3 sm:py-4 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 sm:gap-3 min-w-0">
          <span className="text-2xl shrink-0">📊</span>
          <h1 className="text-lg sm:text-xl font-bold tracking-tight truncate">Rendix</h1>
          {cartera && (
            <span className="hidden sm:inline text-xs text-gray-500 dark:text-gray-400 ml-2 whitespace-nowrap">
              CCL: <span className="text-green-600 dark:text-green-400 font-mono">ARS {cartera.dolar_ccl.toLocaleString('es-AR', { minimumFractionDigits: 2 })}</span>
            </span>
          )}
        </div>
        <div className="flex gap-2 flex-wrap">
          <button
            onClick={() => setModalAbierto('compra')}
            className="bg-green-600 hover:bg-green-500 text-white px-3 sm:px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            + Compra
          </button>
          <button
            onClick={() => setModalAbierto('venta')}
            className="bg-red-700 hover:bg-red-600 text-white px-3 sm:px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            − Venta
          </button>
          <button
            onClick={() => setImportarAbierto(true)}
            className="bg-gray-200 hover:bg-gray-300 text-gray-700 dark:bg-gray-800 dark:hover:bg-gray-700 dark:text-gray-200 px-3 sm:px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            📂 Importar CSV
          </button>
          <button
            onClick={cargarCartera}
            className="bg-gray-200 hover:bg-gray-300 text-gray-700 dark:bg-gray-800 dark:hover:bg-gray-700 dark:text-gray-200 px-3 py-2 rounded-lg text-sm transition-colors"
            title="Actualizar"
          >
            🔄
          </button>
          <button
            onClick={toggleTema}
            className="bg-gray-200 hover:bg-gray-300 text-gray-700 dark:bg-gray-800 dark:hover:bg-gray-700 dark:text-gray-200 px-3 py-2 rounded-lg text-sm transition-colors"
            title={tema === 'dark' ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'}
          >
            {tema === 'dark' ? '☀️' : '🌙'}
          </button>
        </div>
      </header>

      {/* CCL en mobile (no entra en el header en una sola línea) */}
      {cartera && (
        <div className="sm:hidden px-4 py-2 text-xs text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-800">
          CCL: <span className="text-green-600 dark:text-green-400 font-mono">ARS {cartera.dolar_ccl.toLocaleString('es-AR', { minimumFractionDigits: 2 })}</span>
        </div>
      )}

      {/* Nav tabs */}
      <nav className="border-b border-gray-200 dark:border-gray-800 px-4 sm:px-6 flex gap-4 sm:gap-6 overflow-x-auto">
        {(['cartera', 'retorno-real', 'alertas', 'historial'] as Vista[]).map((v) => (
          <button
            key={v}
            onClick={() => setVista(v)}
            className={`py-3 text-sm font-medium border-b-2 transition-colors capitalize whitespace-nowrap ${
              vista === v
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
            }`}
          >
            {v === 'cartera' ? '💼 Cartera' : v === 'retorno-real' ? '📈 Retorno Real' : v === 'alertas' ? '🔔 Alertas' : '📋 Historial'}
          </button>
        ))}
      </nav>

      {/* Content */}
      <main className="p-4 sm:p-6">
        {loading && (
          <div className="flex justify-center py-20 text-gray-500 dark:text-gray-400">
            <span className="animate-pulse">Cargando cartera...</span>
          </div>
        )}

        {error && (
          <div className="bg-red-100 border border-red-300 text-red-700 dark:bg-red-900/40 dark:border-red-700 dark:text-red-300 rounded-lg p-4 text-sm">
            ⚠️ {error}
          </div>
        )}

        {!loading && !error && cartera && vista === 'cartera' && (
          <>
            {cartera.posiciones_sin_cotizar && cartera.posiciones_sin_cotizar.length > 0 && (
              <div className="bg-yellow-100 border border-yellow-300 text-yellow-800 dark:bg-yellow-900/20 dark:border-yellow-700/50 dark:text-yellow-400 rounded-lg p-3 mb-4 text-sm">
                ⚠️ No se pudo obtener el precio actual de {cartera.posiciones_sin_cotizar.join(', ')} — los totales de abajo no incluyen esas posiciones. Probá actualizar en unos segundos.
              </div>
            )}

            {/* Tarjetas resumen */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 sm:gap-4 mb-6">
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

        {!loading && !error && vista === 'retorno-real' && (
          <RetornoReal />
        )}

        {!loading && vista === 'alertas' && (
          <Alertas />
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

      {/* Modal importar CSV */}
      {importarAbierto && (
        <ImportarCSV
          onClose={() => setImportarAbierto(false)}
          onSuccess={cargarCartera}
        />
      )}
    </div>
  )
}
