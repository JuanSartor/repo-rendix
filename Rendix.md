# RENDIX — Portfolio Tracker para el Inversor Argentino

## ¿Qué es Rendix?
Tracker de inversiones enfocado en el inversor argentino retail. Diferencial clave: muestra el **retorno real** vs inflación, no solo el P&L nominal. Lo que ningún broker local hace hoy.

---

## Mi cartera actual (punto de partida para testear el sistema)

| Activo | Monto ARS | % | Tipo | Broker |
|---|---|---|---|---|
| TSLA | $250.000 | 24% | CEDEAR | Cocos Capital |
| YPF | $250.000 | 24% | Acción local | Cocos Capital |
| KO (Coca-Cola) | $250.000 | 24% | CEDEAR | Cocos Capital |
| PAMP (Pampa Energía) | $250.000 | 24% | Acción local | Cocos Capital |
| Reserva (COCOSPPA) | $18.244 | 4% | FCI pesos | Cocos Capital |
| **TOTAL** | **$1.018.244** | 100% | | |

**Brokers:** Cocos Capital + Bull Market Brokers
**Estrategia:** DCA (entrar en 2 tramos por activo, no todo de golpe)

---

## Stack tecnológico

```
Frontend:    React + TypeScript + Vite + Tailwind CSS
Gráficos:    Recharts + TradingView Lightweight Charts
Backend:     Go + Gin
Base datos:  PostgreSQL (Supabase para empezar, gratis)
Auth:        Supabase Auth (JWT, Google OAuth)
Infra:       Vercel (frontend) + Railway (backend)
Alertas:     Telegram Bot API
Pagos:       MercadoPago (ARS) + Stripe (USD/internacional)
```

---

## Estructura del proyecto

```
rendix/
├── backend/
│   ├── cmd/
│   │   └── main.go                  # Entry point, router, CORS
│   ├── internal/
│   │   ├── db/db.go                 # PostgreSQL: schema, queries
│   │   ├── models/models.go         # Structs compartidos
│   │   ├── handlers/handlers.go     # Endpoints HTTP
│   │   └── services/
│   │       ├── market.go            # Precios Yahoo Finance + CCL
│   │       ├── inflation.go         # INDEC + BLS (inflación ARG y USA)
│   │       └── alerts.go            # Telegram Bot + alertas
│   └── go.mod
├── frontend/
│   └── src/
│       ├── App.tsx
│       ├── pages/
│       │   └── Dashboard.tsx
│       ├── components/
│       │   ├── TarjetaResumen.tsx
│       │   ├── TablaCartera.tsx
│       │   ├── ModalOperacion.tsx
│       │   ├── Historial.tsx
│       │   └── GraficoRendimiento.tsx  # (pendiente)
│       ├── hooks/
│       │   └── useApi.ts
│       └── types/index.ts
├── RENDIX.md                        # Este archivo
└── Makefile
```

---

## API Endpoints definidos

| Método | Ruta | Descripción |
|---|---|---|
| GET | /api/health | Health check |
| GET | /api/cartera | Posiciones + P&L en tiempo real |
| POST | /api/compra | Registrar compra |
| POST | /api/venta | Registrar venta |
| GET | /api/historial | Historial de operaciones |
| GET | /api/cotizacion/:ticker | Precio actual de un activo |
| GET | /api/rendimiento/real | P&L ajustado por inflación ARG y USA |
| GET | /api/alertas | Lista de alertas configuradas |
| POST | /api/alertas | Crear nueva alerta de precio |

---

## APIs externas a integrar

| Dato | API | Costo |
|---|---|---|
| Precios USA | Yahoo Finance (yfinance) | Gratis |
| Dólar CCL | criptoya.com | Gratis |
| Inflación ARG | INDEC (api.estadisticas.gob.ar) | Gratis |
| Inflación USA | FRED / BLS | Gratis |
| Earnings calendar | Finnhub | Gratis con límites |
| Noticias financieras | Finnhub News / NewsAPI | Gratis con límites |
| Feriados bursátiles | Tradier API | Gratis |
| Telegram alertas | Telegram Bot API | Gratis |

---

## Módulos a construir (en orden)

### ✅ Fase 1 — Base (levantada y corriendo contra Supabase real)
- [x] CRUD compras/ventas con precio promedio ponderado
- [x] Precios en tiempo real (Yahoo Finance)
- [x] Dólar CCL automático (criptoya.com)
- [x] P&L en ARS y USD por posición
- [x] Historial de operaciones
- [x] Soporte multi-broker (Cocos / Bull Market)
- [x] API REST con Gin + CORS
- [x] Frontend React con dashboard básico
- Nota: DB es Postgres/Supabase desde el arranque (no SQLite), y el driver es
  `pgx` puro Go (no CGO), por decisión tomada durante el armado inicial.

### ✅ Fase 2 — Retorno real (el diferencial más importante)
- [x] Integrar API INDEC para inflación mensual ARG (Series de Tiempo, sin API key)
- [x] Integrar FRED para inflación USA (CPI) — opcional, requiere `FRED_API_KEY` propia
- [x] Calcular rendimiento real vs inflación por posición y total (tab "Retorno Real")
- [x] Reporte mensual: "Ganaste X% nominal pero Y% real"
- [x] Incluir comisiones en el cálculo (registrar comisión por operación)
- Nota: el retorno real usa la fecha de apertura de la posición (primera compra)
  como ancla de inflación; el real en USD requiere que la posición tenga
  `ccl_apertura` (solo posiciones creadas después de esta fase lo tienen).

### 🔲 Fase 3 — Alertas
- [ ] Telegram Bot (un bot por usuario)
- [ ] Alerta cuando activo llega a precio objetivo
- [ ] Alerta cuando cae más del X% en un día
- [ ] Alerta de earnings próximos (Finnhub)
- [ ] Resumen diario automático por Telegram

### 🔲 Fase 4 — Calendario de eventos
- [ ] Feriados bursátiles Argentina y NYSE
- [ ] Fechas de earnings de activos en cartera
- [ ] Eventos macro: reuniones Fed, datos INDEC, elecciones
- [ ] Eventos estacionales: rally navideño, sell in may, etc.

### 🔲 Fase 5 — Multi-usuario (SaaS)
- [ ] Migrar de SQLite a PostgreSQL
- [ ] Auth con Supabase (registro, login, JWT)
- [ ] Cada usuario ve solo su cartera
- [ ] Planes Free / Pro / Premium
- [ ] Importar CSV desde Cocos o Bull Market

### 🔲 Fase 6 — Gráficos y UI avanzada
- [ ] Gráfico evolución de cartera en el tiempo (Recharts)
- [ ] Gráfico de velas por activo (TradingView Lightweight Charts)
- [ ] Gráfico torta distribución de cartera
- [ ] Gráfico comparativo: rendimiento vs inflación vs dólar

### 🔲 Fase 7 — Producto y ventas
- [ ] Landing page
- [ ] Integrar MercadoPago (planes en ARS)
- [ ] Integrar Stripe (planes en USD para Latam)
- [ ] Deploy: Vercel (frontend) + Railway (backend) + Supabase (DB)

---

## Modelo de negocio

| Plan | Precio | Incluye |
|---|---|---|
| **Free** | $0 | 1 broker, hasta 5 activos, sin alertas |
| **Pro** | USD 5-8/mes | Multi-broker, activos ilimitados, alertas Telegram, retorno real vs inflación |
| **Premium** | USD 12-15/mes | Todo Pro + calendario eventos, reportes PDF, importar CSV, API propia |

**Cómo se conecta al broker:** por ahora carga manual + importación de CSV (no hay API pública de Cocos ni Bull Market). El valor está en el procesamiento de los datos, no en la obtención.

---

## Diferencial vs competencia

| Feature | Brokers locales | Apps internacionales | Rendix |
|---|---|---|---|
| Vista unificada multi-broker | ❌ | ❌ | ✅ |
| Retorno real vs inflación ARG | ❌ | ❌ | ✅ |
| Retorno en USD al CCL | ❌ | ❌ | ✅ |
| Alertas por Telegram | ❌ | ❌ | ✅ |
| Calendario eventos Argentina | ❌ | ❌ | ✅ |
| CEDEARs con ratio correcto | ✅ | ❌ | ✅ |
| Comisiones descontadas del P&L | ❌ | Parcial | ✅ |

---

## Canales de distribución

- Reddit: r/merval, r/argentina_invertir
- Twitter/X: comunidad financiera argentina
- Telegram: grupos de inversores
- YouTube: tutoriales = tráfico orgánico
- Product Hunt: visibilidad internacional

---

## Setup inicial — Lo que necesitás instalar

```bash
# 1. Verificar Go (bajar de https://go.dev/dl/)
go version

# 2. Verificar Node (bajar de https://nodejs.org — versión LTS)
node --version
npm --version

# 3. Inicializar el backend
cd rendix/backend
go mod init rendix
go mod tidy   # descarga Gin, SQLite, CORS

# 4. Correr el backend
go run cmd/main.go
# → Servidor en http://localhost:8080

# 5. Inicializar el frontend (en otra terminal)
cd rendix/frontend
npm install
npm run dev
# → App en http://localhost:5173
```

---

## Próximos pasos inmediatos

1. Instalar Go y Node en tu máquina
2. Copiar el código base generado en esta sesión a la carpeta rendix/
3. Levantar backend y frontend localmente
4. Cargar tu cartera real (TSLA, YPF, KO, PAMP) como primer test
5. Arrancar Fase 2: módulo de inflación y retorno real

---

## Contexto del mercado (agosto 2026)

- Sector energético lidera el Merval 2026 (+22% en el año)
- YPF: mejor performance del panel líder (+49% en 2026), split de acción el 4 de agosto
- TSLA: rango 2026 entre $337-$453 USD, catalizadores: Robotaxi y Optimus
- KO: defensiva, dividendos estables, ideal para perfil moderado
- PAMP: upside estimado 59% según analistas, diversificada en energía
- Dólar CCL: referencia para calcular rendimiento real en USD