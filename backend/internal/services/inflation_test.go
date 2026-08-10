package services

import (
	"testing"
	"time"
)

func fechaTest(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func serieTest() []IndicePrecio {
	// Serie ascendente: ene, mar, jun 2026 (con huecos, como pasaría con datos reales
	// donde no siempre hay un punto para cada mes exacto).
	return []IndicePrecio{
		{Fecha: fechaTest(2026, time.January, 1), Valor: 100},
		{Fecha: fechaTest(2026, time.March, 1), Valor: 120},
		{Fecha: fechaTest(2026, time.June, 1), Valor: 150},
	}
}

func TestBuscarIndiceEnFecha_MesExacto(t *testing.T) {
	got, ok := BuscarIndiceEnFecha(serieTest(), fechaTest(2026, time.March, 15))
	if !ok {
		t.Fatal("esperaba encontrar un valor")
	}
	if got != 120 {
		t.Errorf("esperaba 120 (marzo), obtuve %v", got)
	}
}

func TestBuscarIndiceEnFecha_MesEntrePuntos(t *testing.T) {
	// Abril no tiene punto propio: debe devolver el último disponible antes (marzo=120),
	// no interpolar ni saltar al de junio.
	got, ok := BuscarIndiceEnFecha(serieTest(), fechaTest(2026, time.April, 10))
	if !ok {
		t.Fatal("esperaba encontrar un valor")
	}
	if got != 120 {
		t.Errorf("esperaba 120 (arrastrado desde marzo), obtuve %v", got)
	}
}

func TestBuscarIndiceEnFecha_AntesDelInicioDeLaSerie(t *testing.T) {
	// Una fecha anterior al primer punto de la serie: no hay dato real, así que
	// usamos el primero disponible en vez de fallar o devolver cero.
	got, ok := BuscarIndiceEnFecha(serieTest(), fechaTest(2020, time.January, 1))
	if !ok {
		t.Fatal("esperaba encontrar un valor")
	}
	if got != 100 {
		t.Errorf("esperaba 100 (primer punto de la serie), obtuve %v", got)
	}
}

func TestBuscarIndiceEnFecha_ExactamenteElUltimoPunto(t *testing.T) {
	got, ok := BuscarIndiceEnFecha(serieTest(), fechaTest(2026, time.June, 1))
	if !ok {
		t.Fatal("esperaba encontrar un valor")
	}
	if got != 150 {
		t.Errorf("esperaba 150 (junio), obtuve %v", got)
	}
}

func TestBuscarIndiceEnFecha_DespuesDelUltimoPunto(t *testing.T) {
	// Fecha futura respecto a la serie: debe quedarse con el último punto conocido.
	got, ok := BuscarIndiceEnFecha(serieTest(), fechaTest(2026, time.December, 1))
	if !ok {
		t.Fatal("esperaba encontrar un valor")
	}
	if got != 150 {
		t.Errorf("esperaba 150 (último punto conocido), obtuve %v", got)
	}
}

func TestBuscarIndiceEnFecha_SerieVacia(t *testing.T) {
	_, ok := BuscarIndiceEnFecha(nil, fechaTest(2026, time.January, 1))
	if ok {
		t.Error("esperaba ok=false con serie vacía")
	}
}

func TestUltimoIndice(t *testing.T) {
	ultimo := UltimoIndice(serieTest())
	if ultimo.Valor != 150 {
		t.Errorf("esperaba el último punto (150), obtuve %v", ultimo.Valor)
	}
}
