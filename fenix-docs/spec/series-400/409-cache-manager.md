# SPEC-409: Cache Manager

## 1. Executive Summary

El *Cache Manager* es el componente que permite a Fénix alcanzar latencias de milisegundos en *Large Workspaces*. Dado que el Pipeline de Compilación puro (Serie 300) repite muchos cálculos en archivos que no han cambiado, el Cache Manager intercepta y almacena resultados intermedios.

Se encarga de las políticas de almacenamiento en RAM, el desalojo (Eviction/GC) basado en el ciclo de vida de los Snapshots, y la serialización/deserialización (Persistencia) hacia el disco duro.

## 2. Tipos de Caches

El Host no utiliza un "caché mágico universal", sino estratos altamente tipados:

### 2.1. Structural Hash Cache (Level 1)
* **Qué almacena:** Hashes SHA-256 de las firmas públicas de módulos/proyectos. (Ver SPEC-309).
* **Ubicación:** 100% RAM, respaldado en disco.
* **Costo de retención:** Minúsculo (KBs). Nunca se desaloja a menos que el proyecto sea removido.

### 2.2. AST Cache (Level 2)
* **Qué almacena:** Nodos sintácticos (Green Trees) resultados del Parser.
* **Ubicación:** RAM (referenciado por el `WorkspaceSnapshot`).
* **Costo de retención:** Medio. El AST de Fénix es altamente compacto, pero un workspace gigante consume MBs. Desalojado solo bajo alta presión de memoria (SPEC-413).

### 2.3. Semantic Model Cache (Level 3)
* **Qué almacena:** Tipos resueltos, tablas de símbolos atadas, control-flow graphs.
* **Ubicación:** RAM puramente temporal.
* **Costo de retención:** Muy alto. Si no hay archivos abiertos en un proyecto (Lazy state), este caché se destruye y se recalcula bajo demanda.

## 3. Políticas de Desalojo (Eviction)

Fénix no usa un algoritmo LRU/LFU ciego de tiempo (como Redis). El desalojo está **Topológicamente Orientado**.

1. **Snapshot-Driven Eviction:** Cuando el `SnapshotManager` recolecta (GC) el snapshot $S_0$, todos los ASTs y Símbolos que existían en $S_0$ pero no en el actual $S_1$ quedan huérfanos y son limpiados naturalmente por el GC de Go.
2. **Pressure-Driven Eviction (Shedding):** Si el `ResourceManager` declara *High Memory Pressure*, el Cache Manager fuerza un trim:
   * Identifica qué Proyectos no tienen documentos activos (`Opened`).
   * Anula los punteros a sus *Semantic Model Caches*.

## 4. Persistencia en Disco (Cold Start)

Para evitar la penalización de parsear miles de archivos al abrir el IDE:
1. Durante el estado `Shutdown` (SPEC-417), el Cache Manager vuelca el *Structural Hash Cache* y los *ASTs* comprimidos en `.fenix/cache/v1_xxxx.pack`.
2. Al arrancar (Bootstrap), los mapea a memoria (`mmap`), permitiendo lecturas *Zero-Copy* directo desde el disco (page cache del OS).
3. La validación de vigencia se hace usando los *mtimes* de los archivos o *Version Tokens* de un watcher externo (Watchman/fsnotify).

## 5. Interface Contract

```go
package cache

import "context"

type CacheManager interface {
    // Lectura (O(1) lookups)
    GetSyntaxTree(uri URI, version int64) (SyntaxNode, bool)
    GetSemanticModel(projectID string, version int64) (SemanticModel, bool)
    
    // Escritura
    PutSyntaxTree(uri URI, version int64, tree SyntaxNode)
    
    // Gestión del ciclo de vida
    Trim(pressure PressureLevel)
    PersistToDisk(ctx context.Context, dirPath string) error
    HydrateFromDisk(ctx context.Context, dirPath string) error
}
```
