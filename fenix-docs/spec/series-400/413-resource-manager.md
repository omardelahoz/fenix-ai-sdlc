# SPEC-413: Resource Manager

## 1. Executive Summary

El compilador Fénix puede operar en entornos altamente restringidos (ej. un contenedor Docker en un CI/CD server, o un Web IDE con 1GB de RAM máxima) o en estaciones de trabajo de desarrollador de gama alta (64GB RAM, 32 Cores).
El *Resource Manager* es un subsistema guardián encargado de monitorizar el consumo real del Sistema Operativo y aplicar contramedidas defensivas de *Backpressure* o *Shedding* para prevenir que el Host provoque OOM (Out-of-Memory) panics o asfixie el sistema del usuario.

## 2. Memory Pressure Limits

El Host arranca evaluando los límites reales (CGroups en Linux, MemoryStatus en Windows, o cuotas impuestas por CLI). Define tres umbrales de memoria:

1. **Target Watermark (ej. 50% de Cuota):** Comportamiento normal. Los cachés se llenan, el Garbage Collector de Go funciona a su ritmo estándar.
2. **High Watermark (ej. 80% de Cuota):** Entra en modo *Defensivo*. El Resource Manager ordena al `CacheManager` (SPEC-409) reducir agresivamente (Trim) los cachés semánticos de proyectos inactivos.
3. **Critical Watermark (ej. 95% de Cuota):** Entra en modo *Supervivencia*. Se bloquean todas las nuevas peticiones de Agentes de IA. Se fuerzan ejecuciones sincrónicas de `runtime.GC()`. Si es necesario, el *Snapshot Manager* descarta (Evict) el estado inmutable por completo obligando a recompilar desde cero en el próximo request.

## 3. CPU Throttling and Yielding

Si el usuario define `MaxCPU = 50%`, el Resource Manager instruye al Scheduler y al Worker Pool:
* Los ciclos de *Background Queues* introducen `time.Sleep()` artificiales en cada interrupción de contexto para ceder la CPU al OS.
* Las compilaciones tardarán más, pero los ventiladores de la laptop no girarán al máximo.

## 4. Backpressure a Agentes de IA

A diferencia del LSP del IDE, que usualmente tiene un usuario humano (lento), los Agentes de IA vía el `AI Gateway` pueden emitir ráfagas de miles de peticiones `GoToDefinition` por segundo para analizar el código base.

Si el Resource Manager detecta sobrecarga (Saturación de Cola o High Watermark), envía comandos de **Backpressure**:
* Devuelve inmediatamente errores `HTTP 429 Too Many Requests` a los clientes de IA.
* Aumenta el *Timeout* forzado en la cola del `Execution Engine` para tareas P2/P3.

## 5. Interface Contract

```go
package resource

import "context"

type PressureLevel int

const (
    Normal PressureLevel = iota
    High
    Critical
)

type ResourceManager interface {
    // Monitoreo en tiempo real.
    CurrentMemoryPressure() PressureLevel
    
    // Configuración dinámica (ej. el IDE envía un aviso de batería baja).
    SetLimits(config ResourceLimits)
    
    // Suscripción de subsistemas (como CacheManager) a alertas.
    OnPressureChange(handler func(level PressureLevel))
}

type ResourceLimits struct {
    MaxMemoryBytes uint64
    MaxCPUPercent  float32
}
```
