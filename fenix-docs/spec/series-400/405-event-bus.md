# SPEC-405: Host Event Bus

## 1. Executive Summary
El *Host Event Bus* es el tejido conectivo asíncrono intra-proceso del Compiler Host de Fénix. Su única responsabilidad es distribuir notificaciones inmutables de cambio de estado (Eventos) entre los componentes core (Serie 300/400) y los consumidores de efectos secundarios (LSP, Telemetría, Caché, Indexadores IA) garantizando un desacoplamiento absoluto. Es un sistema Pub/Sub local, diseñado para latencia ultrabaja, *zero-copy routing*, y tolerancia estricta a fallos de los suscriptores.

## 2. Definitions
* **Host:** El runtime central de compilación. El Host es el dueño (*owner*) del Event Bus y provee los límites físicos del sistema.
* **Event Envelope:** La estructura portadora (Header) estandarizada que envuelve un cambio de estado.
* **Event Payload:** El struct inmutable de datos tipados (Generics) asociados al evento.
* **Topic:** Un identificador jerárquico (ej. `workspace.snapshot.created`) usado para el enrutamiento.
* **Publisher:** Cualquier componente del sistema que emite un Evento (ej. Scheduler, ExecutionEngine).
* **Subscriber:** Un componente periférico que declara interés en uno o más Topics.
* **Subscription:** El contrato vinculante entre el Bus y un Subscriber, materializado como un canal de Go (`<-chan Event[T]`).
* **Shedding (Descarte):** La acción de eliminar un Evento de la cola de un suscriptor lento para proteger la latencia general del sistema.

## 3. Scope
* **In Scope:** Enrutamiento in-memory rápido, filtrado por Topics jerárquicos (Prefix/Exact), políticas de *Backpressure* y *Shedding*, aislamiento estricto de bloqueos de suscriptores, y propagación causal de eventos.
* **Out of Scope:** Colas distribuidas (Kafka, RabbitMQ), persistencia a disco, RPC, comunicación Inter-Process (IPC) o Request/Response síncrono.

## 4. Design Principles
1. **Asynchronous by Default:** Publicar un evento jamás debe bloquear al hilo productor.
2. **Absolute Immutability:** El payload de un evento es estructuralmente inmutable.
3. **Symmetric Ignorance:** Los productores desconocen a los suscriptores y viceversa.
4. **Push-to-Channel, Not Callbacks:** El Bus entrega eventos empujándolos a un canal. Obliga a los suscriptores a usar bucles de control (`for/select` loop) propios (Actor Model), eliminando la necesidad de que el Bus maneje concurrencia delegada.

## 5. Host Integration & Formal Resource Graph

El *Host Event Bus* vive exclusivamente dentro del límite del Host y actúa como su sistema nervioso central.

```text
Host
 │
 ├── Scheduler
 ├── Host Event Bus
 │    │
 │    ├── (owns) Routing Table
 │    └── (owns) Dispatcher Engine
 │
 ├── Execution Engine
 ├── Resource Manager
 ├── Worker Pools
 └── Telemetry
```

El flujo de pertenencia física de la memoria del Bus es implacable:

```text
Publisher
    │  (emits)
    ▼
EventEnvelope (Immutable)
    │  (routed by)
    ▼
EventBus
    │  (dispatches to)
    ▼
Subscription (Channel Buffer)
    │  (consumed by)
    ▼
Subscriber
    │  (owns)
    ▼
Side Effects (I/O, UI, Metrics)
```

**Ownership Invariants:**
* **OWN-001:** El Host *owns* al EventBus.
* **OWN-002:** El Suscriptor *owns* el hilo de ejecución (goroutine) que procesa el evento. El EventBus jamás presta hilos.
* **OWN-003:** Una vez publicado, **NADIE** *owns* la mutabilidad del Payload. Pasa a ser solo-lectura absoluta.

## 6. Identity Model
* `EventID` (UUIDv7: Garantiza ordenamiento temporal local y unicidad sin contadores globales).
* `ParentEventID` (UUIDv7: Para reconstrucción de cadenas causales).
* `TopicID` (String jerárquico canónico).
* `SubscriberID` (UUID).
* `PublisherID` (Enum / Const String, ej. `Host`, `Scheduler`, `ExecutionEngine`).
* `CorrelationID` (TraceID para Telemetría).

## 7. Formal System Properties (Safety & Liveness)

### 7.1. Safety Properties (Invariants)
* **SP-001 (Immutability):** Una vez que `Publish(ctx, Event)` es invocado, el contenido del `Event` no puede ser alterado en memoria.
* **SP-002 (Isolation):** Un `panic` dentro de un Subscriber jamás puede propagarse ni colapsar al `EventBus`.
* **SP-003 (Subscription Exclusivity):** Every Subscription belongs to exactly one EventBus and owns exactly one Channel.
* **SP-004 (Non-Blocking Publish):** Un `Publisher` jamás debe bloquear su ejecución de forma indefinida porque un `Subscriber` esté colapsado.
* **SP-005 (Concurrent Publish Safety):** `Publish()` is completely safe for concurrent invocation by multiple goroutines without explicit locking.

### 7.2. Liveness Properties (Progress)
* **LP-001 (Eventual Delivery or Drop):** Todo evento publicado es *eventualmente* entregado al canal del suscriptor, o *explícitamente descartado* (Shedding).
* **LP-002 (Slow Subscriber Eviction):** Un suscriptor que crónicamente bloquea sus canales *eventualmente* es desconectado forzosamente (`Evicted`) para proteger la integridad del Bus.
* **LP-003 (Eventual Drain):** Al iniciar un Shutdown gracefully, el Bus *eventualmente* vacía todos los eventos encolados o aborta mediante `ShutdownTimeout`.

## 8. ADR Summary
| ID | Decisión | Justificación |
|---|---|---|
| **EVT-ADR-001** | Channel-based Subscribers | Exigir que el Suscriptor provea un canal en lugar de un *Callback*. Libera al Bus de manejar concurrencia. |
| **EVT-ADR-002** | Dual Delivery Guarantees | Separar el trato a eventos críticos (Backpressure) vs telemetría (Shedding) previene que métricas colapsen al Host. |
| **EVT-ADR-003** | UUIDv7 over Monotonic | Se prefiere UUIDv7 sobre enteros monótonos para evitar contención global y facilitar el Merge de trazas. |
| **EVT-ADR-004** | Generic Wrapper Functions | Utilizar Generics a nivel de paquete en Go en lugar de métodos genéricos en interfaces (limitación técnica) para asegurar el Type Safety estricto. |

## 9. Non Goals
El EventBus **NO**:
* Garantiza entrega de red (es puramente intra-proceso).
* Provee consultas históricas (no es una base de datos).
* Soporta **Sticky Events / Replay Last**: Si un suscriptor nace tarde (ej. un LSP client), no recibe el estado anterior. Debe consultar la API síncrona pertinente (ej. `SnapshotManager`).

## 10. System Component Model
```text
Host Event Bus
 ├── Publisher API           (Package-level Generic Wrappers)
 ├── Topic Router            (Radix Tree para matching exacto/prefijo)
 ├── Dispatcher Engine       (Thread asíncrono para push a canales)
 ├── Subscription Registry   (Manejo de altas/bajas)
 └── QoS Monitor             (Vigila latencia y aplica Eviction)
```

## 11. Domain Model (Aggregates & Lifetimes)

### 11.1. EventBus Aggregate
El orquestador central del componente.
* `Registry` (SubscriptionRegistry)
* `RoutingTable` (TopicRouter)
* `Dispatcher` (DispatcherEngine)
* `QoSMonitor` (Monitorea las violaciones de timeout)
* `Metrics`
* `Configuration` 
* `State` (`Created`, `Starting`, `Running`, `Draining`, `Stopped`)

**Formal State Invariants:**
* `Created`: `RoutingTable == nil`, `Dispatcher == nil`
* `Running`: `RoutingTable != nil`, `Dispatcher != nil`, `Queue != nil`
* `Stopped`: `No Channels`, `No Dispatcher`, `No Queue` (Recursos liberados).

### 11.2. EventBus Configuration
* `DispatcherQueueCapacity`: Límite máximo de eventos pendientes globales.
* `ShutdownTimeout`: Tiempo límite para el drenaje de eventos.
* `CriticalPublishTimeout`: Tolerancia máxima de Backpressure antes del *Eviction*.
* `DefaultSubscriptionCapacity`: Tamaño buffer de los canales.
* `EvictionThreshold`: Regla formal (`N` violaciones de timeout consecutivas en una `RollingWindow` T).
* `MetricsEnabled`: Booleano para inyección de Prometheus.

### 11.3. Dispatcher Engine Aggregate
El worker dedicado a rutar y empujar eventos.
* `Thread/Loop`: Single Background Goroutine (aislado para evitar locks en el enrutamiento interno).
* `Queue`: **Bounded MPSC Queue** (Multi-Producer, Single-Consumer). Recibe eventos desde N `Publishers` de forma asíncrona, y los consume la única `Single Goroutine` del Dispatcher.
  * *Nota de Implementación:* La SPEC exige comportamiento MPSC Bounded. La implementación puede utilizar un Lock-Free Ring Buffer, una CAS-based Queue, o un Buffered Channel optimizado.
* `Capacity` / `Head` / `Tail` / `Size`
* `DropPolicy`
* `OverflowCounter`
* `State` (`Idle`, `Processing`, `Draining`, `Terminated`)

### 11.4. Subscription Registry Aggregate
El gestor concurrente de suscripciones.
* `Subscriptions` (Map interno)
* `Index` (Búsqueda rápida por SubscriberID)
* `RWLock` (Protege mutaciones del map durante altas/bajas)
* `Version` (**Atomic Uint64**). Se incrementa atómicamente en cada adición o borrado para invalidar cachés en el Dispatcher. No requiere locks para su lectura.

### 11.5. Topic Router Aggregate
El cerebro topológico, basado en un Radix Tree en memoria.
* `Root` (Nodo raíz del árbol)
* `Nodes`
* `PrefixIndex` (Optimización para wildcard `*`)
* `SubscriberLists` (Caché pre-calculado por Topic)
* **Algoritmos formales:**
  * **Inserción:** `O(K)` donde `K` es la profundidad del topic. Incrementa el `Version` del Registry.
  * **Borrado:** `O(K)` o *Lazy Deletion* marcando el nodo como inactivo. Incrementa el `Version`.
  * **Cache Invalidation:** Cuando el Dispatcher detecta que `Registry.Version` cambió, destruye su caché local de `SubscriberLists` (Read-Copy-Update pattern) para evitar bloqueos durante el enrutamiento.

### 11.6. Event Envelope & Payload Aggregates
La estructura principal separa rigurosamente el metadato del dato.

**Header (Envelope):**
* `EventID` (UUIDv7)
* `ParentEventID` (Opcional, causalidad)
* `CorrelationID`
* `PublisherID`
* `Topic`
* `SchemaVersion`
* `Timestamp`
* `Priority` (Critical, Normal, Low)

**Payload:**
* `Data` (`T EventPayload` interface sellada).

### 11.7. PublishResult Aggregate
El retorno de la llamada asíncrona.
* `Accepted`: El evento entró a la Dispatcher Queue (MPSC).
* `Rejected`: El Bus rechazó (ej. `State != Running`).
* `Dropped`: La Dispatcher Queue estaba llena (QoS 0 Shedding).
* `Shutdown`: El bus se está apagando.
* `TypeMismatch`: (o `InvalidContract`) El Payload no corresponde al Topic registrado (Error del desarrollador).

### 11.8. Subscription Aggregate
* `SubscriptionID`
* `SubscriberID`
* `TopicFilter`
* `Channel` (`chan Event[T]`)
* `BufferCapacity`
* `State` (`Active`, `Suspended`, `Evicted`, `Released`)

### 11.9. Formal Resource Lifetime
| Resource | Created By | Destroyed By |
|---|---|---|
| `EventBus` | `Host` | `Host` (Post Shutdown) |
| `EventEnvelope`| `Publisher` | `Garbage Collector` |
| `Subscription` | `Registry` | `Registry` (On Unsubscribe or Eviction) |

## 12. Formal Event Flow

Para garantizar una implementación inequívoca, el flujo normativo de un evento desde el Publisher hasta el Subscriber es el siguiente:

1. **Publisher:** Invoca `Publish(Event)`.
2. **Validate Contract:** El EventBus verifica que el tipo del Payload cumpla el contrato del Topic registrado. Si falla ➔ `PublishResult.TypeMismatch`.
3. **Queue Admission:** Intenta insertar en la Bounded MPSC Queue.
   * Si está llena (Prioridad Baja) ➔ `PublishResult.Dropped`.
   * Si el Host se apaga ➔ `PublishResult.Shutdown`.
   * Si entra exitosamente ➔ `PublishResult.Accepted`. (El Publisher retorna de la función aquí).
4. **Dispatcher Loop:** La Single Goroutine del Dispatcher extrae el evento de la Queue.
5. **Topic Match:** El Dispatcher consulta el `Topic Router` (Radix Tree) para obtener la lista de Suscriptores.
6. **Subscription Iteration:** El Dispatcher itera secuencialmente sobre todos los Suscriptores que hicieron Match.
7. **Channel Push:** El Dispatcher intenta empujar el Evento al canal del Suscriptor mediante un `select{}`.
8. **QoS Decision:** 
   * Si el canal empuja inmediatamente ➔ Éxito (Delivery).
   * Si el canal bloquea más de `CriticalPublishTimeout` ➔ Se aborta el Push, el Evento se descarta (para ese suscriptor específico) y el QoS Monitor interviene.
   * Si el QoS Monitor detecta violaciones continuas ➔ Transición a `Evicted`.
9. **Subscriber:** La goroutine independiente del suscriptor extrae el evento del canal en su propio tiempo.

## 13. Lifecycle, Shutdown & State Machines

### 13.1. Bus Lifecycle
`Created` ➔ `Starting` ➔ `Running` ➔ `Draining` ➔ `Stopping` ➔ `Stopped`

### 13.2. Subscription Lifecycle
`Created` ➔ `Active` ➔ `Suspended` (Temporal Backpressure) ➔ `Active` ➔ `Evicted` ➔ `Released`

### 13.3. Shutdown Semantics (Graceful vs Force)
Un apagado `Graceful` requiere pasos estrictos para prevenir *Race Conditions*.
**Secuencia Formal:**
1. `Reject Publish`: El Bus responde `PublishResult.Shutdown` a nuevas solicitudes.
2. `Stop Subscriptions`: Se bloquea el Registry. No entran nuevos Suscriptores.
3. `Drain Queue`: El Dispatcher vacía la cola a los canales existentes.
4. **TIMEOUT CHECK:** Si expira el `ShutdownTimeout`, procede a cierre forzado.
5. `Close Dispatcher`: Se detiene el hilo ruteador interno.
6. `Close Channels`: Cierre explícito de todos los `chan Event[T]` para liberar a los Actors de los suscriptores.
7. `Stopped`.

## 14. Delivery Guarantees, QoS & Eviction

Para manejar situaciones límite, el EventBus define capacidades y monitoreos estrictos.

### 14.1. QoS Monitor & Eviction Threshold
* **¿Quién mide la Tensión T?** El `QoS Monitor` es un componente pasivo invocado por el `Dispatcher` durante un *Channel Push*. 
* Si la inserción en el canal de un Suscriptor bloquea más de `CriticalPublishTimeout`, el Dispatcher aborta el envío y notifica al `QoS Monitor`.
* El `QoS Monitor` guarda un *Timestamp* del bloqueo. Si se registran `N` bloqueos consecutivos dentro de una ventana de tiempo `T` (Rolling Window), el Monitor envía un comando asíncrono al `Registry` para marcar la suscripción como `Evicted` y cerrar su canal.

### 14.2. QoS 0: At-Most-Once (Telemetry/Metrics)
Eventos marcados con `Priority = Low`.
* **Mecánica:** Si el Dispatcher Queue o el Subscription Buffer se llenan, retorna `PublishResult.Dropped` inmediatamente (`Shedding`).

### 14.3. QoS 1: Bounded Backpressure (State/Critical)
Eventos marcados con `Priority = Critical`.
* **Mecánica:** Bloqueo asíncrono temporal en el Dispatcher. Tras superar el límite, interviene el QoS Monitor.

## 15. Formal Failure Domains

Para aislar fallos, el diseño impone un árbol estricto de confinamiento:

`Subscriber` ⊂ `Subscription` ⊂ `Dispatcher` ⊂ `EventBus` ⊂ `Host`

* Un *Panic* en el **Subscriber** destruye al Subscriber. La `Subscription` queda huérfana pero el Dispatcher sobrevive.
* Un *Panic* interno severo destruye al **EventBus**. El Host sobrevive e invoca recuperación.

**DP-001: Dispatcher Panic Recovery (Self-Healing)**
The EventBus MUST supervise the Dispatcher (Supervisor Loop).
If the Dispatcher terminates unexpectedly due to a panic (e.g. routing anomaly), the EventBus MUST:
1. Trap the panic via `recover()`.
2. Transition the failed Dispatcher to `Terminated`.
3. Recreate a new Dispatcher Goroutine.
4. Resume processing from the remaining MPSC Queue.
*This completes the Self-Healing property for the core routing component.*

## 16. Topic Taxonomy & Filters

### 16.1. Formal Topic Grammar
Todo topic en Fénix **DEBE** obedecer esta gramática matemática estricta:
`<domain>.<aggregate>.<verb>`

### 16.2. Subscription Filters & Type Safety Rules
Existe una tensión inherente entre filtros jerárquicos (`engine.*`) y el tipado fuerte (`Event[T]`). Fénix resuelve esto con dos reglas normativas:

* **Regla 1: Tipado Estricto para Exact Match.**
  Si un filtro es exacto (`engine.attempt.panicked`), la suscripción **DEBE** usar el tipo concreto exacto: `Subscribe[EnginePanickedPayload]("engine.attempt.panicked")`.
* **Regla 2: Tipado Polimórfico para Prefix Match.**
  Si un filtro es jerárquico (`engine.*`), múltiples tipos de payloads fluirán por el canal. La suscripción **DEBE** usar una interfaz polimórfica (ej. `Subscribe[EngineEventPayload]("engine.*")`) o hacer fallback a la base de `EventPayload`. 

## 17. Event Contract Registry & Runtime Type Validation

El sistema exige un vínculo inquebrantable entre un Topic Canónico y su Payload asociado garantizado por el sistema de tipos.

* **Startup Rule:** En la inicialización del `Host`, los subsistemas invocan la función genérica `RegisterType[T](Topic, SchemaVersion)`.

| Canonical Topic | Permitted Payload Type | SchemaVersion |
|---|---|---|
| `workspace.snapshot.created` | `SnapshotCreatedPayload` | v1 |
| `engine.attempt.completed` | `AttemptCompletedPayload` | v2 |

**Runtime Type Validation (Type Erasure boundary):**
The Event Contract Registry MUST verify that the runtime payload type matches the registered Topic contract before enqueueing. 
*The validation mechanism is implementation-defined* (e.g., using `reflect`, `unsafe.Type`, generated IDs, or type hashes). Si la validación falla, el Bus aborta el encolamiento inmediatamente devolviendo `PublishResult.TypeMismatch`, preservando la inmutabilidad de tipos del sistema.

## 18. Concurrency, Reentrancy & Formal Event Ordering

### 18.1. Event Ordering Guarantees
* **Per-Publisher / Per-Subscription:** El orden **ESTÁ GARANTIZADO**. Si el Publisher `A` emite `Ev1` y luego `Ev2`, el Suscriptor los leerá estrictamente en el orden `Ev1 ➔ Ev2`.
* **Global Ordering:** El orden global entre **DISTINTOS Publishers** está **INTENCIONALMENTE INDEFINIDO**.

### 18.2. Reentrancy Policy
* Subscribers may safely publish new Events while processing Events.
* The EventBus guarantees no reentrant dispatcher execution (el hilo interno jamás se auto-llama recursivamente).

### 18.3. Happens-Before Semantics (Memory Model)
La cadena de visibilidad se basa puramente en sincronizaciones reales del modelo de memoria de Go:
1. Mutación de memoria (Producer State) **happens-before** llamada a `Publish()`.
2. `Publish()` **happens-before** el Atomic Enqueue (o Channel Send a la MPSC Queue).
3. Dispatcher MPSC Queue Read **happens-before** Successful Channel Send al Suscriptor.
4. Channel Send **happens-before** Subscriber Receive (`select{}`).
*Esta cadena inquebrantable garantiza visibilidad total del grafo de memoria modificado sin necesidad de Mutex adicionales.*

## 19. Formal Operation Contracts (Design by Contract)

### 19.1. Publish (Generic Wrapper)
* **Requires:** `Event.Topic` complies with grammar, `EventBus.State == Running`.
* **Ensures:** Evento encolado. Retorna `PublishResult.Accepted`.

### 19.2. Subscribe (Generic Wrapper)
* **Requires:** `BufferSize > 0`, tipo `T` is compatible with `TopicFilter`.
* **Ensures:** `Subscription.State == Active`, retorna `<-chan Event[T]`.

### 19.3. Unsubscribe(SubscriptionID)
* **Requires:** `Subscription.State == Active`.
* **Ensures:** `Subscription.Channel` Closed, Registry removes entry, `Subscription.State == Released`.

## 20. Consistency & Memory Ownership Model

**Serialization-Free Boundary:**
El Event Bus intra-proceso es matemáticamente *Serialization-free*. Los eventos **jamás** son serializados a JSON o Protobuf mientras viajan por el Host.

**Zero-Copy Routing Commitment:**
The Publisher MUST NOT mutate any object graph reachable from the Payload after a successful `Publish()`. If later mutation is required, the Publisher is strictly responsible for creating and publishing an immutable deep copy.

**Context Ownership:**
The EventBus **never owns nor cancels** contexts (`ctx`). The Publisher owns the lifecycle of the context passed a `Publish()`.

## 21. Algorithmic Complexity Bounds

| Operation | Target Complexity | Notas |
|---|---|---|
| `Enqueue (Publish)`| **O(1)** | Atomic push into Bounded MPSC Queue. |
| `Routing`| **O(K)** | `K` = Profundidad del topic. Uso de Caché de Listas mitigado por Version increment. |
| `Dispatch` | **O(M)** | `M` = Suscriptores que hicieron Match. El Dispatcher realiza un channel push **secuencial** hacia cada Suscriptor afectado. La concurrencia de procesamiento la dictan los Subscribers. |
| `Shutdown` | **O(S)** | `S` = Total Subscriptions a drenar/cerrar. |

## 22. Normative Metrics

El EventBus debe emitir obligatoriamente este conjunto de métricas para la telemetría del Host:
* `events_published_total` (Count)
* `events_dispatched_total` (Count)
* `events_dropped_total` (Count)
* `subscriber_evictions_total` (Count)
* `dispatcher_queue_depth` (Gauge)
* `dispatcher_latency_seconds` (Histogram)
* `subscription_backpressure_total` (Count)

## 23. Interface Contract

Para superar la limitación técnica de Go 1.18+ (que prohíbe métodos genéricos en interfaces o en structs instanciados genéricamente para toda la aplicación), el Event Bus utiliza **funciones envoltorio genéricas (Package-level Generic Wrappers)** que interactúan internamente con el Bus *type-erased*.

```go
package events

import (
    "context"
    "time"
)

type EventPayload interface {
    IsImmutable() bool
}

type EventHeader struct {
    EventID        string // UUIDv7
    ParentEventID  string // UUIDv7
    CorrelationID  string // TraceID
    PublisherID    string 
    Topic          string 
    SchemaVersion  int
    Timestamp      time.Time
    Priority       Priority
}

// Estructura genérica que blinda el Payload
type Event[T EventPayload] struct {
    Header  EventHeader
    Payload T
}

type PublishResult string
const (
    Accepted      PublishResult = "Accepted"
    Rejected      PublishResult = "Rejected"
    Dropped       PublishResult = "Dropped"
    Shutdown      PublishResult = "Shutdown"
    TypeMismatch  PublishResult = "TypeMismatch"
)

type Subscription[T EventPayload] interface {
    Channel() <-chan Event[T]
    Unsubscribe()
}

// Interfaz interna type-erased usada por el Host
type EventBus interface {
    publishInternal(ctx context.Context, topic string, e interface{}) PublishResult
    subscribeInternal(topicFilter string, bufferSize int) interface{}
    Shutdown(graceful bool)
}

// ==========================================
// Package-level Generic API (Strict Typing)
// ==========================================

// RegisterType asocia estáticamente un Topic a un Tipo
func RegisterType[T EventPayload](bus EventBus, topic string, schemaVersion int) {
    // Registra el contrato tipo-topic en tiempo de inicialización
}

// Publish expone tipado fuerte hacia el productor
func Publish[T EventPayload](bus EventBus, ctx context.Context, e Event[T]) PublishResult {
    // La capa interna aplica Runtime Type Validation
    return bus.publishInternal(ctx, e.Header.Topic, e)
}

// Subscribe exige tipado fuerte exacto
func Subscribe[T EventPayload](bus EventBus, topicExact string, buffer int) Subscription[T] {
    // Si 'topicExact' es un Prefix ("engine.*"), T debe ser una interfaz base polimórfica.
    return bus.subscribeInternal(topicExact, buffer).(Subscription[T])
}
```

## 24. Guarantees
| Garantía | Descripción |
|---|---|
| **O(1) Publisher Time** | Un componente Core jamás se retrasará en la llamada por culpa de un suscriptor lento. |
| **Actor-Based Consumers**| El Bus distribuye sin prestar threads. Cada suscriptor debe levantar su Actor. |
| **Self-Healing** | El sistema desaloja automáticamente a los suscriptores que amenazan la cola general, y revive su propio Dispatcher ante un Panic. |

## 25. Cross References
**Normative References:**
* **[SPEC-404](404-task-execution-engine.md):** Task Execution Engine (Publisher).
* **[SPEC-416](416-telemetry-and-metrics.md):** Telemetry (Consumer principal).
