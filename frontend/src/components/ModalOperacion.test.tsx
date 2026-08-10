import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import ModalOperacion from './ModalOperacion'
import { api } from '../hooks/useApi'

vi.mock('../hooks/useApi', () => ({
  api: {
    postCompra: vi.fn().mockResolvedValue({ message: 'ok' }),
    postVenta: vi.fn().mockResolvedValue({ message: 'ok' }),
  },
}))

describe('ModalOperacion — cálculo del total', () => {
  it('en una compra, el total muestra cantidad*precio + comisión', async () => {
    const user = userEvent.setup()
    render(<ModalOperacion tipo="compra" onClose={vi.fn()} onSuccess={vi.fn()} />)

    await user.type(screen.getByPlaceholderText('AAPL, GGAL...'), 'TSLA')
    await user.type(screen.getByPlaceholderText('0'), '10')
    await user.type(screen.getAllByPlaceholderText('0.00')[0], '1000')
    await user.type(screen.getAllByPlaceholderText('0.00')[1], '500')

    // 10 * 1000 + 500 comisión = 10500
    expect(screen.getByText(/ARS\s*10\.500/)).toBeInTheDocument()
    expect(screen.getByText(/Total a pagar/)).toBeInTheDocument()
  })

  it('en una venta, el total muestra cantidad*precio - comisión (neto)', async () => {
    const user = userEvent.setup()
    render(<ModalOperacion tipo="venta" onClose={vi.fn()} onSuccess={vi.fn()} />)

    await user.type(screen.getByPlaceholderText('AAPL, GGAL...'), 'TSLA')
    await user.type(screen.getByPlaceholderText('0'), '10')
    await user.type(screen.getAllByPlaceholderText('0.00')[0], '1000')
    await user.type(screen.getAllByPlaceholderText('0.00')[1], '500')

    // 10 * 1000 - 500 comisión = 9500
    expect(screen.getByText(/ARS\s*9\.500/)).toBeInTheDocument()
    expect(screen.getByText(/Total a cobrar/)).toBeInTheDocument()
  })

  it('no deja enviar sin ticker/cantidad/precio y muestra el error de validación', async () => {
    const user = userEvent.setup()
    const onSuccess = vi.fn()
    render(<ModalOperacion tipo="compra" onClose={vi.fn()} onSuccess={onSuccess} />)

    await user.click(screen.getByRole('button', { name: /Registrar Compra/ }))

    expect(await screen.findByText(/Completá ticker, cantidad y precio/)).toBeInTheDocument()
    expect(onSuccess).not.toHaveBeenCalled()
    expect(api.postCompra).not.toHaveBeenCalled()
  })

  it('envía fecha_opera y es_cedear en el payload de compra', async () => {
    const user = userEvent.setup()
    const onSuccess = vi.fn()
    render(<ModalOperacion tipo="compra" onClose={vi.fn()} onSuccess={onSuccess} />)

    await user.type(screen.getByPlaceholderText('AAPL, GGAL...'), 'ypf')
    await user.type(screen.getByPlaceholderText('0'), '5')
    await user.type(screen.getAllByPlaceholderText('0.00')[0], '2000')
    await user.click(screen.getByRole('button', { name: /Registrar Compra/ }))

    expect(api.postCompra).toHaveBeenCalledWith(
      expect.objectContaining({
        ticker: 'YPF', // se normaliza a mayúsculas
        cantidad: 5,
        precio_ars: 2000,
        fecha_opera: expect.stringMatching(/^\d{4}-\d{2}-\d{2}$/),
      })
    )
    expect(onSuccess).toHaveBeenCalled()
  })
})
