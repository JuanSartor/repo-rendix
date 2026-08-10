package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

// IndicePrecio es un punto de una serie de índice de precios (fecha del mes → valor del índice).
type IndicePrecio struct {
	Fecha time.Time
	Valor float64
}

// ─── Inflación Argentina (INDEC vía API de Series de Tiempo) ─────────────────

// Serie oficial: IPC Nivel General Nacional, base diciembre 2016 = 100, mensual.
const indecSerieIPC = "148.3_INIVELNAL_DICI_M_26"

type indecResp struct {
	Data [][2]interface{} `json:"data"`
}

var (
	cacheIPC     []IndicePrecio
	cacheIPCTime time.Time
	cacheIPCMu   sync.Mutex

	cacheCPI     []IndicePrecio
	cacheCPITime time.Time
	cacheCPIMu   sync.Mutex
)

const ttlCacheInflacion = 6 * time.Hour

// ObtenerIPCArgentina devuelve la serie histórica del IPC Nacional (INDEC), ordenada ascendente por fecha.
func ObtenerIPCArgentina() ([]IndicePrecio, error) {
	cacheIPCMu.Lock()
	defer cacheIPCMu.Unlock()

	if cacheIPC != nil && time.Since(cacheIPCTime) < ttlCacheInflacion {
		return cacheIPC, nil
	}

	url := fmt.Sprintf(
		"https://apis.datos.gob.ar/series/api/series/?ids=%s&limit=1000&format=json",
		indecSerieIPC,
	)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error consultando INDEC: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data indecResp
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("error parseando INDEC: %w", err)
	}

	serie := make([]IndicePrecio, 0, len(data.Data))
	for _, punto := range data.Data {
		fechaStr, ok := punto[0].(string)
		if !ok {
			continue
		}
		valor, ok := punto[1].(float64)
		if !ok {
			continue
		}
		fecha, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			continue
		}
		serie = append(serie, IndicePrecio{Fecha: fecha, Valor: valor})
	}
	sort.Slice(serie, func(i, j int) bool { return serie[i].Fecha.Before(serie[j].Fecha) })

	if len(serie) == 0 {
		return nil, fmt.Errorf("INDEC no devolvió datos de IPC")
	}

	cacheIPC = serie
	cacheIPCTime = time.Now()
	return serie, nil
}

// ─── Inflación USA (FRED — CPI para todos los consumidores urbanos) ──────────

// Serie oficial FRED: CPI-U, todos los ítems, ajustado estacionalmente, base 1982-84=100.
const fredSerieCPI = "CPIAUCSL"

type fredResp struct {
	Observations []struct {
		Date  string `json:"date"`
		Value string `json:"value"`
	} `json:"observations"`
}

// ObtenerCPIUSA devuelve la serie histórica del CPI-U (FRED), ordenada ascendente por fecha.
// Requiere la variable de entorno FRED_API_KEY (gratis en fred.stlouisfed.org/docs/api/api_key.html).
func ObtenerCPIUSA() ([]IndicePrecio, error) {
	apiKey := os.Getenv("FRED_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("falta FRED_API_KEY: conseguí una gratis en https://fred.stlouisfed.org/docs/api/api_key.html")
	}

	cacheCPIMu.Lock()
	defer cacheCPIMu.Unlock()

	if cacheCPI != nil && time.Since(cacheCPITime) < ttlCacheInflacion {
		return cacheCPI, nil
	}

	url := fmt.Sprintf(
		"https://api.stlouisfed.org/fred/series/observations?series_id=%s&api_key=%s&file_type=json&frequency=m&sort_order=asc",
		fredSerieCPI, apiKey,
	)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error consultando FRED: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FRED respondió %d: %s", resp.StatusCode, string(body))
	}

	var data fredResp
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("error parseando FRED: %w", err)
	}

	serie := make([]IndicePrecio, 0, len(data.Observations))
	for _, obs := range data.Observations {
		if obs.Value == "." {
			continue // FRED usa "." para valores faltantes
		}
		valor, err := strconv.ParseFloat(obs.Value, 64)
		if err != nil {
			continue
		}
		fecha, err := time.Parse("2006-01-02", obs.Date)
		if err != nil {
			continue
		}
		serie = append(serie, IndicePrecio{Fecha: fecha, Valor: valor})
	}

	if len(serie) == 0 {
		return nil, fmt.Errorf("FRED no devolvió datos de CPI")
	}

	cacheCPI = serie
	cacheCPITime = time.Now()
	return serie, nil
}

// ─── Helpers ───────────────────────────────────────────────────────────────

// BuscarIndiceEnFecha devuelve el valor del índice correspondiente al mes de la fecha dada
// (o el más cercano disponible: el último punto anterior o igual a esa fecha; si no hay
// ninguno anterior, el primero disponible). La serie debe estar ordenada ascendente.
func BuscarIndiceEnFecha(serie []IndicePrecio, fecha time.Time) (float64, bool) {
	if len(serie) == 0 {
		return 0, false
	}
	objetivo := time.Date(fecha.Year(), fecha.Month(), 1, 0, 0, 0, 0, time.UTC)

	mejor := serie[0]
	encontrado := false
	for _, p := range serie {
		pMes := time.Date(p.Fecha.Year(), p.Fecha.Month(), 1, 0, 0, 0, 0, time.UTC)
		if !pMes.After(objetivo) {
			mejor = p
			encontrado = true
		} else {
			break
		}
	}
	if !encontrado {
		// La fecha pedida es anterior al inicio de la serie: usamos el primer punto disponible.
		return serie[0].Valor, true
	}
	return mejor.Valor, true
}

// UltimoIndice devuelve el punto más reciente de la serie (debe estar ordenada ascendente).
func UltimoIndice(serie []IndicePrecio) IndicePrecio {
	return serie[len(serie)-1]
}
