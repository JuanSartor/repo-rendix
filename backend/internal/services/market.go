package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Ratios CEDEAR → 1 acción USA (cuántos CEDEARs equivalen a 1 acción)
// Actualizar periódicamente desde BYMA
var CedearRatios = map[string]float64{
	"AAPL":  10,
	"GOOGL": 10,
	"MSFT":  10,
	"AMZN":  10,
	"TSLA":  10,
	"SPY":   10,
	"QQQ":   10,
	"BRKB":  10,
	"KO":    10,
	"META":  10,
	"NVDA":  10,
	"NFLX":  10,
}

var httpClient = &http.Client{Timeout: 8 * time.Second}

// ─── Cache en memoria de precios y CCL ────────────────────────────────────────
// Evita golpear Yahoo/criptoya en cada request de /api/cartera (que ya de por sí
// paraleliza el fetch por posición) y acorta la ventana en la que un fetch lento
// bloquea el request completo.

const ttlCachePrecio = 30 * time.Second

type precioCacheEntry struct {
	precio    float64
	err       error
	timestamp time.Time
}

var (
	precioCacheUSD = map[string]precioCacheEntry{}
	precioCacheMu  sync.Mutex

	cclCache   precioCacheEntry
	cclCacheMu sync.Mutex
)

// TickerLocalAlias mapea el ticker "coloquial" al símbolo real que usa Yahoo
// Finance para acciones locales en BYMA, cuando difieren (ej: YPF cotiza como YPFD.BA).
var TickerLocalAlias = map[string]string{
	"YPF": "YPFD",
}

// ResolverTickerLocal devuelve el símbolo BYMA correcto para consultar Yahoo Finance.
func ResolverTickerLocal(ticker string) string {
	if alias, ok := TickerLocalAlias[ticker]; ok {
		return alias
	}
	return ticker
}

// ─── Dólar CCL ────────────────────────────────────────────────────────────────

type criptoyaResp map[string]struct {
	Ask float64 `json:"ask"`
	Bid float64 `json:"bid"`
}

// ObtenerDolarCCL devuelve el CCL, cacheado por ttlCachePrecio para no golpear
// criptoya.com en cada llamada dentro de la misma ventana de tiempo.
func ObtenerDolarCCL() (float64, error) {
	cclCacheMu.Lock()
	if cclCache.timestamp.IsZero() == false && time.Since(cclCache.timestamp) < ttlCachePrecio {
		precio, err := cclCache.precio, cclCache.err
		cclCacheMu.Unlock()
		return precio, err
	}
	cclCacheMu.Unlock()

	precio, err := obtenerDolarCCLSinCache()

	cclCacheMu.Lock()
	cclCache = precioCacheEntry{precio: precio, err: err, timestamp: time.Now()}
	cclCacheMu.Unlock()

	return precio, err
}

func obtenerDolarCCLSinCache() (float64, error) {
	resp, err := httpClient.Get("https://criptoya.com/api/dolar")
	if err != nil {
		return 0, fmt.Errorf("error consultando CCL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data criptoyaResp
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("error parseando CCL: %w", err)
	}

	if ccl, ok := data["ccl"]; ok && ccl.Ask > 0 {
		return ccl.Ask, nil
	}
	if blue, ok := data["blue"]; ok && blue.Ask > 0 {
		return blue.Ask, nil
	}
	return 0, fmt.Errorf("no se encontró cotización CCL ni blue")
}

// ─── Precios USA via Yahoo Finance ────────────────────────────────────────────

type yahooResp struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
			} `json:"meta"`
		} `json:"result"`
		Error *struct {
			Code string `json:"code"`
		} `json:"error"`
	} `json:"chart"`
}

// ObtenerPrecioUSD devuelve el precio de un ticker en USD, cacheado por ttlCachePrecio.
// Es la función que más se llama en paralelo (una por posición), así que el cache
// es lo que evita golpear Yahoo N veces cuando varias posiciones comparten ticker
// o cuando el frontend refresca cada 60s.
func ObtenerPrecioUSD(ticker string) (float64, error) {
	precioCacheMu.Lock()
	if entry, ok := precioCacheUSD[ticker]; ok && time.Since(entry.timestamp) < ttlCachePrecio {
		precioCacheMu.Unlock()
		return entry.precio, entry.err
	}
	precioCacheMu.Unlock()

	precio, err := obtenerPrecioUSDSinCache(ticker)

	precioCacheMu.Lock()
	precioCacheUSD[ticker] = precioCacheEntry{precio: precio, err: err, timestamp: time.Now()}
	precioCacheMu.Unlock()

	return precio, err
}

func obtenerPrecioUSDSinCache(ticker string) (float64, error) {
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1d",
		ticker,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	// Yahoo bloquea requests sin User-Agent con "Edge: Too Many Requests"
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error consultando Yahoo Finance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data yahooResp
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("error parseando Yahoo Finance: %w", err)
	}
	if data.Chart.Error != nil {
		return 0, fmt.Errorf("yahoo error: %s", data.Chart.Error.Code)
	}
	if len(data.Chart.Result) == 0 {
		return 0, fmt.Errorf("sin resultados para %s", ticker)
	}
	return data.Chart.Result[0].Meta.RegularMarketPrice, nil
}

// ObtenerPrecioARS calcula el precio en ARS de un CEDEAR dado el CCL
func ObtenerPrecioARS(ticker string, ccl float64) (precioUSD, precioARS float64, err error) {
	ratio, ok := CedearRatios[ticker]
	if !ok {
		ratio = 10 // default
	}

	precioUSD, err = ObtenerPrecioUSD(ticker)
	if err != nil {
		return 0, 0, err
	}

	precioARS = (precioUSD / ratio) * ccl
	return precioUSD, precioARS, nil
}
