# SPEC-415: Session Model

## 1. Executive Summary

El *Session Model* permite que un único Compiler Host sirva a **múltiples clientes concurrentemente** de manera aislada y segura, sobre el mismo Workspace. 
A diferencia de un compilador de línea de comandos que se ejecuta una vez por usuario, un Host tipo demonio podría estar escuchando peticiones de:
1. El usuario humano en su ventana principal del IDE.
2. Un Agente de IA que está "imaginando" (RAG/Sandbox) un refactoring en una rama paralela de la memoria.
3. Un cliente de Terminal que ejecuta scripts de validación.

## 2. Definición de Session

Una Sesión es el contexto de conexión de un cliente específico. Posee:
* **Session ID:** Identificador único criptográfico (UUID).
* **Capabilities:** Un registro de lo que el cliente soporta (ej., el IDE soporta formateo, la CLI no).
* **Isolation Context:** (Opcional) Un overlay VFS privado si el cliente requiere un "Sandbox" (ver abajo).

## 3. Sandbox y Virtual Workspaces

El caso de uso principal del Session Model avanzado es el **AI Agent Sandboxing**.

Cuando un Agente de IA solicita escribir código, no debería modificar el Workspace global del usuario directamente, ya que arruinaría el código en pantalla del desarrollador humano si se equivoca.

En su lugar, el Host implementa *Virtual Workspaces* ligados a una Session:
1. El agente inicia una sesión con `IsSandbox: true`.
2. El Host crea un `MemoryMount` en el VFS (SPEC-408) exclusivo para ese Agente.
3. El Agente escribe y recompila su código. El Compiler Host crea Snapshots que *solo* son visibles para el Session ID de ese agente.
4. Si la IA confirma su código, emite un comando `CommitSession`, que fusiona su Overlay privado al Workspace Global del usuario (generando el Snapshot definitivo).

## 4. Lifecycle de una Sesión

1. **Handshake:** El cliente (vía LSP o RPC) se conecta y envía credenciales/capacidades.
2. **Active:** El cliente retiene conexiones. Sus Snapshots locales se mantienen vivos.
3. **Disconnected / Timeout:** Si el socket se rompe, la Sesión entra en `GracePeriod`.
4. **Disposal:** Tras el GracePeriod (ej. 30 segs), el Host destruye los overlays virtuales del Sandbox y limpia el VFS.

## 5. Interface Contract

```go
package session

import "context"

type SessionManager interface {
    // Registro
    CreateSession(client rpc.Client, config SessionConfig) (Session, error)
    GetSession(id string) (Session, error)
    
    // Sandbox (Solo para IA y Testing)
    PromoteSandboxToGlobal(ctx context.Context, sessionID string) error
}

type Session interface {
    ID() string
    Config() SessionConfig
    
    // Obtiene una vista del Workspace "desde la perspectiva" del cliente.
    // Si es Sandbox, verá sus propios overrides. Si no, verá el global.
    AcquireWorkspaceView() (WorkspaceSnapshot, error)
}
```
