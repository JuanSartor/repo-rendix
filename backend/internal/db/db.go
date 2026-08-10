package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"portfolio-tracker/internal/models"
)

var DB *sql.DB

// Init abre la conexión a Postgres (Supabase) y crea las tablas si no existen
func Init(connString string) error {
	// Supabase usa PgBouncer en modo transaction para el pooler (puerto 6543),
	// que no soporta prepared statements reutilizados entre conexiones.
	// Forzamos el protocolo simple para evitar errores de "prepared statement already exists".
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("error parseando connection string: %w", err)
	}
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	DB = stdlib.OpenDB(*cfg)
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error conectando db: %w", err)
	}
	return crearTablas()
}

func crearTablas() error {
	schema := `
	CREATE TABLE IF NOT EXISTS posiciones (
		id              SERIAL PRIMARY KEY,
		ticker          TEXT NOT NULL UNIQUE,
		cantidad        DOUBLE PRECISION NOT NULL DEFAULT 0,
		precio_prom_ars DOUBLE PRECISION NOT NULL DEFAULT 0,
		es_cedear       BOOLEAN NOT NULL DEFAULT FALSE,
		broker          TEXT NOT NULL DEFAULT '',
		creado_en       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		actualizado_en  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		ccl_apertura    DOUBLE PRECISION
	);

	CREATE TABLE IF NOT EXISTS operaciones (
		id           SERIAL PRIMARY KEY,
		tipo         TEXT NOT NULL CHECK(tipo IN ('COMPRA','VENTA')),
		ticker       TEXT NOT NULL,
		cantidad     DOUBLE PRECISION NOT NULL,
		precio_ars   DOUBLE PRECISION NOT NULL,
		total_ars    DOUBLE PRECISION NOT NULL,
		es_cedear    BOOLEAN NOT NULL DEFAULT FALSE,
		broker       TEXT NOT NULL DEFAULT '',
		notas        TEXT DEFAULT '',
		fecha_opera  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		comision_ars DOUBLE PRECISION NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS alertas (
		id              SERIAL PRIMARY KEY,
		ticker          TEXT NOT NULL,
		precio_objetivo DOUBLE PRECISION NOT NULL,
		direccion       TEXT NOT NULL CHECK(direccion IN ('ARRIBA','ABAJO')),
		es_cedear       BOOLEAN NOT NULL DEFAULT FALSE,
		activa          BOOLEAN NOT NULL DEFAULT TRUE,
		creado_en       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		disparada_en    TIMESTAMPTZ
	);

	-- Migraciones idempotentes para bases creadas antes de estas columnas.
	ALTER TABLE posiciones ADD COLUMN IF NOT EXISTS ccl_apertura DOUBLE PRECISION;
	ALTER TABLE operaciones ADD COLUMN IF NOT EXISTS comision_ars DOUBLE PRECISION NOT NULL DEFAULT 0;
	`
	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("error creando tablas: %w", err)
	}
	log.Println("✅ Base de datos lista")
	return nil
}

// ─── Posiciones ───────────────────────────────────────────────────────────────

func ObtenerPosiciones() ([]models.Posicion, error) {
	rows, err := DB.Query(`
		SELECT id, ticker, cantidad, precio_prom_ars, es_cedear, broker, creado_en, actualizado_en, ccl_apertura
		FROM posiciones WHERE cantidad > 0
		ORDER BY ticker
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posiciones []models.Posicion
	for rows.Next() {
		var p models.Posicion
		err := rows.Scan(
			&p.ID, &p.Ticker, &p.Cantidad, &p.PrecioPromARS,
			&p.EsCedear, &p.Broker, &p.CreadoEn, &p.ActualizadoEn, &p.CCLApertura,
		)
		if err != nil {
			return nil, err
		}
		posiciones = append(posiciones, p)
	}
	return posiciones, nil
}

// RegistrarCompra registra una compra con precio promedio ponderado (la comisión se
// prorratea en el costo unitario, así que encarece el precio promedio de la posición).
// cclActual se guarda como ccl_apertura solo si la posición es nueva; se usa más
// adelante para calcular el retorno real en USD.
func RegistrarCompra(req models.CompraRequest, cclActual float64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ticker := req.Ticker
	now := time.Now()
	precioEfectivo := req.PrecioARS + req.ComisionARS/req.Cantidad

	// Buscar posición existente
	var id int
	var cantActual, promActual float64
	err = tx.QueryRow(
		`SELECT id, cantidad, precio_prom_ars FROM posiciones WHERE ticker = $1`, ticker,
	).Scan(&id, &cantActual, &promActual)

	if err == sql.ErrNoRows {
		// Nueva posición
		_, err = tx.Exec(`
			INSERT INTO posiciones (ticker, cantidad, precio_prom_ars, es_cedear, broker, creado_en, actualizado_en, ccl_apertura)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			ticker, req.Cantidad, precioEfectivo,
			req.EsCedear, req.Broker, now, now, cclActual,
		)
	} else if err == nil {
		// Actualizar con precio promedio ponderado (incluye comisión de esta compra)
		nuevaCant := cantActual + req.Cantidad
		nuevoProm := ((cantActual * promActual) + (req.Cantidad * precioEfectivo)) / nuevaCant
		_, err = tx.Exec(`
			UPDATE posiciones SET cantidad=$1, precio_prom_ars=$2, broker=$3, actualizado_en=$4
			WHERE id=$5`,
			nuevaCant, nuevoProm, req.Broker, now, id,
		)
	}
	if err != nil {
		return fmt.Errorf("error actualizando posición: %w", err)
	}

	// Registrar en historial (total_ars incluye la comisión: es el costo real desembolsado)
	_, err = tx.Exec(`
		INSERT INTO operaciones (tipo, ticker, cantidad, precio_ars, total_ars, es_cedear, broker, notas, fecha_opera, comision_ars)
		VALUES ('COMPRA', $1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		ticker, req.Cantidad, req.PrecioARS,
		req.Cantidad*req.PrecioARS+req.ComisionARS,
		req.EsCedear, req.Broker, req.Notas, now, req.ComisionARS,
	)
	if err != nil {
		return fmt.Errorf("error registrando operación: %w", err)
	}

	return tx.Commit()
}

func RegistrarVenta(req models.VentaRequest) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id int
	var cantActual float64
	err = tx.QueryRow(
		`SELECT id, cantidad FROM posiciones WHERE ticker = $1`, req.Ticker,
	).Scan(&id, &cantActual)

	if err == sql.ErrNoRows {
		return fmt.Errorf("no tenés posición en %s", req.Ticker)
	}
	if req.Cantidad > cantActual {
		return fmt.Errorf("cantidad a vender (%.2f) supera la posición actual (%.2f)", req.Cantidad, cantActual)
	}

	nuevaCant := cantActual - req.Cantidad
	_, err = tx.Exec(
		`UPDATE posiciones SET cantidad=$1, actualizado_en=$2 WHERE id=$3`,
		nuevaCant, time.Now(), id,
	)
	if err != nil {
		return err
	}

	// total_ars neto de comisión: es lo que efectivamente cobrás por la venta
	_, err = tx.Exec(`
		INSERT INTO operaciones (tipo, ticker, cantidad, precio_ars, total_ars, es_cedear, broker, notas, fecha_opera, comision_ars)
		VALUES ('VENTA', $1, $2, $3, $4, FALSE, $5, $6, $7, $8)`,
		req.Ticker, req.Cantidad, req.PrecioARS,
		req.Cantidad*req.PrecioARS-req.ComisionARS,
		req.Broker, req.Notas, time.Now(), req.ComisionARS,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ObtenerComprasPorTicker devuelve todas las operaciones de COMPRA de un ticker,
// ordenadas ascendente por fecha. Se usa para calcular el retorno real ajustando
// la inflación por cada tramo de compra (DCA) en vez de una sola fecha de apertura.
func ObtenerComprasPorTicker(ticker string) ([]models.Operacion, error) {
	rows, err := DB.Query(`
		SELECT id, tipo, ticker, cantidad, precio_ars, total_ars, es_cedear, broker, notas, fecha_opera, comision_ars
		FROM operaciones
		WHERE ticker = $1 AND tipo = 'COMPRA'
		ORDER BY fecha_opera ASC
	`, ticker)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []models.Operacion
	for rows.Next() {
		var op models.Operacion
		err := rows.Scan(
			&op.ID, &op.Tipo, &op.Ticker, &op.Cantidad,
			&op.PrecioARS, &op.TotalARS, &op.EsCedear,
			&op.Broker, &op.Notas, &op.FechaOpera, &op.ComisionARS,
		)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func ObtenerHistorial(ticker string, limit int) ([]models.Operacion, error) {
	query := `SELECT id, tipo, ticker, cantidad, precio_ars, total_ars, es_cedear, broker, notas, fecha_opera, comision_ars
		FROM operaciones`
	args := []interface{}{}
	argN := 1

	if ticker != "" {
		query += fmt.Sprintf(" WHERE ticker = $%d", argN)
		args = append(args, ticker)
		argN++
	}
	query += " ORDER BY fecha_opera DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argN)
		args = append(args, limit)
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []models.Operacion
	for rows.Next() {
		var op models.Operacion
		err := rows.Scan(
			&op.ID, &op.Tipo, &op.Ticker, &op.Cantidad,
			&op.PrecioARS, &op.TotalARS, &op.EsCedear,
			&op.Broker, &op.Notas, &op.FechaOpera, &op.ComisionARS,
		)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

// ─── Alertas ──────────────────────────────────────────────────────────────────

func CrearAlerta(req models.AlertaRequest) error {
	_, err := DB.Exec(`
		INSERT INTO alertas (ticker, precio_objetivo, direccion, es_cedear)
		VALUES ($1, $2, $3, $4)`,
		req.Ticker, req.PrecioObjetivo, req.Direccion, req.EsCedear,
	)
	if err != nil {
		return fmt.Errorf("error creando alerta: %w", err)
	}
	return nil
}

func ObtenerAlertas() ([]models.Alerta, error) {
	return consultarAlertas("SELECT id, ticker, precio_objetivo, direccion, es_cedear, activa, creado_en, disparada_en FROM alertas ORDER BY creado_en DESC")
}

func ObtenerAlertasActivas() ([]models.Alerta, error) {
	return consultarAlertas("SELECT id, ticker, precio_objetivo, direccion, es_cedear, activa, creado_en, disparada_en FROM alertas WHERE activa = TRUE ORDER BY creado_en")
}

func consultarAlertas(query string) ([]models.Alerta, error) {
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alertas []models.Alerta
	for rows.Next() {
		var a models.Alerta
		err := rows.Scan(
			&a.ID, &a.Ticker, &a.PrecioObjetivo, &a.Direccion,
			&a.EsCedear, &a.Activa, &a.CreadoEn, &a.DisparadaEn,
		)
		if err != nil {
			return nil, err
		}
		alertas = append(alertas, a)
	}
	return alertas, nil
}

// MarcarAlertaDisparada desactiva la alerta y registra cuándo se disparó, para
// no volver a avisar por el mismo cruce de precio en el próximo chequeo.
func MarcarAlertaDisparada(id int) error {
	_, err := DB.Exec(
		`UPDATE alertas SET activa = FALSE, disparada_en = NOW() WHERE id = $1`, id,
	)
	return err
}

func EliminarAlerta(id int) error {
	res, err := DB.Exec(`DELETE FROM alertas WHERE id = $1`, id)
	if err != nil {
		return err
	}
	filas, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if filas == 0 {
		return fmt.Errorf("no existe alerta con id %d", id)
	}
	return nil
}
