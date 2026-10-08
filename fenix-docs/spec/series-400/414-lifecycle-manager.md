# SPEC-414: Lifecycle Manager

## 1. Executive Summary

El Lifecycle Manager es un orquestador altamente acoplado a la máquina de estados del Host (SPEC-417). Se encarga de la correcta gestión de la Memoria, Conexiones de Red (LSP, AI Gateway), y Caches de disco durante tres fases críticas: la Inicialización (Bootstrap), la Suspensión (Hibernation) y el Apagado Limpio (Graceful Shutdown).
Garantiza que el Compiler Host sea robusto ante corrupciones de estado, caídas repentinas y arranques repetidos.

## 2. Bootstrapping Phase (Initialization)

Cuando el Host recibe el comando `initialize` (por el LSP o CLI), el Lifecycle Manager asume el control:

1. **Environment Verification:** Verifica la validez de los binarios base del runtime Fénix, la versión de Go, y las variables de entorno (`FENIX_WORKSPACE`, `FENIX_CACHE_DIR`).
2. **Crash Recovery Check:** Busca un archivo `.fenix-lock`. Si existe, indica que la ejecución anterior crasheó o terminó en `SIGKILL`. El manager procederá a purgar los cachés volátiles posiblemente corruptos antes de cargar nada.
3. **Cache Hydration:** (SPEC-409) Lee los cachés estructurales y semánticos guardados en disco. Los mapea a memoria (mmap) si es posible para un arranque rápido.
4. **Subsystem Boot:**
   * Levanta el *Event Bus*.
   * Inicia los *Worker Pools*.
   * Llama a `WorkspaceManager.Initialize(rootURI)`.
5. **State Transition:** Mueve la máquina de estados de `Uninitialized` a `WorkspaceLoading` y finalmente a `Ready`.

## 3. Suspension / Hibernation (Optional)

Si el IDE detecta que el usuario ha estado inactivo durante un largo período de tiempo (>30 minutos) o el sistema operativo emite una presión fuerte de memoria, el Lifecycle Manager puede entrar en Hibernación.

1. **Eviction:** Descarga los *Full Semantic Models* (la capa más pesada en memoria) de todos los proyectos que no tienen archivos abiertos. Mantiene solo los Hash Caches y las estructuras superficiales.
2. **Worker Sleep:** Detiene los *Background Queues*.
3. **Wake up:** Al primer evento de red, tecleo o mouse hover, restaura el modelo usando compilación perezosa (Lazy Compilation).

## 4. Graceful Shutdown

El compilador nunca debe cerrarse inmediatamente ante una señal externa normal (ej., `exit` LSP o `SIGTERM`).

La secuencia de apagado toma menos de 2 segundos pero garantiza la consistencia:
1. **Barrier:** El Router deja de aceptar nuevas peticiones y lanza un error `HostShuttingDown` a cualquier nuevo requerimiento del IDE o IA.
2. **Forced Cancellation:** Se invoca el cancelamiento en cascada en todos los *contextos* de todas las colas de trabajo activas (Scheduler P0, P1, P2, P3).
3. **Queue Drain:** Espera a que los Workers respondan a las cancelaciones (generalmente < 50ms).
4. **Cache Persistence:** Serializa los `Hash Caches` limpios al directorio persistente `.fenix/cache`. Escribe las métricas finales al colector de telemetría.
5. **Lock Release:** Borra el `.fenix-lock` para un arranque limpio futuro.
6. **OS Exit:** Llama a `os.Exit(0)`.

## 5. Fallback & Panic Recovery

Dado que el Fénix Compiler Pipeline maneja código arbitrario (potencialmente defectuoso), las *Panics* internas en el Parser o el Binder son posibles pero inaceptables a nivel del Host (botar todo el IDE por un bug del compilador es un fallo catastrófico).

El Lifecycle Manager envuelve la ejecución de los Workers con protecciones de `recover()`:
* Si un subsistema (ej. Semantic Analyzer) tiene un panic, la fase se cancela.
* El Lifecycle Manager captura el stack trace y genera un *Internal Diagnostic Error* que se publica en el IDE sin derribar el proceso daemon.
* Se invalida el cache de ese único documento y se pide reintento en el próximo cambio.
