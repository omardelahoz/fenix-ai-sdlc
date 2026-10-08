# SPEC-406: Background Services & Daemon Orchestration

## 1. Executive Summary
El Compiler Host de Fénix ejecuta un volumen significativo de trabajo crítico que no es desencadenado directamente por las pulsaciones del usuario. Características como la indexación semántica global (Find All References), vectorización IA (RAG) o la recolección proactiva de basura, requieren hilos de ejecución perpetuos. 

El *Background Services Orchestrator* es el subsistema responsable de hospedar, supervisar, aislar y estrangular (throttle) a estos demonios asíncronos. A diferencia del *Task Execution Engine* (SPEC-404) que procesa ráfagas finitas de compilación, el Orchestrator gestiona actores de larga duración (*Long-lived Actors*) protegiendo férreamente los recursos físicos del Host mediante políticas estrictas de suspensión no destructiva.

## 2. Definitions
* **Daemon (Background Service):** Un actor lógico de larga duración que opera en segundo plano para computar metadatos colaterales al proceso de compilación.
* **Orchestrator:** El componente del Host que actúa como supervisor jerárquico de todos los Daemons (Supervisor Tree).
* **Suspension:** El acto formal de pausar temporalmente un Daemon para liberar ciclos de CPU/RAM, sin destruir su estado en caché (Yield).
* **Termination:** La destrucción final y cancelación irreversible del contexto de un Daemon.
* **Daemon Context (Sandbox):** El contenedor de ejecución inyectado al Daemon, que provee acceso restringido a utilidades del Host (Logger, EventBus) y señales de ciclo de vida.

## 3. Scope & Non-Goals

* **In Scope:** Ciclo de vida de Daemons (Arranque, Pausa, Reanudación, Muerte), sandboxing mediante interfaces, taxonomía de servicios (Reactivos vs Periódicos) y aislamiento de fallos (Supervisor Tree).
* **Out of Scope:** Ejecución de AST, resolución de dependencias síncrona, orquestación de workers de compilación de corta duración (eso es SPEC-404).

## 4. Design Principles
1. **Never Crash the Host:** Un *Panic* dentro de un Background Service jamás debe derribar al Host. Debe ser atrapado y aislado por el Orchestrator.
2. **Yield over Kill (Non-Destructive Throttling):** Cuando el Host enfrenta escasez de recursos, los Daemons no son aniquilados brutalmente. Se les pide cortésmente que se suspendan conservando su estado útil.
3. **Sandbox by Default:** Los Daemons no reciben punteros al `Host` o al `ResourceManager`. Reciben un `DaemonContext` que actúa como Proxy seguro.
4. **Complete Silence (Zero-UI):** Un Daemon jamás solicita Input al usuario ni notifica errores a la UI. Solo emite Eventos (SPEC-405) o métricas.

## 5. Host Integration & Formal Resource Graph

El *Daemon Orchestrator* vive como un componente top-level dentro del Host, operando en paralelo al *Task Execution Engine*.

```text
Host
 │
 ├── Scheduler (SPEC-403)
 ├── Host Event Bus (SPEC-405)
 ├── Resource Manager (SPEC-413)
 │
 ├── Background Services Orchestrator (SPEC-406)
 │    │
 │    ├── (owns) DaemonRegistry
 │    │
 │    ├── (supervises) AI Indexer Daemon
 │    ├── (supervises) Global Reference Daemon
 │    └── (supervises) GC Invoker Daemon
```

**Ownership Invariants:**
* **OWN-001:** El Host *owns* al Orchestrator. The Host exclusively owns the Orchestrator lifecycle. The Orchestrator is not self-restarting. Failure of the Orchestrator is treated as Host failure.
* **OWN-002:** El Orchestrator *owns* el hilo físico (goroutine) en el que corre el Daemon.
* **OWN-003:** El Daemon *owns* exclusivamente la memoria generada en sus índices internos.

## 6. Identity Model
* `DaemonID` (Const String, ej. `fenix.daemon.ai_indexer`).
* `RunID` (UUIDv7 asignado en cada ejecución, incluyendo el arranque inicial y cada reinicio tras un crash).

## 7. Formal System Properties (Safety & Liveness)

### 7.1. Safety Properties (Invariants)
* **SP-001 (Fault Containment):** Ningún error, *panic* o bloqueo mortal (`deadlock`) dentro de un Daemon afectará el hilo principal del Host.
* **SP-002 (Sandbox Purity):** Un Daemon no puede suscribirse ni publicar Eventos fuera de los permisos otorgados en su `DaemonContext`.
* **SP-003 (One Active Instance):** At any time, there exists at most one `Running` instance of a given `DaemonID`. El sistema impide terminantemente colisiones de orquestación donde dos hilos asuman el mismo demonio.

### 7.2. Liveness Properties (Progress)
* **LP-001 (Self-Healing / Respawn):** Si un Daemon crashea, el Orchestrator garantiza un reinicio eventual tras aplicar un *Exponential Backoff*.
* **LP-002 (Eventual Resumption):** Si el Resource Manager ordena suspender a un Daemon por presión de memoria, el Daemon será reanudado *eventualmente* cuando la alerta ceda.

## 8. Taxonomy of Daemons

Matemáticamente, los Daemons se dividen en dos familias ortogonales de activación:

### 8.1. Reactive Daemons (Event-Driven)
Despiertan exclusivamente cuando el *Host Event Bus* publica eventos de su interés.
* **Ejemplo:** El *Global Reference Indexer* despierta al recibir `workspace.snapshot.created`.
* **Mecánica:** Consumen recursos únicamente respondiendo a deltas. Permanecen en `select{}` pasivo el 99% del tiempo.

### 8.2. Periodic Daemons (Time-Driven)
Despiertan mediante un `time.Ticker` gestionado internamente.
* **Ejemplo:** El *Garbage Collector Invoker* o el *Telemetry Flusher*.
* **Mecánica:** Ejecutan ráfagas programadas. Sensibles al estrangulamiento de CPU.

## 9. System Component Model

```text
Daemon Orchestrator
 ├── API Server              (Start, Stop, Suspend)
 ├── Supervisor Tree         (Manejo de pánicos y reinicios)
 ├── Context Provider        (Inyección de dependencias seguras)
 └── Throttling Controller   (Acoplamiento con Resource Manager)
```

## 10. Domain Model (Aggregates & Lifetimes)

### 10.1. Orchestrator Aggregate
* `Registry` (Mapa de Daemons registrados)
* `Supervisors` (Mapa de hilos de control)
* `State` (`Created`, `Running`, `Throttled`, `Stopping`)

**Formal State Invariants (Orchestrator):**
* `Created`: El único estado donde `Register()` es permitido. Dynamic registration is prohibited.
* `Running`: `Supervisors != nil`. **Once `StartAll()` has been invoked, the Registry becomes immutable for the remainder of the Host lifetime.**
* `Throttled`: Estado administrativo global. Ocurre cuando al menos 1 nivel de prioridad ha recibido la señal de suspensión. (Internamente, cada daemon mantiene su propio estado lógico `Suspended`).
* `Stopping`: Nuevos Daemons rechazados. `ShutdownTimeout` reloj iniciado.

### 10.2. Daemon Registration Aggregate
* `DaemonID`
* `Priority` (High, Normal, Low)
* `Factory` (Función constructora del Daemon)
* `CrashCount` (Decays after `StableWindow`, e.g. 15 minutes, to prevent permanent death from ancient transient errors).
* `State` (`Registered`, `Starting`, `Running`, `Suspended`, `Crashed`, `Stopping`, `Terminated`)

**Formal State Invariants (Daemon Registration):**
* `Running`: `RunID != nil`, goroutine física en ejecución, CPU > 0.
* `Suspended`: `RunID != nil`, goroutine física bloqueada (Yield), CPU == 0.

### 10.3. Daemon Context (Sandbox Aggregate)
El entorno inyectado hacia la lógica del demonio.
* `Context` (Go context para terminación total).
* `SuspensionSignal` (`<-chan struct{}` para avisar que debe entrar a Yield).
* `ResumeSignal` (`<-chan struct{}` para despertar).
* `Logger` (Etiquetado automáticamente).
* `EventPublisher` (Wrapper del Event Bus).
* `HostQueryAPI` (Interfaz *Read-Only* al estado del sistema).
* `TimeProvider` (`Clock()` para garantizar un Host determinista en pruebas y mocks de tiempo temporal).

## 11. Lifecycle, Throttling & State Machines

El ciclo de vida del Daemon introduce el concepto vital de *Suspensión no destructiva*.

### 11.1. Daemon State Machine
`Registered` ➔ `Starting` ➔ `Running` ⭤ `Suspended` ➔ `Stopping` ➔ `Terminated`
(con transiciones asíncronas a `Crashed` derivadas de un panic en `Running` o `Suspended`).

### 11.2. Throttling Protocol & Yield Contract
Cuando el *Resource Manager* (SPEC-413) decreta estado de escasez (Ej. *RAM > 90%*), invoca la suspensión por prioridades.
* **Priority Semantics:** `Suspend(Threshold)` pauses the `Threshold` priority **AND all lower priorities**. For example, `Suspend(PriorityNormal)` suspends Normal + Low.
* **Latch Model:** Suspension is idempotent. Multiple `Suspend()` calls while already suspended have no additional effect (no stacks or counters).

**El Yield Contract Formal:**
1. Upon receiving `SuspensionSignal`, a Daemon **MUST** stop initiating new computational work before 50 ms.
2. A Daemon **MAY** finish an atomic internal operation already in progress.
3. A Daemon **MUST NOT** begin a new work item until `ResumeSignal`.
4. Suspension pauses execution *only*. Event delivery semantics remain governed exclusively by SPEC-405. 
5. **Reactive Daemons MAY coalesce pending Events while suspended.** The coalescing policy is daemon-specific. The EventBus delivery guarantees remain unchanged.

**Código Normativo del Throttling Loop:**
```go
select {
case <-dCtx.Suspended():
    // Yield: El demonio acata el contrato deteniéndose aquí.
    // Esperamos reanudación o cancelación final.
    select {
    case <-dCtx.Resumed():
        // Despierta y vuelve al trabajo
    case <-dCtx.Done():
        return ErrTerminated // Shutdown limpio durante suspensión
    }
case <-dCtx.Done():
    return ErrTerminated
default:
    // ... continuar trabajo (ráfagas < 50ms)
}
```

### 11.3. Startup & Shutdown Semantics

**Startup Ordering:**
Daemons are started sequentially in registration order. Each daemon is considered `Running` immediately after its `Run()` goroutine is successfully spawned. The Orchestrator **MUST NOT** wait for the daemon to finish its initialization logic.

**Shutdown (Graceful vs Force):**
1. El Orchestrator cierra el canal `Done()` del `DaemonContext`.
2. El Orchestrator inicia el `ShutdownTimeout` (definido formalmente en la configuración del Host, ej. `const DefaultShutdownTimeout = 5 * time.Second`).
3. **Graceful:** Si el Daemon retorna de su `Run()` antes del timeout, muere limpio (`Terminated`).
4. **Force:** Si expira el `ShutdownTimeout`, el Orchestrator abandona la goroutine (la desregistra) y permite al Host salir (`os.Exit`).

## 12. Fault Tolerance & Supervisor Tree

Implementa el patrón *Erlang OTP Supervisor Tree* simplificado (Estrategia `One-For-One`).

* **Panic Trap:** El Orchestrator envuelve la invocación de `Daemon.Run()` en un `defer recover()`.
* **Backoff Strategy:** Si el `Daemon` crashea, el Orchestrator transita a estado `Crashed`, incrementa el `CrashCount` e intenta reiniciarlo usando *Exponential Backoff* (1, 2, 4, 8, 16, 30, 30... **capped at 30 seconds**). Si un Daemon de prioridad `PriorityHigh` entra en fallo permanente, el Orchestrator **DEBE** emitir un evento crítico de fallo sistémico. Un fallo permanente en PriorityHigh desencadena la degradación gracefully del Host o un reinicio del proceso principal, dependiendo de la política de configuración.
* **State Inheritance:** If a suspended daemon crashes, the replacement instance inherits the `Suspended` state until `Resume(threshold)`.
* **Stable Decay:** `CrashCount decays after StableWindow`. Si el daemon sobrevive 15 minutos sin errores, el contador vuelve a cero.

## 13. System Communication: The Read-Only Rule

Dada la política estricta de aislamiento, los Daemons interactúan con el resto del Host bajo una regla asimétrica de mutabilidad:

* **Lecturas (Read-Only):** Los Daemons pueden consultar APIs síncronas del Host a través del `dCtx.HostQueryAPI`.
* **Escrituras (Mutations via Events Only):** Los Daemons tienen **PROHIBIDO** mutar el estado del Host directamente. Debe emitir un Evento.
* **Reentrancy:** Daemons *may* publish Events at any point during `Run()`.

## 14. Telemetry & Logger Injection

El Orchestrator libera al autor del demonio de escribir código repetitivo para la observabilidad distribuida.
Al inyectar el `dCtx.Log()`, el Orchestrator preconfigura el Logger inyectando silenciosamente los campos estructurados obligatorios:
* `daemon_id="fenix.daemon.ai_indexer"`
* `run_id="01H...XYZ"` (UUIDv7 actual de ejecución).

## 15. Formal Happens-Before Semantics

Para alinear el Orchestrator con el modelo de concurrencia matemática, definimos las siguientes fronteras de sincronización inquebrantables:
1. `Orchestrator sends SuspensionSignal` **happens-before** `Daemon enters Yield`.
2. `Orchestrator sends ResumeSignal` **happens-before** `Daemon resumes computational work`.
3. `Orchestrator sends Done()` **happens-before** `Daemon.Run() returns`.
4. `recover() (Crash)` **happens-before** `new Run() (Restart)`.

## 16. Official Core Background Services

Fénix incluye nativamente estos Daemons (y delega el resto a plugins):

### 16.1. Global Reference Indexer (Reactive, Priority: Normal)
* **Trigger:** `workspace.snapshot.created`.
* **Duty:** Invertir la referencia de símbolos para resolver *Find All References* en O(1).
* **Yield:** Pausa la indexación frente a picos de CPU del compilador central.

### 16.2. AI Project Embedder (Reactive, Priority: Low)
* **Trigger:** `workspace.snapshot.created` (Debounced).
* **Duty:** Extraer la semántica del AST y empujar embeddings a la base vectorial local.
* **Yield:** Altamente sensible. Se suspende a la menor provocación.

### 16.3. Dependency Fetcher (Reactive, Priority: Normal)
* **Trigger:** Eventos custom de resolución de módulos.
* **Duty:** Descargar librerías externas faltantes (I/O intensivo).

### 16.4. GC Invoker (Periodic, Priority: High)
* **Trigger:** Intervalo dinámico basado en la presión de memoria.
* **Duty:** Vigilar recolecciones de memoria manuales en Fénix para evitar picos de latencia en la compilación.

## 17. Operation Contracts (Design by Contract)

### 17.1. Orchestrator.Register(Daemon)
* **Requires:** `OrchestratorState == Created`. Prohibido el registro dinámico. Priority **MUST** be one of the predefined constants. Implementations **MUST** reject unknown priorities.
* **Ensures:** Retorna `error` (fail-fast) si el ID está duplicado, el estado es incorrecto o la prioridad es inválida. De lo contrario, Daemon agregado al Registry en estado `Registered`.

### 17.2. Orchestrator.Suspend(ThresholdPriority)
* **Requires:** `OrchestratorState == Running`.
* **Ensures:** Todos los daemons con prioridad igual **o menor** al threshold reciben la `SuspensionSignal`. Estado transitado a `Suspended`. Suspension is idempotent.

### 17.3. Orchestrator.Resume(ThresholdPriority)
* **Requires:** `OrchestratorState == Running` o `Throttled`.
* **Ensures:** Reanuda daemons con prioridad **igual o mayor** al threshold que estén suspendidos. Permite reanudación escalonada (ej. despertar Normales, pero mantener Low suspendidos). `Running` daemons remain unaffected. `Terminated` daemons are ignored.

### 17.4. Daemon.Run(DaemonContext)
* **Requires:** Entorno de ejecución puro y aislado por el Orchestrator.
* **Ensures:** El demonio respeta religiosamente el *Yield Contract*, y la cancelación de `Done()` liberando sus locks internos en el plazo del `ShutdownTimeout`.

## 18. Concurrency, Reentrancy & Memory Ownership

* **Aislamiento Físico:** Cada Daemon corre en exactamente una goroutine principal provista por el Orchestrator.
* **Memory Purity:** Los Daemons jamás comparten memoria entre sí. Toda comunicación Inter-Daemon ocurre estricta y obligatoriamente por el **Event Bus (SPEC-405)**.
* **Sandbox Mutability:** El Daemon es dueño absoluto de su memoria interna. Ningún otro componente puede leer sus índices sin pasar por las APIs de consulta correspondientes.

## 19. Algorithmic Complexity Bounds

| Operation | Target Complexity | Notas |
|---|---|---|
| `Orchestrator.Suspend`| **O(D)** | `D` = Número total de Daemons registrados. Notificación push a canales. |
| `Daemon.Yield Time`| **< 50ms** | Contractual. Un demonio debe responder a la señal de suspensión en menos de 50ms. |

## 20. Normative Metrics

El Orchestrator emite obligatoriamente:
* `daemon_crashes_total{daemon_id="X"}` (Count)
* `daemon_suspension_duration_seconds{daemon_id="X"}` (Histogram)
* `daemon_active_goroutines_total` (Gauge)
* `daemon_uptime_seconds{daemon_id="X"}` (Gauge)

## 21. Interface Contract

```go
package background

import "context"

type Priority int
const (
    PriorityHigh Priority = iota
    PriorityNormal
    PriorityLow
)

// DaemonContext es el Sandbox inyectado
type DaemonContext interface {
    // Señal destructiva (Shutdown final)
    Done() <-chan struct{}
    
    // Señales no destructivas (Throttling Yield Contract)
    Suspended() <-chan struct{}
    Resumed() <-chan struct{}
    
    // Utilidades seguras y pre-etiquetadas
    Log() Logger 
    EventPublisher() EventPublisher // Wrapper seguro hacia SPEC-405
    HostQuery() HostQueryAPI        // Read-only queries hacia el AST
    TimeProvider() Clock            // Determinismo en pruebas
}

type Daemon interface {
    ID() string
    Priority() Priority
    
    // El punto de entrada principal. 
    // Obligado contractualmente a implementar select{} con dCtx.Suspended()
    Run(dCtx DaemonContext) error
}

type Orchestrator interface {
    // Falla rápido (error) si el estado no es Created o si hay colisiones
    // Implementations MUST reject unknown custom priorities.
    Register(d Daemon) error
    
    StartAll(ctx context.Context) error
    StopAll() error // Dispara el ShutdownTimeout
    
    // API para el Resource Manager (SPEC-413)
    // threshold = PriorityNormal suspende Normal y Low. (Idempotent)
    Suspend(threshold Priority) 
    
    // Reanuda exclusivamente a los suspendidos con prioridad >= threshold.
    // Permite reanudación escalonada.
    Resume(threshold Priority) 
}
```

## 22. Cross References
**Normative References:**
* **[SPEC-413](413-resource-manager.md):** Resource Manager (Actor que invoca Suspend/Resume).
* **[SPEC-405](405-event-bus.md):** Host Event Bus (Fuente de Eventos Reactivos).
* **[SPEC-404](404-task-execution-engine.md):** Execution Engine (Cuyos picos de CPU causan el Throttling de los Daemons).
