import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import TablaCartera from './TablaCartera'
import { Posicion } from '../types'

const posicionBase: Posicion = {
  id: 1,
  ticker: 'TSLA',
  cantidad: 10,
  precio_prom_ars: 1000,
  precio_disponible: true,
  precio_actual_ars: 1500,
  precio_actual_usd: 1,
  es_cedear: true,
  broker: 'Cocos',
  pnl_ars: 5000,
  pnl_pct: 50,
  valor_total_ars: 15000,
}

describe('TablaCartera — bug de -100% falso por fallo de cotización', () => {
  it('una posición con precio_disponible=false muestra "sin cotización", no -100%', () => {
    const sinPrecio: Posicion = {
      ...posicionBase,
      precio_disponible: false,
      precio_actual_ars: 0,
      pnl_ars: 0,
      pnl_pct: 0,
      valor_total_ars: 0,
    }
    render(<TablaCartera posiciones={[sinPrecio]} />)

    expect(screen.getByText(/sin cotización/i)).toBeInTheDocument()
    expect(screen.queryByText(/-100/)).not.toBeInTheDocument()
  })

  it('una posición con precio disponible muestra el P&L normalmente', () => {
    render(<TablaCartera posiciones={[posicionBase]} />)

    expect(screen.getByText('TSLA')).toBeInTheDocument()
    expect(screen.getByText(/\+50\.00%/)).toBeInTheDocument()
    expect(screen.queryByText(/sin cotización/i)).not.toBeInTheDocument()
  })

  it('sin posiciones, muestra el estado vacío en vez de una tabla vacía', () => {
    render(<TablaCartera posiciones={[]} />)
    expect(screen.getByText(/No tenés posiciones abiertas/)).toBeInTheDocument()
  })
})
