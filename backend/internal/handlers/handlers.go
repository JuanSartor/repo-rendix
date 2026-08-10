package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
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

	totales := services.CalcularPreciosActuales(posiciones, ccl)

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
	totalesPrecios := services.CalcularPreciosActuales(posiciones, ccl)

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

// GET /api/alertas
func GetAlertas(c *gin.Context) {
	alertas, err := db.ObtenerAlertas()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if alertas == nil {
		alertas = []models.Alerta{}
	}
	c.JSON(http.StatusOK, alertas)
}

// POST /api/alertas
func PostAlertas(c *gin.Context) {
	var req models.AlertaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Ticker = strings.ToUpper(req.Ticker)

	if err := db.CrearAlerta(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Alerta creada correctamente"})
}

// DELETE /api/alertas/:id
func DeleteAlerta(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	if err := db.EliminarAlerta(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Alerta eliminada"})
}

// POST /api/importar-csv
// Carga masiva de operaciones desde un CSV (multipart, campo "file"). Columnas
// requeridas: tipo,ticker,cantidad,precio_ars,broker. Opcionales: comision_ars,
// es_cedear,fecha (YYYY-MM-DD),notas. El orden de columnas no importa (se leen
// por nombre de header), y las filas se procesan ordenadas por fecha para que
// una VENTA no falle por ejecutarse "antes" que su COMPRA correspondiente si el
// archivo no vino ya ordenado cronológicamente.
func PostImportarCSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "falta el archivo CSV (campo 'file')"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	filas, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error leyendo CSV: " + err.Error()})
		return
	}
	if len(filas) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el CSV no tiene filas de datos"})
		return
	}

	header := filas[0]
	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, req := range []string{"tipo", "ticker", "cantidad", "precio_ars", "broker"} {
		if _, ok := col[req]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("falta la columna '%s' en el CSV", req)})
			return
		}
	}

	get := func(fila []string, nombre string) string {
		idx, ok := col[nombre]
		if !ok || idx >= len(fila) {
			return ""
		}
		return strings.TrimSpace(fila[idx])
	}

	datos := filas[1:]
	sort.SliceStable(datos, func(i, j int) bool {
		// Fechas vacías (= "hoy") ordenan al final; YYYY-MM-DD ordena bien como string.
		fi, fj := get(datos[i], "fecha"), get(datos[j], "fecha")
		if fi == "" {
			return false
		}
		if fj == "" {
			return true
		}
		return fi < fj
	})

	ccl, _ := services.ObtenerDolarCCL() // si falla queda en 0; solo afecta ccl_apertura de compras "de hoy"

	var importadas int
	var errores []string

	for i, fila := range datos {
		numFila := i + 2 // +1 por el header, +1 porque las filas se cuentan desde 1
		tipo := strings.ToUpper(get(fila, "tipo"))
		ticker := strings.ToUpper(get(fila, "ticker"))
		cantidad, errC := strconv.ParseFloat(get(fila, "cantidad"), 64)
		precio, errP := strconv.ParseFloat(get(fila, "precio_ars"), 64)
		comision := 0.0
		if s := get(fila, "comision_ars"); s != "" {
			comision, _ = strconv.ParseFloat(s, 64)
		}
		esCedear := strings.EqualFold(get(fila, "es_cedear"), "true") || get(fila, "es_cedear") == "1"
		broker := get(fila, "broker")
		fecha := get(fila, "fecha")
		notas := get(fila, "notas")

		if errC != nil || errP != nil || cantidad <= 0 || precio <= 0 || ticker == "" || broker == "" {
			errores = append(errores, fmt.Sprintf("fila %d: datos inválidos o incompletos", numFila))
			continue
		}

		var errOp error
		switch tipo {
		case "COMPRA":
			errOp = db.RegistrarCompra(models.CompraRequest{
				Ticker: ticker, Cantidad: cantidad, PrecioARS: precio,
				ComisionARS: comision, EsCedear: esCedear, Broker: broker,
				Notas: notas, FechaOpera: fecha,
			}, ccl)
		case "VENTA":
			errOp = db.RegistrarVenta(models.VentaRequest{
				Ticker: ticker, Cantidad: cantidad, PrecioARS: precio,
				ComisionARS: comision, Broker: broker, Notas: notas, FechaOpera: fecha,
			})
		default:
			errOp = fmt.Errorf("tipo '%s' inválido (debe ser COMPRA o VENTA)", tipo)
		}

		if errOp != nil {
			errores = append(errores, fmt.Sprintf("fila %d (%s): %s", numFila, ticker, errOp.Error()))
			continue
		}
		importadas++
	}

	c.JSON(http.StatusOK, gin.H{
		"importadas": importadas,
		"errores":    errores,
	})
}
