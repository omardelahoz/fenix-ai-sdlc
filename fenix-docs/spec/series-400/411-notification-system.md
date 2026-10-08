# SPEC-411: Notification System

## 1. Executive Summary

El Compiler Host de Fénix es un ente *Server-Push*. No depende exclusivamente de que los clientes hagan polling para saber si la compilación falló o si hay progreso. El *Notification System* es la interfaz de salida, responsable de convertir los eventos asíncronos del *Event Bus* (SPEC-405) en mensajes de protocolo de red hacia el IDE, CLI, o Agentes de IA.

## 2. LSP Notifier (Language Server Protocol)

La integración primaria es el IDE (VSCode, IntelliJ). El Notification System implementa notificaciones LSP críticas:

### 2.1. Server Progress (`$/progress`)
El Host puede tardar en cargar un Workspace grande. Enviar a los usuarios al estado `WorkspaceLoading` en silencio da la ilusión de que el IDE se congeló.
El Notifier escucha eventos del Lifecycle Manager y emite tokens de progreso:
* `begin`: "Fénix: Indexando Workspace (0/15000)".
* `report`: "Fénix: Analizando módulos Semánticos...".
* `end`: "Fénix: Ready".

### 2.2. Diagnostics Push (`textDocument/publishDiagnostics`)
Recibe los paquetes limpios del *Diagnostics Pipeline* (SPEC-410) y los rutea al cliente LSP para que pinte subrayados rojos y verdes en el texto.

### 2.3. File Status and Caching Warnings
Si el Host detecta que está rozando el límite de memoria (SPEC-413) y ha entrado en modo hibernación, el Notifier empuja una alerta visual al IDE (`window/showMessage`) advirtiendo al usuario que la experiencia podría estar degradada.

## 3. AI Gateway (Agent Subscriptions)

A diferencia del IDE, que es una interfaz gráfica (GUI), los Agentes de IA requieren telemetría estructurada para saber si es seguro hacer una consulta de razonamiento (RAG).

El *AI Gateway* permite a un Agente establecer un WebSocket o Long-Polling HTTP y suscribirse a estados específicos:
* Un agente puede suscribirse a `WaitForReady(ProjectX)`.
* El Notification System retiene la conexión hasta que el Host sale del estado `Compiling` y entra en `Ready`. 
* Esto evita que la IA haga consultas semánticas al vacío que devuelvan resultados parciales mientras la compilación estaba en curso.

## 4. Architecture and Buffering

El Notification System es un simple *Consumidor* del Event Bus. Para evitar degradar el rendimiento interno:
1. Emplea buffers anulares (Ring Buffers) de red.
2. Si un cliente (ej. un AI Agent) tiene una conexión de red lenta, el Notification System aplica *Tail Drop* (elimina las notificaciones más antiguas en cola) para evitar presionar la RAM del Compiler Host.

## 5. Interface Contract

```go
package notification

import "context"

type NotificationSystem interface {
    // Inicializa la escucha desde el EventBus subyacente.
    Start(bus EventBus)
    
    // Rutas de comunicación
    RegisterLSPClient(client rpc.Client)
    RegisterAIAgent(agentID string, ch chan<- NotificationPayload)
    
    // Detiene gracefulmente los push.
    Stop(ctx context.Context)
}
```
