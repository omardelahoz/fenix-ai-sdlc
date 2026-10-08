# SPEC-416: Telemetry & Metrics

## 1. Executive Summary

La observabilidad es un pilar arquitectónico de Fénix. Dado que el compilador opera como un servicio de larga duración (daemon) con latencias exigidas del orden de milisegundos, es imposible optimizarlo a ciegas.

El subsistema de *Telemetry & Metrics* recolecta contadores, histogramas y trazas de tiempo de manera *Zero-Allocation* (sin presionar al Garbage Collector), permitiendo auditar la salud del Compiler Host en tiempo real.

## 2. Métricas Clave

Las siguientes métricas son consideradas P0 y deben estar siempre instrumentadas:

### 2.1. Compilation Metrics
* **Total Compilation Time (ms):** Tiempo desde el `DidChange` hasta que el `SemanticAnalyzer` termina. Se mide en histogramas (P50, P90, P99).
* **Incremental Hit Rate (%):** Cuántas compilaciones reutilizaron más del 95% de sus cachés versus una compilación "fría" (Cold Start).
* **AST Node Reuse Rate (%):** Proporción exacta de nodos sintácticos reusados en un Copy-on-Write contra nodos recién parseados.

### 2.2. Resource Metrics
* **Memory Resident Set Size (RSS):** Memoria total cedida por el SO.
* **Heap Allocated / In Use:** Memoria real de los objetos de Go.
* **Stale Snapshots Count:** Cuántos Snapshots viejos siguen vivos en memoria (indicador de agentes de IA atascados).

### 2.3. Queue & Scheduler Metrics
* **Realtime Queue Wait Time (ms):** Cuánto tardan las tareas P0/P1 en salir de la cola. Si el P99 sube de 50ms, el sistema se siente "lag".
* **Work Stealing Invocations:** Cuántas veces un Worker de Background robó trabajo de Realtime.

## 3. Zero-Allocation Observability

Una trampa común en Go es instrumentar el código creando objetos, lo que provoca llamadas al GC que arruinan la latencia que se intenta medir.

1. **Uso de Atomics:** Todos los contadores de Fénix usan `sync/atomic` (ej. `atomic.AddUint64`), garantizando tiempos de microsegundos sin locks.
2. **Context-bound Tracing:** Los tiempos de inicio se inyectan en el `context.Context` (ej. `ctx.Value(StartTime)`), y el tiempo final se calcula al hacer `defer telemetry.Record(ctx)`.

## 4. Exposición (Exporters)

Las métricas nunca deben afectar la experiencia del usuario (Privacy by Default).

1. **Local Prometheus Endpoint:** Si Fénix corre con `--metrics-port=9090`, levanta un servidor HTTP local que expone las métricas para consumo de herramientas de profiling locales.
2. **LSP Extension:** Envía telemetría agregada como notificaciones de extensión LSP al IDE (`fenix/telemetry`), útil para que los desarrolladores de Fénix visualicen los logs en la pestaña de Output de VSCode.

## 5. Interface Contract

```go
package telemetry

import "context"

type MetricsPipeline interface {
    // IncrementCounters
    IncCacheHit()
    IncCacheMiss()
    
    // Histograms
    ObserveLatency(phase string, durationMs float64)
    
    // Scoped Recording
    // Uso: defer metrics.RecordPhase(ctx, "Parser")()
    RecordPhase(ctx context.Context, phase string) func()
}
```
