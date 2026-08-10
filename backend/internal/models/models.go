package models

import "time"

// Posicion representa una posición activa en la cartera
type Posicion struct {
	ID            int       `json:"id" db:"id"`
	Ticker        string    `json:"ticker" db:"ticker"`
	Cantidad      float64   `json:"cantidad" db:"cantidad"`
	PrecioPromARS float64   `json:"precio_prom_ars" db:"precio_prom_ars"`
	EsCedear      bool      `json:"es_cedear" db:"es_cedear"`
	Broker        string    `json:"broker" db:"broker"`
	CreadoEn      time.Time `json:"creado_en" db:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en" db:"actualizado_en"`
	// CCLApertura es el dólar CCL vigente cuando se abrió la posición (primera compra).
	// Nil en posiciones creadas antes de que este campo existiera: sin este dato no se
	// puede calcular el retorno real en USD, así que esos casos se omiten del cálculo.
	CCLApertura *float64 `json:"ccl_apertura,omitempty" db:"ccl_apertura"`

	// Campos calculados (no persistidos)
	// PrecioDisponible es false cuando falló el fetch a Yahoo Finance: en ese caso
	// los campos de abajo quedan en cero y NO representan una pérdida real, así que
	// el frontend debe mostrar "sin cotización" en vez de un P&L falso.
	PrecioDisponible bool    `json:"precio_disponible"`
	PrecioActualARS  float64 `json:"precio_actual_ars,omitempty"`
	PrecioActualUSD  float64 `json:"precio_actual_usd,omitempty"`
	PnlARS           float64 `json:"pnl_ars,omitempty"`
	PnlPct           float64 `json:"pnl_pct,omitempty"`
	ValorTotalARS    float64 `json:"valor_total_ars,omitempty"`
}

// Operacion representa una compra o venta registrada
type Operacion struct {
	ID          int       `json:"id" db:"id"`
	Tipo        string    `json:"tipo" db:"tipo"` // COMPRA | VENTA
	Ticker      string    `json:"ticker" db:"ticker"`
	Cantidad    float64   `json:"cantidad" db:"cantidad"`
	PrecioARS   float64   `json:"precio_ars" db:"precio_ars"`
	ComisionARS float64   `json:"comision_ars" db:"comision_ars"`
	TotalARS    float64   `json:"total_ars" db:"total_ars"`
	EsCedear    bool      `json:"es_cedear" db:"es_cedear"`
	Broker      string    `json:"broker" db:"broker"`
	Notas       string    `json:"notas" db:"notas"`
	FechaOpera  time.Time `json:"fecha_opera" db:"fecha_opera"`
}

// ResumenCartera es el payload completo que devuelve GET /api/cartera
type ResumenCartera struct {
	Posiciones     []Posicion `json:"posiciones"`
	TotalInvertido float64    `json:"total_invertido_ars"`
	TotalActual    float64    `json:"total_actual_ars"`
	TotalPnlARS    float64    `json:"total_pnl_ars"`
	TotalPnlPct    float64    `json:"total_pnl_pct"`
	TotalUSD       float64    `json:"total_usd"`
	DolarCCL       float64    `json:"dolar_ccl"`
	// PosicionesSinCotizar lista los tickers cuyo precio no se pudo obtener en este
	// request; sus montos NO están incluidos en los totales de arriba.
	PosicionesSinCotizar []string `json:"posiciones_sin_cotizar,omitempty"`
}

// Request bodies
type CompraRequest struct {
	Ticker      string  `json:"ticker" binding:"required"`
	Cantidad    float64 `json:"cantidad" binding:"required,gt=0"`
	PrecioARS   float64 `json:"precio_ars" binding:"required,gt=0"`
	ComisionARS float64 `json:"comision_ars"`
	EsCedear    bool    `json:"es_cedear"`
	Broker      string  `json:"broker" binding:"required"`
	Notas       string  `json:"notas"`
	// FechaOpera es opcional, formato "2006-01-02". Vacío = ahora. Se usa para
	// cargar historial real (DCA pasado); ver nota sobre ccl_apertura en db.go.
	FechaOpera string `json:"fecha_opera"`
}

type VentaRequest struct {
	Ticker      string  `json:"ticker" binding:"required"`
	Cantidad    float64 `json:"cantidad" binding:"required,gt=0"`
	PrecioARS   float64 `json:"precio_ars" binding:"required,gt=0"`
	ComisionARS float64 `json:"comision_ars"`
	Broker      string  `json:"broker" binding:"required"`
	Notas       string  `json:"notas"`
	FechaOpera  string  `json:"fecha_opera"`
}

// CotizacionResponse es lo que devuelve GET /api/cotizacion/:ticker
type CotizacionResponse struct {
	Ticker   string  `json:"ticker"`
	PrecioUSD float64 `json:"precio_usd"`
	PrecioARS float64 `json:"precio_ars"`
	DolarCCL float64 `json:"dolar_ccl"`
}

// RendimientoPosicion compara el retorno nominal vs el retorno real (ajustado
// por inflación) de una posición individual.
type RendimientoPosicion struct {
	Ticker              string  `json:"ticker"`
	InvertidoARS        float64 `json:"invertido_ars"`
	ValorActualARS      float64 `json:"valor_actual_ars"`
	InflacionArsPct     float64 `json:"inflacion_ars_pct"`
	RetornoNominalARS   float64 `json:"retorno_nominal_ars"`
	RetornoNominalPct   float64 `json:"retorno_nominal_pct"`
	RetornoRealARS      float64 `json:"retorno_real_ars"`
	RetornoRealPct      float64 `json:"retorno_real_pct"`

	// Campos en USD: null si la posición no tiene CCL de apertura registrado
	// (posiciones creadas antes de la Fase 2).
	InvertidoUSD      *float64 `json:"invertido_usd,omitempty"`
	ValorActualUSD    *float64 `json:"valor_actual_usd,omitempty"`
	InflacionUsaPct   *float64 `json:"inflacion_usa_pct,omitempty"`
	RetornoRealUSD    *float64 `json:"retorno_real_usd,omitempty"`
	RetornoRealUSDPct *float64 `json:"retorno_real_usd_pct,omitempty"`
}

// RendimientoReal es el payload completo de GET /api/rendimiento/real
type RendimientoReal struct {
	Posiciones []RendimientoPosicion `json:"posiciones"`

	TotalInvertidoARS      float64 `json:"total_invertido_ars"`
	TotalActualARS         float64 `json:"total_actual_ars"`
	TotalInflacionArsPct   float64 `json:"total_inflacion_ars_pct"`
	TotalRetornoNominalARS float64 `json:"total_retorno_nominal_ars"`
	TotalRetornoNominalPct float64 `json:"total_retorno_nominal_pct"`
	TotalRetornoRealARS    float64 `json:"total_retorno_real_ars"`
	TotalRetornoRealPct    float64 `json:"total_retorno_real_pct"`

	// Fecha (mes) del dato de inflación más reciente usado en el cálculo.
	IpcFechaActual string `json:"ipc_fecha_actual"`

	// Resumen en lenguaje natural: "Ganaste X% nominal pero Y% real"
	Resumen string `json:"resumen"`

	// PosicionesSinCotizar lista los tickers excluidos del cálculo por no tener
	// precio actual disponible en este request.
	PosicionesSinCotizar []string `json:"posiciones_sin_cotizar,omitempty"`
}

// Alerta es una alerta de precio objetivo que dispara un mensaje de Telegram
// la primera vez que el precio del ticker cruza el objetivo en la dirección indicada.
type Alerta struct {
	ID             int        `json:"id" db:"id"`
	Ticker         string     `json:"ticker" db:"ticker"`
	PrecioObjetivo float64    `json:"precio_objetivo" db:"precio_objetivo"`
	Direccion      string     `json:"direccion" db:"direccion"` // ARRIBA | ABAJO
	EsCedear       bool       `json:"es_cedear" db:"es_cedear"`
	Activa         bool       `json:"activa" db:"activa"`
	CreadoEn       time.Time  `json:"creado_en" db:"creado_en"`
	DisparadaEn    *time.Time `json:"disparada_en,omitempty" db:"disparada_en"`
}

type AlertaRequest struct {
	Ticker         string  `json:"ticker" binding:"required"`
	PrecioObjetivo float64 `json:"precio_objetivo" binding:"required,gt=0"`
	Direccion      string  `json:"direccion" binding:"required,oneof=ARRIBA ABAJO"`
	EsCedear       bool    `json:"es_cedear"`
}

// WaitlistRequest es el body de POST /api/waitlist (landing page pública).
type WaitlistRequest struct {
	Email string `json:"email" binding:"required,email"`
}
