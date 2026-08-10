package services

import (
	"sync"

	"portfolio-tracker/internal/models"
)

// TotalesCartera son los agregados calculados junto con los precios actuales de cada posición.
type TotalesCartera struct {
	TotalInvertido float64
	TotalActual    float64
	// SinCotizar lista los tickers cuyo precio no se pudo obtener; sus montos no
	// entran en TotalInvertido/TotalActual porque no sabemos su valor real hoy.
	SinCotizar []string
}

// CalcularPreciosActuales completa PrecioActualARS/USD, PnlARS/Pct, ValorTotalARS y
// PrecioDisponible de cada posición (mutando el slice in-place, una goroutine por
// posición ya que cada fetch a Yahoo tarda 200-500ms) y devuelve los totales agregados.
// Vive en services (no en handlers) porque también la usan los jobs en background
// (resumen diario por Telegram), que no pasan por HTTP.
//
// Si el fetch de una posición falla, PrecioDisponible queda en false y esa posición
// se excluye de los totales — nunca mostramos un P&L de -100% por un error de red.
func CalcularPreciosActuales(posiciones []models.Posicion, ccl float64) TotalesCartera {
	var wg sync.WaitGroup
	for i := range posiciones {
		wg.Add(1)
		go func(p *models.Posicion) {
			defer wg.Done()

			var precioActualARS float64
			var err error
			if p.EsCedear {
				_, precioActualARS, err = ObtenerPrecioARS(p.Ticker, ccl)
			} else {
				// Acciones locales: sufijo .BA en Yahoo Finance
				tickerBA := ResolverTickerLocal(p.Ticker)
				precioActualARS, err = ObtenerPrecioUSD(tickerBA + ".BA")
			}
			if err != nil || precioActualARS <= 0 {
				p.PrecioDisponible = false
				return
			}

			invertido := p.Cantidad * p.PrecioPromARS
			actualTotal := p.Cantidad * precioActualARS

			p.PrecioDisponible = true
			p.PrecioActualARS = precioActualARS
			p.ValorTotalARS = actualTotal
			p.PnlARS = actualTotal - invertido
			if invertido > 0 {
				p.PnlPct = (p.PnlARS / invertido) * 100
			}
			if ccl > 0 {
				p.PrecioActualUSD = precioActualARS / ccl
			}
		}(&posiciones[i])
	}
	wg.Wait()

	var totales TotalesCartera
	for i := range posiciones {
		p := &posiciones[i]
		if !p.PrecioDisponible {
			totales.SinCotizar = append(totales.SinCotizar, p.Ticker)
			continue
		}
		totales.TotalInvertido += p.Cantidad * p.PrecioPromARS
		totales.TotalActual += p.ValorTotalARS
	}

	return totales
}
