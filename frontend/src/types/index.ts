export interface Posicion {
  id: number
  ticker: string
  cantidad: number
  precio_prom_ars: number
  precio_actual_ars: number
  precio_actual_usd: number
  es_cedear: boolean
  broker: string
  pnl_ars: number
  pnl_pct: number
  valor_total_ars: number
}

export interface ResumenCartera {
  posiciones: Posicion[]
  total_invertido_ars: number
  total_actual_ars: number
  total_pnl_ars: number
  total_pnl_pct: number
  total_usd: number
  dolar_ccl: number
}

export interface Operacion {
  id: number
  tipo: 'COMPRA' | 'VENTA'
  ticker: string
  cantidad: number
  precio_ars: number
  total_ars: number
  es_cedear: boolean
  broker: string
  notas: string
  fecha_opera: string
}

export interface CompraRequest {
  ticker: string
  cantidad: number
  precio_ars: number
  es_cedear: boolean
  broker: string
  notas?: string
}

export interface VentaRequest {
  ticker: string
  cantidad: number
  precio_ars: number
  broker: string
  notas?: string
}

export interface Cotizacion {
  ticker: string
  precio_usd: number
  precio_ars: number
  dolar_ccl: number
}
