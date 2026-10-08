# SPEC-418: Concurrency Model

## 1. Executive Summary

El Compiler Host orquesta a miles de goroutines concurrentes (LSP, AI Agents, Background Workers) que leen y escriben el código fuente. El *Concurrency Model* de Fénix rechaza los diseños tradicionales basados en *Global Interpreter Locks* (GIL) o actores puros bloqueantes, que serializan el acceso al AST. En su lugar, el sistema emplea una arquitectura puramente **Inmutable con Locks Asimétricos (RWLocks)** para lograr un escalamiento O(1) en las lecturas, asilando por completo las mutaciones en un único hilo.

## 2. Invariantes del Modelo de Memoria

Las reglas formales de memoria compartida en Fénix son no negociables:

1. **Lectura Ilimitada y Lock-Free:** Cualquier estructura dentro de un `CompilationSnapshot` (AST, tabla de símbolos, grafos de dependencias) es 100% inmutable. Un número infinito de goroutines puede leerlas en paralelo sin adquirir ningún Mutex de nodo.
2. **Mutación Aislada:** Las escrituras al modelo de Workspace NUNCA se realizan "en su lugar" (in-place mutation). Toda mutación fluye hacia un único *Write Thread* (el Coordinator).
3. **Copy-on-Write (CoW):** El Coordinator toma el Snapshot previo, aplica la diferencia referencial, crea un nuevo Snapshot Inmutable, y cambia el puntero global hacia él usando un breve `Lock()`.

## 3. Asymmetric RWLock Architecture

El modelo concurrente se resume en cómo se adquiere acceso al Workspace global:

```go
// Estructura simplificada del State global
type GlobalState struct {
    mu sync.RWMutex
    activeSnapshot *WorkspaceSnapshot
}
```

### The Reader Path (LSP, AI, CLI)
Cuando un servicio necesita procesar código:
1. Adquiere `mu.RLock()`.
2. Lee la referencia a `activeSnapshot`.
3. Libera inmediatamente `mu.RUnlock()`.
4. El hilo ejecuta su tarea pesada (e.g., Semántica, JSON-RPC response) *exclusivamente* usando su referencia local al snapshot inmutable. La operación puede tomar segundos; no bloquea a nadie.

### The Writer Path (File Watcher, IDE Changes)
Cuando el código fuente cambia (e.g., el usuario teclea):
1. El evento se envía al Worker del *Coordinator* a través de un canal (evitando concurrencia en la mutación).
2. El Coordinator computa el nuevo AST y la estructura de directorios diferencial.
3. Se construye el nuevo `newSnapshot`.
4. El Coordinator adquiere `mu.Lock()` (bloqueando temporalmente a nuevos lectores).
5. Cambia el puntero: `activeSnapshot = newSnapshot`.
6. Libera `mu.Unlock()`.

El tiempo bajo `Lock()` exclusivo es el tiempo que toma copiar un puntero (nanosegundos).

## 4. Contexts y Cancellations

Dado que múltiples lectores retienen Snapshots obsoletos (Stale Snapshots, ver SPEC-402), Fénix confía ciegamente en `context.Context` de Go para la liberación temprana de recursos.

Cada vez que el Writer Path genera un nuevo Snapshot, invoca `cancel()` en el contexto general asociado al Snapshot anterior. 
Es responsabilidad obligatoria de cada rutina en el *Reader Path* chequear periódicamente `ctx.Done()`.

* **Regla Estricta:** Un loop que atraviese ASTs pesados o evalúe resoluciones de símbolos complejas DEBE ceder el control `if ctx.Err() != nil { return nil, ctx.Err() }` al menos cada 10ms (cooperative preemption).

## 5. Prevención de Deadlocks y Data Races

* **Evitar Lock Nesting:** Jamás se puede solicitar `mu.RLock()` si la misma goroutine o subrutina ya posee `mu.Lock()`.
* **Zero Race Conditions:** Puesto que las estructuras expuestas son inmutables tras ser publicadas, las data races en la lectura del AST son matemáticamente imposibles por diseño. Las herramientas como `go test -race` se corren en CI para confirmar esta invariante.
