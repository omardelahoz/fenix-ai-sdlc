# SPEC-410: Diagnostics Pipeline

## 1. Executive Summary

El *Diagnostics Pipeline* es el motor que procesa, deduplica, filtra y publica los errores y advertencias (Diagnostics) generados por el compilador hacia el exterior (IDE/CLI). Dado que el Semantic Analyzer de Fénix es agresivo y veloz, puede generar miles de errores en un proyecto mal configurado. Emitir todo esto en bruto por el LSP congestionaría el IDE y el Event Bus.

El Pipeline asegura que el usuario vea la información más relevante, rápido, y sin "parpadeos" (flickering) cuando el código está en estados intermedios inconsistentes.

## 2. Fases del Pipeline

Cuando una pasada de compilación (Lexical, Sintáctica o Semántica) finaliza, arroja un slice de `[]Diagnostic`. Estos entran al Pipeline:

### Phase 1: Deduplication
* Fénix puede evaluar la misma regla múltiples veces durante análisis incrementales cruzados.
* El pipeline usa un Hash compuesto por `(URI + Range + ErrorCode)` para eliminar errores matemáticamente idénticos.

### Phase 2: Stale Filtering
* Si un diagnóstico pertenece al Snapshot $S_0$, pero mientras el pipeline procesaba el usuario tecleó generando el Snapshot $S_1$, el pipeline detecta que los diagnósticos de $S_0$ son *Stale* (obsoletos).
* **Regla:** Jamás se publican diagnósticos de un Snapshot obsoleto. Se aborta la publicación para evitar que el IDE muestre errores fantasma.

### Phase 3: Cascading Error Suppression
* Fénix implementa supresión inteligente de errores en cascada.
* Si ocurre un error léxico (ej. un string mal cerrado), el Parser inevitablemente fallará, y el Binder fallará al no encontrar las variables.
* **Regla:** Si una línea tiene errores `Lexical`, se filtran y ocultan los errores `Syntax` y `Semantic` para ese mismo nodo. Esto reduce el ruido cognitivo del usuario ("Arregla primero el string").

### Phase 4: Throttling & Pagination
* Si un archivo tiene 2000 errores, el LSP no soporta enviar todos en un JSON masivo sin colapsar el hilo de UI de VSCode/IntelliJ.
* El Pipeline envía los primeros N (ej. 250) errores más severos (Errors > Warnings > Hints) y trunca el resto agregando un diagnóstico sintético: `"Existen 1750 errores más omitidos..."`.

## 3. Diagnostic Data Model

Un Diagnóstico no es un simple string, es una estructura rica.

```go
type Severity int

const (
    Error Severity = iota
    Warning
    Information
    Hint
)

type Diagnostic struct {
    Code     string // e.g., "FNX-3012"
    Message  string
    Severity Severity
    Range    TextRange
    
    // Telemetría de origen
    SourcePhase string // "Lexer", "Parser", "Binder", "Semantic"
    SnapshotID  string
    
    // Quick Fixes (Code Actions) atados al error
    SuggestedFixes []CodeFix 
}
```

## 4. Emisión y Suscripción

Una vez filtrado el set final, el Pipeline envuelve la lista en el evento `DiagnosticsUpdated` y lo arroja al `Event Bus` (SPEC-405).
El `LSP Server` está suscrito a este evento y lo traduce a notificaciones `textDocument/publishDiagnostics` hacia el cliente.
