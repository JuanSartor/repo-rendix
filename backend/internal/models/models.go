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

	// Campos calculados (no persistidos)
	PrecioActualARS float64 `json:"precio_actual_ars,omitempty"`
	PrecioActualUSD float64 `json:"precio_actual_usd,omitempty"`
	PnlARS          float64 `json:"pnl_ars,omitempty"`
	PnlPct          float64 `json:"pnl_pct,omitempty"`
	ValorTotalARS   float64 `json:"valor_total_ars,omitempty"`
}

// Operacion representa una compra o venta registrada
type Operacion struct {
	ID          int       `json:"id" db:"id"`
	Tipo        string    `json:"tipo" db:"tipo"` // COMPRA | VENTA
	Ticker      string    `json:"ticker" db:"ticker"`
	Cantidad    float64   `json:"cantidad" db:"cantidad"`
	PrecioARS   float64   `json:"precio_ars" db:"precio_ars"`
	TotalARS    float64   `json:"total_ars" db:"total_ars"`
	EsCedear    bool      `json:"es_cedear" db:"es_cedear"`
	Broker      string    `json:"broker" db:"broker"`
	Notas       string    `json:"notas" db:"notas"`
	FechaOpera  time.Time `json:"fecha_opera" db:"fecha_opera"`
}

// ResumenCartera es el payload completo que devuelve GET /api/cartera
type ResumenCartera struct {
	Posiciones      []Posicion `json:"posiciones"`
	TotalInvertido  float64    `json:"total_invertido_ars"`
	TotalActual     float64    `json:"total_actual_ars"`
	TotalPnlARS     float64    `json:"total_pnl_ars"`
	TotalPnlPct     float64    `json:"total_pnl_pct"`
	TotalUSD        float64    `json:"total_usd"`
	DolarCCL        float64    `json:"dolar_ccl"`
}

// Request bodies
type CompraRequest struct {
	Ticker      string  `json:"ticker" binding:"required"`
	Cantidad    float64 `json:"cantidad" binding:"required,gt=0"`
	PrecioARS   float64 `json:"precio_ars" binding:"required,gt=0"`
	EsCedear    bool    `json:"es_cedear"`
	Broker      string  `json:"broker" binding:"required"`
	Notas       string  `json:"notas"`
}

type VentaRequest struct {
	Ticker    string  `json:"ticker" binding:"required"`
	Cantidad  float64 `json:"cantidad" binding:"required,gt=0"`
	PrecioARS float64 `json:"precio_ars" binding:"required,gt=0"`
	Broker    string  `json:"broker" binding:"required"`
	Notas     string  `json:"notas"`
}

// CotizacionResponse es lo que devuelve GET /api/cotizacion/:ticker
type CotizacionResponse struct {
	Ticker   string  `json:"ticker"`
	PrecioUSD float64 `json:"precio_usd"`
	PrecioARS float64 `json:"precio_ars"`
	DolarCCL float64 `json:"dolar_ccl"`
}
