package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"portfolio-tracker/internal/db"
	"portfolio-tracker/internal/handlers"
	"portfolio-tracker/internal/services"
)

func main() {
	// Carga variables desde .env si existe (no falla si no está)
	_ = godotenv.Load()

	// Base de datos (Postgres / Supabase)
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("Falta la variable de entorno DATABASE_URL (connection string de Supabase)")
	}
	if err := db.Init(connString); err != nil {
		log.Fatalf("Error inicializando DB: %v", err)
	}

	// Router
	r := gin.Default()

	// CORS: permite requests desde el frontend React (localhost:5173)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Rutas
	api := r.Group("/api")
	{
		api.GET("/health",           handlers.Health)
		api.GET("/cartera",          handlers.GetCartera)
		api.POST("/compra",          handlers.PostCompra)
		api.POST("/venta",           handlers.PostVenta)
		api.GET("/historial",        handlers.GetHistorial)
		api.GET("/cotizacion/:ticker", handlers.GetCotizacion)
		api.GET("/rendimiento/real", handlers.GetRendimientoReal)
		api.GET("/alertas",          handlers.GetAlertas)
		api.POST("/alertas",         handlers.PostAlertas)
		api.DELETE("/alertas/:id",   handlers.DeleteAlerta)
	}

	iniciarJobsAlertas()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Portfolio Tracker corriendo en http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}

// iniciarJobsAlertas arranca dos goroutines en background: una que chequea
// alertas de precio cada 5 minutos, y otra que manda el resumen diario por
// Telegram una vez al día a la hora configurada en RESUMEN_HORA (default 9,
// hora del servidor). Si TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID no están
// configurados, los jobs igual corren pero no hacen nada (fail silently,
// se loguea solo si Telegram estaba configurado y falló el envío).
func iniciarJobsAlertas() {
	horaResumen := 9
	if v := os.Getenv("RESUMEN_HORA"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h >= 0 && h <= 23 {
			horaResumen = h
		}
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			services.ChequearAlertas()
			<-ticker.C
		}
	}()

	go func() {
		ultimoEnvio := ""
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			ahora := time.Now()
			hoy := ahora.Format("2006-01-02")
			if ahora.Hour() == horaResumen && ultimoEnvio != hoy {
				services.EnviarResumenDiario()
				ultimoEnvio = hoy
			}
			<-ticker.C
		}
	}()

	log.Printf("🔔 Jobs de alertas iniciados (chequeo cada 5min, resumen diario a las %d:00)", horaResumen)
}
