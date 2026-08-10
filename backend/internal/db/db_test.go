package db

import (
	"os"
	"testing"

	"portfolio-tracker/internal/models"
)

// Estos tests necesitan una Postgres real (no un mock): usan sql.DB.Query/Exec
// contra columnas y constraints reales (CHECK de tipo/dirección, etc.). Corren
// contra DATABASE_URL_TEST si está seteada, o contra un Postgres local en
// localhost:5433 (ver README/CI: `docker run -p 5433:5432 postgres:16`).
func TestMain(m *testing.M) {
	dsn := os.Getenv("DATABASE_URL_TEST")
	if dsn == "" {
		dsn = "postgresql://postgres:postgres@localhost:5433/rendix_test?sslmode=disable"
	}
	if err := Init(dsn); err != nil {
		panic("no se pudo conectar a la DB de test (" + dsn + "): " + err.Error())
	}
	os.Exit(m.Run())
}

// limpiarTicker borra cualquier resto de un ticker de test antes de correr, y
// programa la limpieza para después del test también (por si el test falla a
// mitad de camino y deja basura).
func limpiarTicker(t *testing.T, ticker string) {
	t.Helper()
	borrar := func() {
		DB.Exec(`DELETE FROM operaciones WHERE ticker = $1`, ticker)
		DB.Exec(`DELETE FROM posiciones WHERE ticker = $1`, ticker)
	}
	borrar()
	t.Cleanup(borrar)
}

func buscarPosicion(t *testing.T, posiciones []models.Posicion, ticker string) models.Posicion {
	t.Helper()
	for _, p := range posiciones {
		if p.Ticker == ticker {
			return p
		}
	}
	t.Fatalf("no se encontró la posición %s entre las devueltas por ObtenerPosiciones", ticker)
	return models.Posicion{}
}

func TestRegistrarCompra_PromedioPonderadoConComision(t *testing.T) {
	ticker := "TESTPPC"
	limpiarTicker(t, ticker)

	// Primera compra: 10 u a 100, comisión 50 -> precio efectivo = 100 + 50/10 = 105.
	// La comisión se prorratea en el costo unitario, así que encarece el promedio.
	if err := RegistrarCompra(models.CompraRequest{
		Ticker: ticker, Cantidad: 10, PrecioARS: 100, ComisionARS: 50,
		EsCedear: true, Broker: "Test",
	}, 1000); err != nil {
		t.Fatalf("error en primera compra: %v", err)
	}

	posiciones, err := ObtenerPosiciones()
	if err != nil {
		t.Fatalf("error obteniendo posiciones: %v", err)
	}
	p := buscarPosicion(t, posiciones, ticker)
	if p.PrecioPromARS != 105 {
		t.Errorf("esperaba precio_prom_ars=105, obtuve %v", p.PrecioPromARS)
	}
	if p.CCLApertura == nil || *p.CCLApertura != 1000 {
		t.Errorf("esperaba ccl_apertura=1000 (compra registrada hoy), obtuve %v", p.CCLApertura)
	}

	// Segunda compra: 10 u a 200, sin comisión -> precio efectivo = 200.
	// Nuevo promedio ponderado = (10*105 + 10*200) / 20 = 152.5
	if err := RegistrarCompra(models.CompraRequest{
		Ticker: ticker, Cantidad: 10, PrecioARS: 200,
		EsCedear: true, Broker: "Test",
	}, 1100); err != nil {
		t.Fatalf("error en segunda compra: %v", err)
	}

	posiciones, _ = ObtenerPosiciones()
	p = buscarPosicion(t, posiciones, ticker)
	if p.Cantidad != 20 {
		t.Errorf("esperaba cantidad=20, obtuve %v", p.Cantidad)
	}
	if p.PrecioPromARS != 152.5 {
		t.Errorf("esperaba precio_prom_ars=152.5, obtuve %v", p.PrecioPromARS)
	}
	// ccl_apertura se fija solo al abrir la posición, no debe cambiar en compras siguientes.
	if p.CCLApertura == nil || *p.CCLApertura != 1000 {
		t.Errorf("esperaba que ccl_apertura siguiera en 1000 tras la segunda compra, obtuve %v", p.CCLApertura)
	}
}

func TestRegistrarCompra_RetroactivaNoGuardaCCLApertura(t *testing.T) {
	ticker := "TESTRETRO"
	limpiarTicker(t, ticker)

	if err := RegistrarCompra(models.CompraRequest{
		Ticker: ticker, Cantidad: 5, PrecioARS: 100, EsCedear: false,
		Broker: "Test", FechaOpera: "2020-01-15",
	}, 999); err != nil {
		t.Fatalf("error registrando compra retroactiva: %v", err)
	}

	posiciones, _ := ObtenerPosiciones()
	p := buscarPosicion(t, posiciones, ticker)
	if p.CCLApertura != nil {
		t.Errorf("esperaba ccl_apertura=nil para una compra retroactiva (no hay CCL histórico real), obtuve %v", *p.CCLApertura)
	}
	if got := p.CreadoEn.Format("2006-01-02"); got != "2020-01-15" {
		t.Errorf("esperaba creado_en=2020-01-15, obtuve %s", got)
	}
}

func TestRegistrarVenta_ErroresYExito(t *testing.T) {
	ticker := "TESTVENTA"
	limpiarTicker(t, ticker)

	if err := RegistrarVenta(models.VentaRequest{Ticker: ticker, Cantidad: 1, PrecioARS: 100, Broker: "Test"}); err == nil {
		t.Error("esperaba error al vender un ticker sin posición abierta")
	}

	if err := RegistrarCompra(models.CompraRequest{
		Ticker: ticker, Cantidad: 10, PrecioARS: 100, EsCedear: true, Broker: "Test",
	}, 1000); err != nil {
		t.Fatalf("error preparando la posición: %v", err)
	}

	if err := RegistrarVenta(models.VentaRequest{Ticker: ticker, Cantidad: 20, PrecioARS: 100, Broker: "Test"}); err == nil {
		t.Error("esperaba error al vender más cantidad de la que hay en la posición")
	}

	if err := RegistrarVenta(models.VentaRequest{Ticker: ticker, Cantidad: 4, PrecioARS: 150, Broker: "Test"}); err != nil {
		t.Fatalf("no esperaba error en una venta válida: %v", err)
	}

	posiciones, _ := ObtenerPosiciones()
	p := buscarPosicion(t, posiciones, ticker)
	if p.Cantidad != 6 {
		t.Errorf("esperaba cantidad=6 tras la venta (10-4), obtuve %v", p.Cantidad)
	}
	// Una venta no debe tocar el precio promedio ponderado de la posición.
	if p.PrecioPromARS != 100 {
		t.Errorf("esperaba precio_prom_ars sin cambios (100), obtuve %v", p.PrecioPromARS)
	}
}

func TestObtenerComprasPorTicker_OrdenAscendentePorFecha(t *testing.T) {
	ticker := "TESTORDEN"
	limpiarTicker(t, ticker)

	fechas := []string{"2026-03-01", "2026-01-01", "2026-02-01"}
	for _, f := range fechas {
		if err := RegistrarCompra(models.CompraRequest{
			Ticker: ticker, Cantidad: 1, PrecioARS: 100, Broker: "Test", FechaOpera: f,
		}, 1000); err != nil {
			t.Fatalf("error registrando compra de %s: %v", f, err)
		}
	}

	compras, err := ObtenerComprasPorTicker(ticker)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(compras) != 3 {
		t.Fatalf("esperaba 3 compras, obtuve %d", len(compras))
	}
	esperado := []string{"2026-01-01", "2026-02-01", "2026-03-01"}
	for i, e := range esperado {
		if got := compras[i].FechaOpera.Format("2006-01-02"); got != e {
			t.Errorf("posición %d: esperaba fecha %s, obtuve %s (las compras deben venir ordenadas ascendente por fecha, no por orden de inserción)", i, e, got)
		}
	}
}
