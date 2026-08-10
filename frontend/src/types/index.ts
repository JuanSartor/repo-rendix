export interface Posicion {
  id: number
  ticker: string
  cantidad: number
  precio_prom_ars: number
  precio_disponible: boolean
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
  posiciones_sin_cotizar?: string[]
}

export interface Operacion {
  id: number
  tipo: 'COMPRA' | 'VENTA'
  ticker: string
  cantidad: number
  precio_ars: number
  comision_ars: number
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
  comision_ars?: number
  es_cedear: boolean
  broker: string
  notas?: string
}

export interface VentaRequest {
  ticker: string
  cantidad: number
  precio_ars: number
  comision_ars?: number
  broker: string
  notas?: string
}

export interface RendimientoPosicion {
  ticker: string
  invertido_ars: number
  valor_actual_ars: number
  inflacion_ars_pct: number
  retorno_nominal_ars: number
  retorno_nominal_pct: number
  retorno_real_ars: number
  retorno_real_pct: number
  invertido_usd?: number
  valor_actual_usd?: number
  inflacion_usa_pct?: number
  retorno_real_usd?: number
  retorno_real_usd_pct?: number
}

export interface RendimientoReal {
  posiciones: RendimientoPosicion[]
  total_invertido_ars: number
  total_actual_ars: number
  total_inflacion_ars_pct: number
  total_retorno_nominal_ars: number
  total_retorno_nominal_pct: number
  total_retorno_real_ars: number
  total_retorno_real_pct: number
  ipc_fecha_actual: string
  resumen: string
  posiciones_sin_cotizar?: string[]
}

export interface Cotizacion {
  ticker: string
  precio_usd: number
  precio_ars: number
  dolar_ccl: number
}

export interface Alerta {
  id: number
  ticker: string
  precio_objetivo: number
  direccion: 'ARRIBA' | 'ABAJO'
  es_cedear: boolean
  activa: boolean
  creado_en: string
  disparada_en?: string
}

export interface AlertaRequest {
  ticker: string
  precio_objetivo: number
  direccion: 'ARRIBA' | 'ABAJO'
  es_cedear: boolean
}
