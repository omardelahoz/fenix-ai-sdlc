# SPEC-408: File System Abstraction (VFS)

## 1. Executive Summary

Si el Compiler Host dependiera directamente de las llamadas de sistema del OS (`os.File`, `os.Stat`), quedaría permanentemente acoplado a discos físicos. Esto limitaría severamente su despliegue en entornos como Cloud IDEs, navegadores web (WASM), o sistemas con Agentes de IA que operan sobre miles de archivos generados "al vuelo" que nunca deben tocar el disco físico por rendimiento.

La *File System Abstraction* especifica el Virtual File System (VFS) de Fénix, el cual aísla al Workspace Manager y a la Serie 300 de la realidad del almacenamiento.

## 2. VFS Topology

El Workspace Manager (SPEC-401) lee el código exclusivamente a través del VFS. Este VFS está compuesto en capas o *Mounts*:

1. **Physical Mount:** Lee archivos reales en `C:\` o `/home`. Utilizado por defecto en un IDE local.
2. **Overlay / Memory Mount:** Carpeta virtual donde residen los "archivos sucios" (modificados por el usuario pero aún no guardados en disco).
3. **Generated / Scratch Mount:** Utilizado por Agentes de IA o Macros. Archivos sintéticos que existen puramente en RAM.
4. **Network Mount (Opcional):** Permite resolver paquetes de terceros transparentemente desde GitHub o registros NPM/Go.

## 3. Overlay Architecture

El patrón dominante en Fénix VFS es el *Overlay FS*. 

Cuando el IDE notifica `OpenDocument` y luego `DidChange`:
* El VFS enmascara (Shadows) el archivo físico. 
* Cualquier petición del Parser a ese archivo será enrutada al `Memory Mount`.
* Si el archivo físico cambia en disco al mismo tiempo (ej. por un `git checkout`), el Host recibe el evento pero debe orquestar una política de resolución de conflictos.
* Cuando el IDE emite `DidSave`, el archivo del Memory Mount se vuelca al Physical Mount y la sombra desaparece.

## 4. Path Normalization

Fénix usa URIs (Uniform Resource Identifiers) universalmente en lugar de OS paths, alineándose con el estándar Language Server Protocol (LSP).

* Todo Path (ej. `C:\dev\file.fdl`) se convierte a URI (`file:///c:/dev/file.fdl`).
* Los Virtual files generados por la IA pueden usar esquemas custom (`fenix-ai://agent-44/synthetic.fdl`).
* El compilador **nunca** usa los separadores `\` o `/` nativos del OS en las rutas lógicas del Grafo de Dependencias.

## 5. Interface Contract

```go
package vfs

import "context"

// Scheme define el tipo de File System subyacente.
type Scheme string

const (
    FileScheme Scheme = "file"
    MemoryScheme Scheme = "fenix-mem"
    AiScratchScheme Scheme = "fenix-ai"
)

type VFS interface {
    // Lectura
    ReadFile(ctx context.Context, uri URI) ([]byte, error)
    Stat(ctx context.Context, uri URI) (FileInfo, error)
    ReadDir(ctx context.Context, uri URI) ([]FileInfo, error)
    
    // Escritura (Generalmente restringido al Overlay)
    WriteFile(ctx context.Context, uri URI, content []byte) error
    DeleteFile(ctx context.Context, uri URI) error
    
    // Monitoreo
    Watch(ctx context.Context, uri URI, callback func(Event)) error
}

type FileInfo interface {
    URI() URI
    Size() int64
    ModTime() int64
    IsDir() bool
}
```

## 6. WASM y Cloud Compatibility

Dado el objetivo de Fénix de correr potencialmente dentro del navegador, ninguna implementación del `VFS` puede usar de forma incondicional el paquete `os`. Todas las llamadas nativas deben estar separadas mediante *build tags* (ej. `//go:build !wasm`), proveyendo una implementación de `JS-Interop` o indexDB para la variante Web.
