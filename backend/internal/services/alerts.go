package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"portfolio-tracker/internal/db"
)

// ─── Telegram ────────────────────────────────────────────────────────────────

func telegramConfigurado() bool {
	return os.Getenv("TELEGRAM_BOT_TOKEN") != "" && os.Getenv("TELEGRAM_CHAT_ID") != ""
}

// EnviarMensajeTelegram manda un mensaje (Markdown) al chat configurado en
// TELEGRAM_CHAT_ID usando el bot de TELEGRAM_BOT_TOKEN.
func EnviarMensajeTelegram(texto string) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		return fmt.Errorf("faltan TELEGRAM_BOT_TOKEN / TELEGRAM_CHAT_ID")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	body, err := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"text":       texto,
		"parse_mode": "Markdown",
	})
	if err != nil {
		return err
	}

	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error enviando mensaje a Telegram: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram respondió %d", resp.StatusCode)
	}
	return nil
}

// ─── Chequeo de alertas de precio ─────────────────────────────────────────────

// ChequearAlertas recorre las alertas activas, consulta el precio actual de cada
// ticker y dispara (envía Telegram + marca como disparada) las que cruzaron su
// precio objetivo. Pensada para correr periódicamente desde un goroutine en main.
func ChequearAlertas() {
	if !telegramConfigurado() {
		return
	}
	alertas, err := db.ObtenerAlertasActivas()
	if err != nil {
		log.Printf("⚠️ error obteniendo alertas activas: %v", err)
		return
	}
	if len(alertas) == 0 {
		return
	}

	ccl, err := ObtenerDolarCCL()
	if err != nil {
		log.Printf("⚠️ error obteniendo CCL para chequeo de alertas: %v", err)
		return
	}

	for _, a := range alertas {
		var precioActual float64
		var errPrecio error
		if a.EsCedear {
			_, precioActual, errPrecio = ObtenerPrecioARS(a.Ticker, ccl)
		} else {
			precioActual, errPrecio = ObtenerPrecioUSD(ResolverTickerLocal(a.Ticker) + ".BA")
		}
		if errPrecio != nil || precioActual <= 0 {
			continue
		}

		disparada := (a.Direccion == "ARRIBA" && precioActual >= a.PrecioObjetivo) ||
			(a.Direccion == "ABAJO" && precioActual <= a.PrecioObjetivo)
		if !disparada {
			continue
		}

		direccionTexto := "subió a"
		if a.Direccion == "ABAJO" {
			direccionTexto = "bajó a"
		}
		texto := fmt.Sprintf(
			"🔔 *%s* %s ARS %.2f (tu objetivo: ARS %.2f)",
			a.Ticker, direccionTexto, precioActual, a.PrecioObjetivo,
		)
		if err := EnviarMensajeTelegram(texto); err != nil {
			log.Printf("⚠️ error enviando alerta de %s: %v", a.Ticker, err)
			continue // no la marcamos disparada si no pudimos avisar; se reintenta después
		}
		if err := db.MarcarAlertaDisparada(a.ID); err != nil {
			log.Printf("⚠️ error marcando alerta %d como disparada: %v", a.ID, err)
		}
	}
}

// ─── Resumen diario ────────────────────────────────────────────────────────────

// EnviarResumenDiario manda un mensaje con el P&L nominal total de la cartera.
// Pensada para correr una vez por día desde un goroutine en main.
func EnviarResumenDiario() {
	if !telegramConfigurado() {
		return
	}
	posiciones, err := db.ObtenerPosiciones()
	if err != nil {
		log.Printf("⚠️ error obteniendo posiciones para resumen diario: %v", err)
		return
	}
	if len(posiciones) == 0 {
		return
	}

	ccl, err := ObtenerDolarCCL()
	if err != nil {
		log.Printf("⚠️ error obteniendo CCL para resumen diario: %v", err)
		return
	}
	totales := CalcularPreciosActuales(posiciones, ccl)

	totalPnl := totales.TotalActual - totales.TotalInvertido
	totalPct := 0.0
	if totales.TotalInvertido > 0 {
		totalPct = (totalPnl / totales.TotalInvertido) * 100
	}

	signo := ""
	if totalPnl >= 0 {
		signo = "+"
	}
	texto := fmt.Sprintf(
		"📊 *Resumen diario Rendix*\n\nInvertido: ARS %.0f\nValor actual: ARS %.0f\nP&L: %sARS %.0f (%s%.2f%%)\nDólar CCL: ARS %.2f",
		totales.TotalInvertido, totales.TotalActual, signo, totalPnl, signo, totalPct, ccl,
	)
	if len(totales.SinCotizar) > 0 {
		texto += fmt.Sprintf("\n\n⚠️ Sin cotizar hoy: %v", totales.SinCotizar)
	}

	if err := EnviarMensajeTelegram(texto); err != nil {
		log.Printf("⚠️ error enviando resumen diario: %v", err)
	}
}
