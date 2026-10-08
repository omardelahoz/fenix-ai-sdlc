# SPEC-412: Worker Pool Architecture

## 1. Executive Summary

El *Worker Pool* es el mecanismo físico subyacente que ejecuta las tareas encoladas por el *Task Execution Engine* (SPEC-404). Su objetivo es mantener el Compiler Host dentro de un perfil de uso de recursos predecible (CPU y Memoria), limitando el número máximo de goroutines activas que realizan trabajo computacionalmente denso, evitando así la saturación de los cores del procesador y los costosos *context switches* a nivel del Sistema Operativo.

## 2. Dynamic Sizing

A diferencia de los servidores web tradicionales que lanzan una goroutine por cada request HTTP, el Compiler Host Fénix computa transformaciones intensas de AST (Abstract Syntax Trees). Por ende, usa *Fixed Size Worker Pools* que escalan dinámicamente según el hardware subyacente.

El número de Workers está directamente ligado a la topología del CPU:
* **CPU Workers:** `NumWorkers = max(2, runtime.NumCPU() - 1)`
  * Reservamos siempre al menos 1 core físico para el Sistema Operativo y el sub-hilo del Router LSP para asegurar que la UI nunca se congele.
* **IO Workers:** `NumWorkers = runtime.NumCPU() * 4`
  * Las tareas de I/O (red, disco) pasan casi todo su tiempo bloqueadas. Un pool mucho mayor permite alta concurrencia en descargas o file-reads sin consumir ciclos de CPU.

## 3. Worker Lifecycle & Supervision

Cada Worker en el Pool es una goroutine longeva (long-lived) supervisada por el *Lifecycle Manager* (SPEC-414).

### The Worker Loop
```go
func (w *Worker) start() {
    for {
        select {
        case task := <-w.queue:
            w.executeWithPanicRecovery(task)
        case <-w.ctx.Done():
            // Graceful shutdown signal received
            return
        }
    }
}
```

### Panic Recovery & Respawning
Si la función `task()` lanza un `panic()` (ej., *slice bounds out of range* en un plugin defectuoso):
1. El `executeWithPanicRecovery` atrapa la excepción.
2. Emite la telemetría correspondiente al *Diagnostics Pipeline*.
3. El Worker continúa iterando. **La goroutine del Worker no muere**, lo que previene el costo de *teardown* y *respawn*.

## 4. Work Stealing Algorithm

Si el *Realtime Pool* está sobrepasado (ej., un ráfaga de peticiones del IDE) pero el *Background Pool* está inactivo, Fénix NO reasigna workers (debido al aislamiento arquitectónico), sino que implementa **Work Stealing**:
* Un Worker del *Background Pool* sin trabajo puede espiar la cola del *Realtime Pool*.
* Si la cola de Realtime tiene tareas esperando más de `X` ms, el worker roba la tarea y la procesa.
* Esto maximiza la utilización de la CPU manteniendo los límites de prioridad (SPEC-403).

## 5. Interface Contract

```go
package pool

import "context"

type WorkerPool interface {
    // Inicia N workers según el perfil de Hardware.
    Start(ctx context.Context)
    
    // Cambia dinámicamente el tamaño del pool (ej. si el SO impone límites).
    Resize(numWorkers int)
    
    // Obtiene telemetría (cuántos workers están Idle vs Busy).
    Stats() PoolStats
}
```
