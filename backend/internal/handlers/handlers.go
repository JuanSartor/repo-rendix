package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"portfolio-tracker/internal/db"
	"portfolio-tracker/internal/models"
	"portfolio-tracker/internal/services"
)

// totalesCartera son los agregados calculados junto con los precios actuales de cada posición.
type totalesCartera struct {
	TotalInvertido float64
	TotalActual    float64
}

// calcularPreciosActuales completa PrecioActualARS/USD, PnlARS/Pct y ValorTotalARS de cada
// posición (mutando el slice in-place) y devuelve los totales agregados. Compartido entre
// GetCartera y GetRendimientoReal para no duplicar las llamadas a Yahoo Finance.
func calcularPreciosActuales(posiciones []models.Posicion, ccl float64) totalesCartera {
	var totales totalesCartera

	for i := range posiciones {
		p := &posiciones[i]
		invertido := p.Cantidad * p.PrecioPromARS
		totales.TotalInvertido += invertido

		var precioActualARS float64
		if p.EsCedear {
			_, precioActualARS, _ = services.ObtenerPrecioARS(p.Ticker, ccl)
		} else {
			// Acciones locales: sufijo .BA en Yahoo Finance
			tickerBA := services.ResolverTickerLocal(p.Ticker)
			precioActualARS, _ = services.ObtenerPrecioUSD(tickerBA + ".BA")
		}

		actualTotal := p.Cantidad * precioActualARS
		totales.TotalActual += actualTotal

		p.PrecioActualARS = precioActualARS
		p.ValorTotalARS = actualTotal
		p.PnlARS = actualTotal - invertido
		if invertido > 0 {
			p.PnlPct = (p.PnlARS / invertido) * 100
		}
		if ccl > 0 {
			p.PrecioActualUSD = precioActualARS / ccl
		}
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
		Posiciones:     posiciones,
		TotalInvertido: totales.TotalInvertido,
		TotalActual:    totales.TotalActual,
		TotalPnlARS:    totalPnl,
		TotalPnlPct:    totalPct,
		TotalUSD:       totalUSD,
		DolarCCL:       ccl,
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
	calcularPreciosActuales(posiciones, ccl)

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
		Posiciones:     make([]models.RendimientoPosicion, 0, len(posiciones)),
		IpcFechaActual: ipcActual.Fecha.Format("2006-01"),
	}

	var totalInvertidoAjustadoARS float64

	for _, p := range posiciones {
		invertido := p.Cantidad * p.PrecioPromARS
		ipcInicio, _ := services.BuscarIndiceEnFecha(serieIPC, p.CreadoEn)
		inflacionPct := (ipcActual.Valor/ipcInicio - 1) * 100
		invertidoAjustado := invertido * (ipcActual.Valor / ipcInicio)
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

	resultado.TotalInvertidoARS = 0
	resultado.TotalActualARS = 0
	for _, p := range posiciones {
		resultado.TotalInvertidoARS += p.Cantidad * p.PrecioPromARS
		resultado.TotalActualARS += p.ValorTotalARS
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
