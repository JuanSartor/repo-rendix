package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"portfolio-tracker/internal/db"
	"portfolio-tracker/internal/models"
	"portfolio-tracker/internal/services"
)

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

	var totalInvertido, totalActual float64

	for i := range posiciones {
		p := &posiciones[i]
		invertido := p.Cantidad * p.PrecioPromARS
		totalInvertido += invertido

		var precioActualARS float64
		if p.EsCedear {
			_, precioActualARS, _ = services.ObtenerPrecioARS(p.Ticker, ccl)
		} else {
			// Acciones locales: sufijo .BA en Yahoo Finance
			tickerBA := services.ResolverTickerLocal(p.Ticker)
			precioActualARS, _ = services.ObtenerPrecioUSD(tickerBA + ".BA")
		}

		actualTotal := p.Cantidad * precioActualARS
		totalActual += actualTotal

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

	totalPnl := totalActual - totalInvertido
	totalPct := 0.0
	if totalInvertido > 0 {
		totalPct = (totalPnl / totalInvertido) * 100
	}
	totalUSD := 0.0
	if ccl > 0 {
		totalUSD = totalActual / ccl
	}

	c.JSON(http.StatusOK, models.ResumenCartera{
		Posiciones:     posiciones,
		TotalInvertido: totalInvertido,
		TotalActual:    totalActual,
		TotalPnlARS:    totalPnl,
		TotalPnlPct:    totalPct,
		TotalUSD:       totalUSD,
		DolarCCL:       ccl,
	})
}

// POST /api/compra
func PostCompra(c *gin.Context) {
	var req models.CompraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Ticker = strings.ToUpper(req.Ticker)

	if err := db.RegistrarCompra(req); err != nil {
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
