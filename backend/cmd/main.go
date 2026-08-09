package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"portfolio-tracker/internal/db"
	"portfolio-tracker/internal/handlers"
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
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Portfolio Tracker corriendo en http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
