# SPEC-407: Incremental Update Engine

## 1. Executive Summary

El *Incremental Update Engine* es el puente entre el Sistema de Archivos mutante y el Motor funcional puro (Serie 300). Su responsabilidad principal es tomar el Snapshot Inmutable actual del Workspace, comparar los deltas de texto entrantes (ediciones de usuario), y aplicar algoritmos de **Copy-on-Write (CoW)** referencial para generar un nuevo Snapshot en tiempo `O(1)` (proporcional al tamaño de la edición, no al tamaño del proyecto).

Este componente materializa la infraestructura incremental teórica (SPEC-309) a nivel de Host.

## 2. La Máquina de Estados de Actualización

Toda mutación en el Workspace Host fluye a través de este Engine mediante una transición estricta:

1. **Diff Reception:** Recibe un evento (ej. LSP `textDocument/didChange`) que contiene la URI del archivo y los reemplazos (Rangos y Textos nuevos).
2. **Copy-on-Write Initiation:** Toma el `ActiveSnapshot` global. Lo bloquea conceptualmente creando una variable local (Deep Copy referencial).
3. **Text Buffer Splice:** Se edita el `SourceText` (que es inmutable, por lo que genera un buffer nuevo o un *Rope* nuevo).
4. **Token Bump:** Se incrementa el `TextVersion` del documento.
5. **Path Spine Replication:** Se clona exclusivamente la ruta de punteros desde la raíz del Workspace hasta el Documento alterado (reusando punteros de los directorios y proyectos intactos).
6. **Snapshot Validation:** Valida que el nuevo `WorkspaceSnapshot` cumple los invariantes de integridad.
7. **Publishing:** Entrega el nuevo Snapshot al `Workspace Manager` (SPEC-402) para publicarlo globalmente.

## 3. Structural vs Trivial Edits

El motor no trata todos los tecleos por igual. Para disparar inteligentemente al Scheduler (SPEC-403), el Incremental Update Engine clasifica la edición:

* **Trivial Edit (Whitespace/Comments):** Si el motor detecta que el cambio fue un espacio, el Lexer emite tokens triviales. La `SyntaxVersion` puede que ni cambie. No hay recálculo semántico.
* **Intra-body Edit:** El usuario editó una línea dentro del bloque de una función. El AST cambia, pero la Firma (Signature Hash) del módulo no cambia. El motor detiene la propagación del cambio. Proyectos dependientes no son invalidados.
* **API/Structural Edit:** El usuario cambió el nombre de un método público o exportado. El motor detecta la caída del *Structural Hash* y marca todas las aristas entrantes en el Grafo de Dependencias como "Sucias", forzando recompilaciones cruzadas.

## 4. Conflict Resolution & Debouncing

Dado que el usuario puede teclear a ráfagas (60 teclas por segundo), crear un Snapshot completo por milisegundo es ineficiente y sobrecarga al Event Bus.

El Motor aplica **Debouncing**:
* Agrupa (Batches) los cambios que llegan en una ventana de ~50ms.
* Fusiona los deltas (Merge Edits) de texto matemáticamente.
* Crea un solo Snapshot coherente con los cambios de la ráfaga.

## 5. Interface Contract

```go
package update

import "context"

type UpdateEngine interface {
    // Aplica un lote de ediciones y devuelve el nuevo Snapshot inmutable.
    // Esto NO compila, solo actualiza los SourceTexts y la topología.
    ApplyEdits(ctx context.Context, baseSnapshot WorkspaceSnapshot, edits []FileEdit) (WorkspaceSnapshot, UpdateProfile)
}

type UpdateProfile struct {
    IsStructuralChange bool
    AffectedProjects   []string
    MergedEditsCount   int
}

type FileEdit struct {
    URI   URI
    Range TextRange
    Text  string
}
```
