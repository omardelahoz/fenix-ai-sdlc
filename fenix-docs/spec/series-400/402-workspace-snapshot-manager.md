# SPEC-402: Workspace Snapshot Manager Architecture

## 1. Executive Summary

Si el Workspace Model (SPEC-401) define la semántica estructural del conocimiento, el Snapshot Manager define su ciclo de vida operacional. El Snapshot Manager es el subsistema del Compiler Host responsable de orquestar cómo las vistas inmutables del Workspace se construyen, publican, adquieren, conviven y finalmente se destruyen.

En lugar de tratar al Snapshot simplemente como un objeto en memoria, este documento especifica el Snapshot como un **contrato formal** de visibilidad: una representación consistente, inmutable y versionada del Workspace en un instante exacto de tiempo, garantizando concurrencia segura para miles de consumidores simultáneos sin emplear mecanismos de bloqueo en las lecturas.

*Nota de Arquitectura:* El Snapshot **NO contiene** un Workspace; el Snapshot **es** la encarnación inmutable del Workspace en un punto del tiempo.

## 2. Design Principles

El diseño del Snapshot Manager se rige por los siguientes principios:

1. **Todo Snapshot es completamente inmutable.** Una vez congelado, su estado jamás se altera.
2. **Todo Snapshot representa un estado lógico completo.** Nunca existen vistas parciales o desincronizadas.
3. **La publicación de un Snapshot es estrictamente atómica.** Pasa de modo oculto a modo público en una sola instrucción de CPU.
4. **Las lecturas son lock-free.** Los consumidores jamás esperan por un lock para leer código.
5. **La creación reutiliza memoria mediante Copy-on-Write transitivo.**
6. **Los consumidores nunca observan transiciones internas.** 
7. **Múltiples Snapshots pueden coexistir pacíficamente.** 

## 3. Global Snapshot Invariants

Estas invariantes garantizan matemáticamente el aislamiento concurrente del Host:

* **Invariant SN-001 (Totalidad):** Todo Snapshot encarna exactamente la totalidad de un Workspace lógico.
* **Invariant SN-002 (Inmutabilidad):** Todo Snapshot es profundamente inmutable. Ningún nodo accesible puede ser modificado.
* **Invariant SN-003 (Identidad Única):** Todo Snapshot posee un `SnapshotVersion` único y monotónico.
* **Invariant SN-004 (Visibilidad Global):** Todo Snapshot marcado como `Active` es inmediatamente visible para todos los nuevos consumidores.
* **Invariant SN-005 (Unicidad de Activos):** Nunca existen dos Snapshots `Active` simultáneamente para el mismo orígen lógico.
* **Invariant SN-006 (Integridad Estructural):** Todo Snapshot activo es estructural y topológicamente válido.

## 4. Architectural Decision Record (ADR) Summary

| ID | Decisión | Justificación |
|---|---|---|
| **SN-ADR-001** | Snapshot inmutable y lock-free reads | Elimina cuellos de botella en el IDE frente a concurrencia masiva (IA/LSP). |
| **SN-ADR-002** | Structural Sharing transitivo | Permite generar miles de Snapshots con un overhead de memoria insignificante. |
| **SN-ADR-003** | Publicación atómica de Punteros | Previene condiciones de carrera o lecturas de estados intermedios inconsistentes. |
| **SN-ADR-004** | Modelo basado en Préstamos (Leases) | Desacopla la tenencia de la recolección, dejando el Snapshot puro. |
| **SN-ADR-005** | Tracking Explícito de Adquisición | Usa `Active References` en lugar de confiar en Go GC para forzar desalojos OOM. |

## 5. Non Goals

Para mantener una delimitación estricta de responsabilidades, el Snapshot Manager **NO**:
* Compila código.
* Analiza semántica o genera Abstract Syntax Trees (AST).
* Modifica directamente los documentos (responsabilidad de SPEC-407).
* Administra el Scheduler (SPEC-403).
* Publica eventos de LSP a la red.
* Define la estructura interna del Workspace (SPEC-401).

## 6. Snapshot Identity

Un Snapshot no es solo un puntero, posee metadatos operacionales robustos:

* **Snapshot ID (UUID):** Identificador opaco de la instancia.
* **Generation (Counter):** Entero incremental (`Gen 0 -> Gen 1 -> Gen 2`). Crítico para profiling e instrumentación.
* **Snapshot Version:** Marca de tiempo monotónica u ordinal.
* **Workspace Origin:** Procedencia lógica (`Disk`, `Memory`, `AI Sandbox`).
* **Creation Time:** Cuándo inició la bifurcación.
* **Parent Snapshot ID:** El Snapshot base desde el cual fue derivado.

## 7. Snapshot Lifecycle

El ciclo de vida del Snapshot es lineal y determinista:

`Created` (Se aloja en memoria) ➔ `Building` (Se clonan y mutan ramas en aislamiento) ➔ `Active` (El router apunta hacia él; se dispara el evento *Publication*) ➔ `Stale` (Reemplazado por uno más nuevo, pero con lectores vivos) ➔ `Collected` (Destruido por GC al no tener *Leases*).

*(Nota: "Publication" es el evento de transición entre Building y Active, no un estado)*.

## 8. Snapshot States (Operational)

Los *States* definen el comportamiento de memoria frente a los consumidores:

* **Building:** En mutación aislada. Invisible para el mundo.
* **Active:** Target único para cualquier nueva petición de lectura.
* **Stale:** Obsoleto, retenido únicamente para honrar *Leases* previos.
* **Pinned:** Marcado explícitamente para evadir recolección forzada (ej. Debugger Session).
* **Collectable (EligibleForCollection):** Cero préstamos activos. Muerte inminente.

## 9. Runtime View

A nivel de proceso de ejecución, el Manager orquesta la memoria así:

```text
Snapshot Manager
 ├── Active Snapshot Pointer (Atomic CAS)
 ├── Snapshot Version Index (O(1) lookups)
 ├── Snapshot Registry (All living snapshots)
 ├── Reference Tracker (Lease counts)
 ├── GC & Disposal Queue
 ├── Memory Budget Enforcer
 └── Publication Synchronization Queue
```

## 10. Snapshot Consistency Model

El Snapshot Manager ofrece promesas matemáticas de nivel base de datos sobre el código:

1. **Linealizabilidad:** Toda operación de publicación es linealizable respecto al punto exacto de adquisición del Snapshot.
2. **Ausencia de Observación Parcial:** Todo consumidor observa exactamente el Snapshot nuevo completo, o el Snapshot viejo completo. El "rasgado" estructural (tearing) es imposible.
3. **Estabilidad Aislada:** Las lecturas sobre un Snapshot permanecen idénticas y consistentes durante absolutamente toda la duración del préstamo (Lease), independientemente de las mutaciones que estén ocurriendo paralelamente en el Workspace.

## 11. Snapshot Publication Algorithm

El ensamblaje de la consistencia sigue un algoritmo estricto:

1. **Apply Edits:** Sustituir los nodos mutados aisaladamente (SPEC-407).
2. **Freeze:** Congelar la estructura. Convertirla a solo-lectura absoluta.
3. **Validate:** Confirmar integridad estructural (cero dependencias circulares).
4. **Assign Version & Generation:** Firmar la entidad.
5. **Atomic Publish:** Reemplazar el puntero global `ActiveSnapshot` vía Compare-And-Swap de hardware.
6. **Mark Previous:** Transitar el snapshot desplazado a estado `Stale`.
7. **Notify Host:** Disparar evento de finalización al Event Bus.

## 12. Snapshot Acquisition Model & Leases

Ningún componente accede a un Snapshot directamente. Se requiere un préstamo formal (*Lease*).

**Flujo de Consumo:**
`Acquire` ➔ `Lease Emitido (Ref++)` ➔ `Snapshot` ➔ `Read (Lock-free)` ➔ `Release (Ref--)` ➔ `GC Eligible`

## 13. Snapshot Visibility Rules

El aislamiento está regido por tres leyes inquebrantables de visibilidad:

* **Mientras un Snapshot está `Building`**: Ningún `Acquire` puede observarlo.
* **Cuando un Snapshot transita a `Active`**: Todos los `Acquire()` subsecuentes devolverán obligatoriamente un *Lease* para este nuevo estado.
* **Los consumidores existentes (`Stale`)**: Continúan utilizando el Snapshot antiguo al que ataron su *Lease*, sin percibir alteraciones ni interrupciones.

## 14. Structural Sharing Transitivo (Copy-on-Write)

El *Structural Sharing* es el motor que permite generar Snapshots completos en microsegundos (Overhead `O(1)`). Es una propiedad **transitiva**:

Si el usuario escribe "hola" en el archivo `auth.ts`:
1. El buffer nuevo de `auth.ts` se crea.
2. El nuevo Documento `auth.ts` *Borrowea* ese texto.
3. El Módulo `auth` se clona, enlazando al nuevo Documento y *referenciando (sharing)* a los otros 50 intactos.
4. El Proyecto se clona, enlazando al nuevo Módulo y compartiendo los intactos.
5. El Snapshot clona la raíz, compartiendo el resto del mundo.
El resultado es un Workspace lógico nuevo de 15,000 archivos que ocupa apenas un puñado de bytes nuevos en RAM.

## 15. Retention & Memory Strategy

Fénix interactúa con cientos de consumidores (*Background Consumers* como Debuggers, Indexers o Agentes IA), quienes pueden retener Snapshots `Stale` por tiempos prolongados, induciendo un riesgo de *Out-of-Memory (OOM)*.

El **Memory Budget Enforcer** previene catástrofes:
* **Soft Limit (ej. 1.5GB):** Ralentiza a los consumidores de baja prioridad (SPEC-403) y acelera el recolector de *Collectable Snapshots*.
* **Hard Limit (ej. 2.0GB / Emergency Mode):** El Manager fuerza la anulación (`context.cancel()`) de los *Leases* retenidos por los Background Consumers más antiguos, sacrificando la tarea para evadir el colapso del SO. El `Active Snapshot` es inmune a cancelación.

## 16. Snapshot GC (Garbage Collection)

El Manager utiliza un tracking abstracto de **Active References** atado a los *Leases*, sin obligar el uso de *Reference Counting* ingenuo. 
La recolección se procesa en el *GC Queue* asincrónicamente o durante ciclos de hibernación, garantizando que liberar 10,000 nodos de un AST nunca induzca *jitter* (pausas) en la latencia del teclado en el IDE.

## 17. Interface Contract

El modelo de posesión (Ownership) recae en el `SnapshotLease`. El Snapshot no sabe cómo liberarse a sí mismo.

```go
package snapshot

type SnapshotManager interface {
    // Retorna la identidad de la generación actual
    CurrentVersion() SnapshotVersion
    
    // Adquisición Formal
    Acquire(consumer ConsumerContext) (SnapshotLease, error)
    
    // Internal (SPEC-407)
    StartBuild(parentID SnapshotID) (SnapshotBuilder, error)
}

// Representa el contrato de préstamo
type SnapshotLease interface {
    // Retorna la vista inmutable
    Snapshot() WorkspaceSnapshot
    
    // Libera la retención de memoria (Reference--)
    Release()
}

// La encarnación inmutable del Workspace
type WorkspaceSnapshot interface {
    ID() SnapshotID
    Generation() int64
    State() PublicationState
    
    // Topología de lectura (equivalente a workspace.Workspace)
    Projects() ProjectIterator
    GetDocument(uri URI) (Document, error)
}

type ConsumerContext interface {
    SessionID() string
    Reason() ConsumerKind       // IDE, LSP, AI, CLI, Debugger, Indexer
    Priority() ConsumerPriority // P0, P1, P2
    Done() <-chan struct{}      // Señal OOM Emergency
}

type ConsumerKind string
const (
    IDE      ConsumerKind = "IDE"
    AI       ConsumerKind = "AI"
    CLI      ConsumerKind = "CLI"
    Debugger ConsumerKind = "Debugger"
    Indexer  ConsumerKind = "Indexer"
)
```

## 18. Snapshot Guarantees

Resumen de las promesas matemáticas del Snapshot Manager:

| Garantía | Significado |
|---|---|
| **Inmutabilidad** | Jamás muta un solo bit una vez finalizada la fase *Freeze*. |
| **Atomicidad** | Publicación indivisible (Linearizable). |
| **Consistencia** | Encarnación lógica topológicamente completa y válida. |
| **Aislamiento** | Lecturas de un Lease no se ven afectadas por mutaciones externas concurrentes. |
| **Determinismo** | Operaciones idénticas sobre el mismo Lease devuelven resultados idénticos. |
| **Escalabilidad** | Soporte lock-free para alta densidad de Background Consumers. |

## 19. Specification Cross-References

**Normative References (Language, Models & Compiler Core):**
* `Serie 100`: Especificación del Lenguaje.
* `Serie 200`: Modelos de Datos.
* `Serie 300`: Infraestructura del Compilador Puro.

**Informative References (Compiler Runtime & Host):**
* **[SPEC-400](400-compiler-host.md):** Compiler Host Architecture.
* **[SPEC-401](401-workspace-model.md):** Workspace Model Architecture.
* **[SPEC-403](403-scheduler-architecture.md):** Scheduler Architecture.
* **[SPEC-407](407-incremental-update-engine.md):** Incremental Update Engine.
* **[SPEC-415](415-session-model.md):** Session Model.
