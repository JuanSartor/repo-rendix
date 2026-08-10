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

---

# Mejoras sugeridas por Fable

Auditoría del estado actual (post Fase 1 y 2) con todo lo que encontré trabajando en el código, ordenado por prioridad. Lo marcado 🔴 lo arreglaría antes de sumar features nuevas; 🟡 antes de tener usuarios reales; 🟢 es mejora continua.

## 1. Deuda técnica del backend

### 🔴 Precisión del cálculo de retorno real con DCA
Hoy el retorno real ancla la inflación en `creado_en` de la posición (fecha de la **primera** compra). Con estrategia DCA en tramos esto es impreciso: si comprás TSLA en enero y reforzás en junio, todo el capital se ajusta por inflación desde enero, sobreestimando la inflación sufrida por el segundo tramo.
- **Fix:** ajustar cada operación de compra individualmente usando su `fecha_opera` (los datos ya están en la tabla `operaciones`, es solo cambiar el cálculo en `GetRendimientoReal` para iterar operaciones en vez de posiciones).
- Lo mismo aplica al CCL: hoy se guarda `ccl_apertura` solo en la primera compra; lo correcto es guardar el CCL en **cada** operación (columna `ccl_operacion` en `operaciones`) y calcular el costo en USD como suma de tramos.

### 🔴 Errores de precios silenciosos
En `calcularPreciosActuales` los errores de Yahoo se descartan (`_`). Si Yahoo falla, la posición aparece con precio 0 y P&L -100% (nos pasó en vivo con el bloqueo por User-Agent). Es el peor tipo de error: parece dato real.
- **Fix:** agregar campo `precio_error bool` o `precio_stale` a la respuesta; el frontend muestra "sin cotización" en gris en vez de una pérdida falsa. Nunca mostrar -100% por un fetch fallido.

### 🔴 Fetch de precios secuencial
`GetCartera` consulta Yahoo una posición por vez (~200-500ms c/u). Con 10 activos son varios segundos por request, y cada refresh del frontend repite todo.
- **Fix corto:** goroutines con `golang.org/x/sync/errgroup` (paraleliza) + cache en memoria de precios con TTL de 30-60s (mismo patrón que ya usa `inflation.go`).
- **Fix largo:** un job en background que refresca precios cada minuto y los handlers leen solo del cache. Esto además te prepara para las alertas de Fase 3, que necesitan exactamente ese loop.

### 🟡 Venta no calcula P&L realizado
`RegistrarVenta` descuenta cantidad pero no registra contra qué precio promedio se vendió. No podés responder "¿cuánto gané con la venta de AAPL?" ni armar el reporte anual para impuestos (CEDEARs pagan ganancias).
- **Fix:** guardar `precio_prom_al_vender` y `pnl_realizado_ars` en la operación de venta. La tabla de historial después puede mostrar P&L realizado por operación y acumulado del año.

### 🟡 Ratios CEDEAR hardcodeados y todos en 10
El mapa `CedearRatios` tiene 12 tickers, todos con ratio 10 — pero los ratios reales varían (KO es 5:1, NVDA cambió post-split, MELI es 60:1, etc.). Un ratio equivocado da precio ARS equivocado sin que nada falle.
- **Fix:** cargar el listado real de ratios de BYMA (lo publica en Excel/CSV) a una tabla `cedear_ratios` en la DB, con fecha de vigencia. Actualizarlo es un INSERT, no un deploy. Mismo criterio para `TickerLocalAlias` (hoy solo tiene YPF→YPFD).

### 🟡 Producción: flags de Gin y CORS
- `gin.SetMode(gin.ReleaseMode)` cuando `GIN_MODE` no esté seteado a debug explícitamente.
- `r.SetTrustedProxies(nil)` (el warning que aparece en cada arranque).
- Orígenes CORS desde variable de entorno (`ALLOWED_ORIGINS`), no hardcodeados a localhost — lo vas a necesitar sí o sí al deployar a Vercel.
- Timeout/contexto en los handlers: hoy un Yahoo colgado retiene la conexión los 8s completos del `httpClient` por cada activo.

### 🟢 Resiliencia de fuentes de datos
Cada dato externo tiene una sola fuente. Sugerencia de fallbacks en orden:
- **CCL:** criptoya → dolarapi.com → último valor cacheado con timestamp visible ("CCL de hace 2hs").
- **IPC:** API Series de Tiempo → CSV directo de INDEC (`indec.gob.ar/ftp/cuadros/economia/serie_ipc_divisiones.csv`) → último cacheado.
- Persistir el último valor bueno de cada fuente en la DB para sobrevivir reinicios (el cache actual es solo memoria).

## 2. Testing (hoy: cero tests)

El proyecto no tiene ni un test. Antes de Fase 3 (alertas que disparan mensajes reales) conviene tener al menos la capa de cálculo cubierta, porque es donde un bug es plata mal informada.

### Backend (Go, en orden de valor)
1. **Unit tests de cálculo puro** — lo más urgente y lo más fácil:
   - Precio promedio ponderado con compras sucesivas y comisiones (`RegistrarCompra`): casos con comisión 0, comisión alta, compra que promedia hacia arriba/abajo.
   - `BuscarIndiceEnFecha`: fecha exacta, fecha entre meses, fecha anterior al inicio de la serie, serie vacía.
   - Retorno real: cartera sintética con inflación conocida (índice 100→200 = 100% inflación) y verificar nominal vs real a mano.
   - Venta que excede posición, venta exacta, ticker inexistente.
2. **Tests de handlers con `httptest`** + interfaces para mockear `services` (hoy los handlers llaman funciones concretas del paquete; extraer una interfaz `MarketData` y otra `InflacionData` las hace mockeables e inyectables).
3. **DB tests** contra Postgres real en CI (contenedor `postgres:16` en GitHub Actions) — el protocolo simple de PgBouncer ya está encapsulado en `Init`, así que los tests corren contra Postgres pelado sin tocar nada.

### Frontend
- **Vitest + React Testing Library:** ModalOperacion (validación de campos, cálculo del total con comisión en compra vs venta), TablaCartera (posición sin precio no muestra -100%), formateo es-AR.
- **Playwright e2e** (ya quedó instalado en esta máquina): flujo compra → aparece en cartera → aparece en historial → retorno real la incluye. Correrlo contra un backend con DB de test.

### CI (GitHub Actions, gratis en repo público/privado chico)
```yaml
# .github/workflows/ci.yml — esqueleto
- backend: go vet + go test ./... (con servicio postgres)
- frontend: npm run build (ya typecheckea) + vitest
- e2e: opcional por ahora, correr solo en main
```
Con esto cada push a una rama valida solo. Es una tarde de trabajo y evita romper `main` para siempre.

## 3. Seguridad (crítico antes de Fase 5 / usuarios reales)

- **Hoy la API es abierta:** cualquiera en tu red le pega a `localhost:8080` y registra operaciones. Aceptable en localhost, inaceptable deployado. La auth de Supabase (Fase 5) tiene que llegar **antes** que el deploy público, no después — aunque sea un API token fijo como puente.
- **Rate limiting** en el backend (ej. `ulule/limiter` o middleware simple por IP) antes de exponer a internet.
- **Validación de inputs:** ticker hoy acepta cualquier string; limitar a `[A-Z0-9.]{1,10}` corta basura en la DB y qualquier intento de inyección vía query de historial.
- Cuando actives multi-usuario: RLS en Supabase aunque no uses la Data API (defensa en profundidad), y `user_id` en todas las tablas desde el primer día de Fase 5 — migrar datos sin dueño después es doloroso.
- El `.env` nunca versionado ya está bien resuelto; sumar rotación de la password de la DB si alguna vez se pegó en un chat/log (ya pasó una vez en esta sesión con un log que borré — rotarla no cuesta nada y cierra el tema).

## 4. Producto y UX

- **La feature estrella está escondida.** El retorno real es EL diferencial y vive en un tab secundario. Sugerencia: la tarjeta "P&L Total" del dashboard principal debería mostrar nominal Y real juntos ("+1193% nominal / +890% real"). Que el usuario vea el diferencial sin hacer click.
- **El resumen "Ganaste X% nominal pero Y% real" es oro para compartir.** Botón "compartir" que genera una imagen (canvas → PNG) con ese texto + gráfico mini + logo Rendix. Cada tweet de un usuario es marketing gratis. Este es tu growth loop natural.
- **Fecha de operación editable** en el modal: hoy toda compra se registra "ahora", así que no podés cargar tu historial real de meses pasados — y sin fechas reales el retorno real da mal. Es un input date, pero desbloquea la carga de carteras existentes (y con eso, el onboarding de cualquier usuario nuevo).
- **Importar CSV de Cocos/Bull Market** (está en Fase 5, yo lo subiría a Fase 3-4): nadie con 50 operaciones las va a cargar a mano. Es LA fricción de onboarding.
- Estados de carga por posición (skeleton) en vez de spinner global; el refresh de 60s hoy parpadea toda la tabla.
- PWA básica (manifest + service worker): "instalá Rendix en tu teléfono" sin pagar App Store.

## 5. Deploy (checklist concreto para cuando toque)

1. **Backend en Railway:** `PORT` ya se lee de env ✓, falta `GIN_MODE=release`, `ALLOWED_ORIGINS`, y el `DATABASE_URL` del pooler (¡el de puerto 6543, no el directo — Railway puede no tener IPv6 hacia Supabase igual que tu casa!).
2. **Frontend en Vercel:** `BASE` de la API hoy está hardcodeado a `localhost:8080` en `useApi.ts` — moverlo a `import.meta.env.VITE_API_URL` antes del primer deploy.
3. **Healthcheck:** `/api/health` ya existe ✓ — configurarlo en Railway para restart automático.
4. Dominio: `rendix.ar` o `.com.ar` (baratos) > `.com`. Chequear disponibilidad temprano.
5. Supabase free se pausa tras 1 semana sin tráfico: el job de precios de la mejora 1.3 lo mantiene vivo de paso.

## 6. Cómo venderlo (plan de salida al mercado)

### Posicionamiento
Una sola frase, siempre la misma: **"Tu broker te dice cuánto ganaste. Rendix te dice cuánto ganaste de verdad."** Todo el marketing sale de ahí. No competís con Cocos ni IOL — sos la capa de verdad arriba de cualquier broker.

### Secuencia de lanzamiento (en orden, no en paralelo)
1. **Landing con lista de espera** (1 día de laburo): la frase, un screenshot del tab Retorno Real con números reales, y un campo de email. Medí conversión antes de escribir una línea más de código.
2. **Beta cerrada gratis** con 20-50 usuarios de r/merval y Twitter finanzas arg. Pedí a cambio: feedback + permiso para usar sus resúmenes anonimizados ("un usuario de Rendix descubrió que su +45% era -12% real").
3. **Contenido que se comparte solo:** el reporte mensual "nominal vs real" como imagen. Publicá vos mismo el de tu cartera real todos los meses. La transparencia de mostrar tus propias pérdidas reales genera más confianza que cualquier ad.
4. **SEO de calculadora:** una página pública `/calculadora` — "¿cuánto vale hoy tu inversión de $X de hace N meses ajustada por inflación?" — sin registro. Es el tipo de página que rankea para "inflación inversión argentina" y derrama registros.
5. **Product Hunt / Hacker News recién cuando haya onboarding pulido** (CSV import + fechas editables). Un lanzamiento con fricción de carga manual quema la única primera impresión.

### Pricing (ajuste al modelo actual)
- El plan Free con "hasta 5 activos" está bien, pero **el retorno real tiene que estar en Free** aunque sea limitado (solo total, sin desglose por posición). Es el gancho; nadie paga por lo que no probó.
- Pro: anclá el precio en ARS mentalmente comparable — "menos que una pizza por mes" funciona mejor en Argentina que "USD 5".
- Cobrá en ARS por MercadoPago desde el día uno; Stripe/USD puede esperar a que exista demanda de afuera.
- Anual con descuento fuerte (2 meses gratis): en un país con inflación, cobrar por adelantado es doblemente valioso para vos.

### Métricas mínimas desde el día uno
Retención semanal (¿vuelven a mirar su cartera?), % de usuarios que cargan >3 operaciones (activación), y cuántos comparten el reporte. Con esas tres sabés si hay producto. Plausible o PostHog free tier alcanza.

## 7. Orden sugerido para continuar

1. **Ahora (rama `feature/fase-3-alertas` ya creada):** antes de las alertas, meter los tres 🔴 de arriba — errores de precio visibles, fetch paralelo+cache, y retorno real por operación. Las alertas van a leer precios constantemente; conviene que esa capa sea sólida primero. (2-3 días)
2. **Fase 3 reducida:** solo alerta de precio objetivo + resumen diario por Telegram. Earnings/Finnhub y "cayó X% en el día" a Fase 3.5 — un bot que hace poco y funciona vale más que uno ambicioso a medias. (3-4 días)
3. **Testing + CI** en paralelo con la Fase 3 (los unit tests de cálculo no bloquean nada). (1-2 días)
4. **Fechas editables + import CSV** — desbloquea onboarding real, incluida TU cartera verdadera que sigue sin cargar. (2-3 días)
5. **Landing + lista de espera** — mientras tanto, medí interés. (1 día)
6. Fase 5 (auth) → deploy → beta cerrada. Los gráficos (Fase 6) pueden esperar: nadie deja de usar un tracker por falta de gráfico de velas, pero sí por no poder cargar su cartera.

### Nota final
La cartera cargada en la DB sigue siendo la simulada de prueba. Cuando agregues fechas editables (punto 4), cargá tu cartera real con las fechas verdaderas de compra — vas a ser tu propio primer caso de test del retorno real con DCA, que es exactamente el caso que el cálculo actual maneja peor.