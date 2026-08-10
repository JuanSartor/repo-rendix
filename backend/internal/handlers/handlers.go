package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"portfolio-tracker/internal/db"
	"portfolio-tracker/internal/models"
	"portfolio-tracker/internal/services"
)

// totalesCartera son los agregados calculados junto con los precios actuales de cada posición.
type totalesCartera struct {
	TotalInvertido float64
	TotalActual    float64
	// SinCotizar lista los tickers cuyo precio no se pudo obtener; sus montos no
	// entran en TotalInvertido/TotalActual porque no sabemos su valor real hoy.
	SinCotizar []string
}

// calcularPreciosActuales completa PrecioActualARS/USD, PnlARS/Pct, ValorTotalARS y
// PrecioDisponible de cada posición (mutando el slice in-place, una goroutine por
// posición ya que cada fetch a Yahoo tarda 200-500ms) y devuelve los totales agregados.
// Compartido entre GetCartera y GetRendimientoReal para no duplicar las llamadas.
//
// Si el fetch de una posición falla, PrecioDisponible queda en false y esa posición
// se excluye de los totales — nunca mostramos un P&L de -100% por un error de red.
func calcularPreciosActuales(posiciones []models.Posicion, ccl float64) totalesCartera {
	var wg sync.WaitGroup
	for i := range posiciones {
		wg.Add(1)
		go func(p *models.Posicion) {
			defer wg.Done()

			var precioActualARS float64
			var err error
			if p.EsCedear {
				_, precioActualARS, err = services.ObtenerPrecioARS(p.Ticker, ccl)
			} else {
				// Acciones locales: sufijo .BA en Yahoo Finance
				tickerBA := services.ResolverTickerLocal(p.Ticker)
				precioActualARS, err = services.ObtenerPrecioUSD(tickerBA + ".BA")
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

	var totales totalesCartera
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

// GET /api/cartera
// Devuelve todas las posiciones con precios actualizados y P&L calculado
func GetCartera(c *gin.Context) {
	posiciones, err := db.ObtenerPosiciones()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if posiciones == nil {
		posiciones = []models.Posicion{}
	}

	ccl, err := services.ObtenerDolarCCL()
	if err != nil {
		// Continuamos sin CCL, frontend lo maneja
		ccl = 0
	}

	totales := calcularPreciosActuales(posiciones, ccl)

	totalPnl := totales.TotalActual - totales.TotalInvertido
	totalPct := 0.0
	if totales.TotalInvertido > 0 {
		totalPct = (totalPnl / totales.TotalInvertido) * 100
	}
	totalUSD := 0.0
	if ccl > 0 {
		totalUSD = totales.TotalActual / ccl
	}

	c.JSON(http.StatusOK, models.ResumenCartera{
		Posiciones:           posiciones,
		TotalInvertido:       totales.TotalInvertido,
		TotalActual:          totales.TotalActual,
		TotalPnlARS:          totalPnl,
		TotalPnlPct:          totalPct,
		TotalUSD:             totalUSD,
		DolarCCL:             ccl,
		PosicionesSinCotizar: totales.SinCotizar,
	})
}

// GET /api/rendimiento/real
// Compara el retorno nominal contra el retorno ajustado por inflación (ARS: INDEC, USD: FRED).
func GetRendimientoReal(c *gin.Context) {
	posiciones, err := db.ObtenerPosiciones()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(posiciones) == 0 {
		c.JSON(http.StatusOK, models.RendimientoReal{Posiciones: []models.RendimientoPosicion{}, Resumen: "No tenés posiciones abiertas."})
		return
	}

	ccl, _ := services.ObtenerDolarCCL()
	totalesPrecios := calcularPreciosActuales(posiciones, ccl)

	serieIPC, err := services.ObtenerIPCArgentina()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "no se pudo obtener inflación INDEC: " + err.Error()})
		return
	}
	ipcActual := services.UltimoIndice(serieIPC)

	// El CPI de EEUU es opcional (requiere FRED_API_KEY) — si falla, seguimos solo con ARS.
	serieCPI, errCPI := services.ObtenerCPIUSA()
	var cpiActual services.IndicePrecio
	if errCPI == nil {
		cpiActual = services.UltimoIndice(serieCPI)
	}

	resultado := models.RendimientoReal{
		Posiciones:           make([]models.RendimientoPosicion, 0, len(posiciones)),
		IpcFechaActual:       ipcActual.Fecha.Format("2006-01"),
		PosicionesSinCotizar: totalesPrecios.SinCotizar,
	}

	var totalInvertidoAjustadoARS float64

	for _, p := range posiciones {
		if !p.PrecioDisponible {
			// Sin precio actual no podemos calcular retorno de ningún tipo para esta
			// posición; la excluimos en vez de mostrar un número inventado.
			continue
		}
		invertido := p.Cantidad * p.PrecioPromARS

		// Ajustamos la inflación tramo por tramo (cada COMPRA con su propia fecha),
		// ponderado por lo invertido en cada tramo — así una posición armada con DCA
		// en varios meses no queda anclada a la inflación medida desde la primera
		// compra únicamente. El ratio resultante se aplica al invertido ACTUAL de la
		// posición (que ya neteó ventas parciales vía el precio promedio ponderado).
		ratioAjuste := 1.0
		if compras, errCompras := db.ObtenerComprasPorTicker(p.Ticker); errCompras == nil && len(compras) > 0 {
			var nominalTramos, ajustadoTramos float64
			for _, op := range compras {
				montoTramo := op.Cantidad*op.PrecioARS + op.ComisionARS
				ipcTramo, _ := services.BuscarIndiceEnFecha(serieIPC, op.FechaOpera)
				nominalTramos += montoTramo
				ajustadoTramos += montoTramo * (ipcActual.Valor / ipcTramo)
			}
			if nominalTramos > 0 {
				ratioAjuste = ajustadoTramos / nominalTramos
			}
		} else {
			// Fallback defensivo: sin historial de compras, usamos la fecha de apertura.
			ipcInicio, _ := services.BuscarIndiceEnFecha(serieIPC, p.CreadoEn)
			ratioAjuste = ipcActual.Valor / ipcInicio
		}

		inflacionPct := (ratioAjuste - 1) * 100
		invertidoAjustado := invertido * ratioAjuste
		totalInvertidoAjustadoARS += invertidoAjustado

		retornoNominal := p.ValorTotalARS - invertido
		retornoNominalPct := 0.0
		if invertido > 0 {
			retornoNominalPct = (retornoNominal / invertido) * 100
		}
		retornoReal := p.ValorTotalARS - invertidoAjustado
		retornoRealPct := 0.0
		if invertidoAjustado > 0 {
			retornoRealPct = (retornoReal / invertidoAjustado) * 100
		}

		rp := models.RendimientoPosicion{
			Ticker:            p.Ticker,
			InvertidoARS:      invertido,
			ValorActualARS:    p.ValorTotalARS,
			InflacionArsPct:   inflacionPct,
			RetornoNominalARS: retornoNominal,
			RetornoNominalPct: retornoNominalPct,
			RetornoRealARS:    retornoReal,
			RetornoRealPct:    retornoRealPct,
		}

		// USD real: solo si la posición tiene CCL de apertura y FRED respondió.
		if p.CCLApertura != nil && *p.CCLApertura > 0 && errCPI == nil && ccl > 0 {
			cpiInicio, _ := services.BuscarIndiceEnFecha(serieCPI, p.CreadoEn)
			inflacionUsaPct := (cpiActual.Valor/cpiInicio - 1) * 100
			invertidoUSD := invertido / *p.CCLApertura
			invertidoAjustadoUSD := invertidoUSD * (cpiActual.Valor / cpiInicio)
			valorActualUSD := p.ValorTotalARS / ccl
			retornoRealUSD := valorActualUSD - invertidoAjustadoUSD
			retornoRealUSDPct := 0.0
			if invertidoAjustadoUSD > 0 {
				retornoRealUSDPct = (retornoRealUSD / invertidoAjustadoUSD) * 100
			}
			rp.InvertidoUSD = &invertidoUSD
			rp.ValorActualUSD = &valorActualUSD
			rp.InflacionUsaPct = &inflacionUsaPct
			rp.RetornoRealUSD = &retornoRealUSD
			rp.RetornoRealUSDPct = &retornoRealUSDPct
		}

		resultado.Posiciones = append(resultado.Posiciones, rp)
	}

	// Recorremos resultado.Posiciones (no la lista original) para no contar de nuevo
	// las posiciones sin cotización que ya se excluyeron arriba.
	for _, rp := range resultado.Posiciones {
		resultado.TotalInvertidoARS += rp.InvertidoARS
		resultado.TotalActualARS += rp.ValorActualARS
	}
	resultado.TotalRetornoNominalARS = resultado.TotalActualARS - resultado.TotalInvertidoARS
	if resultado.TotalInvertidoARS > 0 {
		resultado.TotalRetornoNominalPct = (resultado.TotalRetornoNominalARS / resultado.TotalInvertidoARS) * 100
	}
	resultado.TotalRetornoRealARS = resultado.TotalActualARS - totalInvertidoAjustadoARS
	if totalInvertidoAjustadoARS > 0 {
		resultado.TotalRetornoRealPct = (resultado.TotalRetornoRealARS / totalInvertidoAjustadoARS) * 100
		resultado.TotalInflacionArsPct = (totalInvertidoAjustadoARS/resultado.TotalInvertidoARS - 1) * 100
	}

	resultado.Resumen = fmt.Sprintf(
		"Ganaste %.1f%% nominal pero %.1f%% real (inflación acumulada del %.1f%% desde que abriste tus posiciones)",
		resultado.TotalRetornoNominalPct, resultado.TotalRetornoRealPct, resultado.TotalInflacionArsPct,
	)

	c.JSON(http.StatusOK, resultado)
}

// POST /api/compra
func PostCompra(c *gin.Context) {
	var req models.CompraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Ticker = strings.ToUpper(req.Ticker)

	ccl, err := services.ObtenerDolarCCL()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "no se pudo obtener el dólar CCL: " + err.Error()})
		return
	}

	if err := db.RegistrarCompra(req, ccl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Compra registrada correctamente"})
}

// POST /api/venta
func PostVenta(c *gin.Context) {
	var req models.VentaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Ticker = strings.ToUpper(req.Ticker)

	if err := db.RegistrarVenta(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Venta registrada correctamente"})
}

// GET /api/historial?ticker=AAPL&limit=50
func GetHistorial(c *gin.Context) {
	ticker := strings.ToUpper(c.Query("ticker"))
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	ops, err := db.ObtenerHistorial(ticker, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ops == nil {
		ops = []models.Operacion{}
	}
	c.JSON(http.StatusOK, ops)
}

// GET /api/cotizacion/:ticker
func GetCotizacion(c *gin.Context) {
	ticker := strings.ToUpper(c.Param("ticker"))

	ccl, _ := services.ObtenerDolarCCL()
	precioUSD, precioARS, err := services.ObtenerPrecioARS(ticker, ccl)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.CotizacionResponse{
		Ticker:    ticker,
		PrecioUSD: precioUSD,
		PrecioARS: precioARS,
		DolarCCL:  ccl,
	})
}

// GET /api/health
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
