import { CompraRequest, Cotizacion, Operacion, RendimientoReal, ResumenCartera, VentaRequest } from '../types'

const BASE = 'http://localhost:8080/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'Error desconocido')
  return data as T
}

export const api = {
  getCartera: () =>
    request<ResumenCartera>('/cartera'),

  postCompra: (body: CompraRequest) =>
    request<{ message: string }>('/compra', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  postVenta: (body: VentaRequest) =>
    request<{ message: string }>('/venta', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  getHistorial: (ticker?: string, limit = 100) => {
    const params = new URLSearchParams()
    if (ticker) params.set('ticker', ticker)
    params.set('limit', String(limit))
    return request<Operacion[]>(`/historial?${params}`)
  },

  getCotizacion: (ticker: string) =>
    request<Cotizacion>(`/cotizacion/${ticker}`),

  getRendimientoReal: () =>
    request<RendimientoReal>('/rendimiento/real'),
}
