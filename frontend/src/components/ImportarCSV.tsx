import { useState } from 'react'
import { api } from '../hooks/useApi'
import { ImportarCSVResultado } from '../types'

interface Props {
  onClose: () => void
  onSuccess: () => void
}

const PLANTILLA = `tipo,ticker,cantidad,precio_ars,comision_ars,es_cedear,broker,fecha,notas
COMPRA,TSLA,10,45000,500,TRUE,Cocos,2026-01-15,Primer tramo DCA
COMPRA,YPF,100,4500,0,FALSE,Cocos,2026-02-10,
VENTA,TSLA,3,52000,300,,Cocos,2026-05-20,Toma de ganancia parcial
`

function descargarPlantilla() {
  const blob = new Blob([PLANTILLA], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'rendix-plantilla-import.csv'
  a.click()
  URL.revokeObjectURL(url)
}

export default function ImportarCSV({ onClose, onSuccess }: Props) {
  const [file, setFile] = useState<File | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [resultado, setResultado] = useState<ImportarCSVResultado | null>(null)

  const handleImportar = async () => {
    if (!file) {
      setError('Elegí un archivo CSV primero.')
      return
    }
    setLoading(true)
    setError('')
    setResultado(null)
    try {
      const res = await api.importarCSV(file)
      setResultado(res)
      if (res.importadas > 0) onSuccess()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 dark:bg-black/70 flex items-center justify-center z-50 p-4 overflow-y-auto">
      <div className="bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-700 rounded-2xl w-full max-w-lg shadow-2xl my-8">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between rounded-t-2xl bg-blue-50 dark:bg-blue-900/20">
          <h2 className="font-bold text-lg text-gray-900 dark:text-white">📂 Importar operaciones (CSV)</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-700 dark:hover:text-white text-xl">✕</button>
        </div>

        <div className="px-6 py-5 space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-300">
            Columnas requeridas: <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">tipo, ticker, cantidad, precio_ars, broker</code>.
            Opcionales: <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">comision_ars, es_cedear, fecha, notas</code>.
            El orden de filas no importa, se reordenan por fecha automáticamente.
          </p>

          <button
            onClick={descargarPlantilla}
            className="text-sm text-blue-600 dark:text-blue-400 hover:underline"
          >
            ⬇️ Descargar plantilla de ejemplo
          </button>

          <div>
            <label className="text-xs text-gray-500 dark:text-gray-400 mb-1 block">Archivo CSV</label>
            <input
              type="file"
              accept=".csv"
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
              className="w-full text-sm text-gray-600 dark:text-gray-300 file:mr-3 file:py-2 file:px-3 file:rounded-lg file:border-0 file:bg-blue-600 file:text-white file:text-sm hover:file:bg-blue-500 file:cursor-pointer cursor-pointer"
            />
          </div>

          {error && (
            <p className="text-red-700 dark:text-red-400 text-sm bg-red-50 border border-red-200 dark:bg-red-900/20 dark:border-red-800 rounded-lg px-3 py-2">
              ⚠️ {error}
            </p>
          )}

          {resultado && (
            <div className="bg-gray-100 dark:bg-gray-800 rounded-lg px-4 py-3 text-sm">
              <p className="text-green-700 dark:text-green-400 font-medium">
                ✅ {resultado.importadas} operaci{resultado.importadas === 1 ? 'ón' : 'ones'} importada{resultado.importadas === 1 ? '' : 's'}
              </p>
              {resultado.errores && resultado.errores.length > 0 && (
                <div className="mt-2">
                  <p className="text-yellow-700 dark:text-yellow-400">⚠️ {resultado.errores.length} fila(s) con errores:</p>
                  <ul className="list-disc list-inside text-xs text-gray-500 dark:text-gray-400 mt-1 space-y-0.5">
                    {resultado.errores.map((e, i) => (
                      <li key={i}>{e}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>

        <div className="px-6 pb-5 flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 py-2.5 rounded-xl border border-gray-300 text-gray-600 hover:bg-gray-100 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800 transition-colors text-sm"
          >
            Cerrar
          </button>
          <button
            onClick={handleImportar}
            disabled={loading || !file}
            className="flex-1 py-2.5 rounded-xl font-medium text-sm transition-colors bg-blue-600 hover:bg-blue-500 text-white disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Importando...' : 'Importar'}
          </button>
        </div>
      </div>
    </div>
  )
}
