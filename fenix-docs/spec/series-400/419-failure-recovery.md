# SPEC-419: Failure Recovery

## 1. Executive Summary

El Compiler Host corre durante horas o días en el ordenador del usuario. Durante este tiempo, se enfrentará inevitablemente a situaciones anómalas: plugins que crashean, fallos de disco al leer un archivo temporal, cortes de red descargando dependencias, o pánicos (panics) inesperados en el propio analizador sintáctico ante código altamente corrupto.

La especificación de *Failure Recovery* define las políticas de resiliencia. El objetivo supremo es: **El Compiler Host nunca debe morir ante un error local de compilación o de I/O**.

## 2. Niveles de Gravedad (Fault Domains)

### 2.1. Soft Failures (Errores de Usuario)
* **Caso:** El usuario escribe código sintácticamente inválido.
* **Manejo:** El Lexer/Parser no crashea. Usa tokens de sincronización (SPEC-304) para saltar el error y emitir un diagnóstico. No es responsabilidad del Host recuperarse, el pipeline simplemente sigue su curso.

### 2.2. Isolated Hard Failures (Panic en Pipeline)
* **Caso:** Una función en el Binder (SPEC-307) sufre un *panic* (`index out of range`) evaluando un closure complejo.
* **Manejo:**
  1. El `Execution Engine` (SPEC-404) atrapa el panic con un `defer recover()`.
  2. El pipeline se aborta limpiamente para ese Snapshot.
  3. El Host emite un Diagnóstico Sintético: `Internal Compiler Error (ICE): Panic in Binder. Please report this issue.` atado a la línea que lo provocó.
  4. El Worker sobrevive y toma otra tarea.
* **Resiliencia:** El usuario ve un error rojo en su pantalla, pero el IDE sigue funcionando. Si borra la línea ofensiva, el Compiler Host recompila exitosamente en el siguiente tecleo.

### 2.3. Subsystem Failures (IO / Red)
* **Caso:** El *Cache Manager* intenta leer el caché de disco y falla por corrupción del file system.
* **Manejo (Fallback to Pure):** El Host asume que el subsistema está roto. Registra una advertencia y desactiva silenciosamente el caché. Pasa a recompilar desde cero (Cold Start). El usuario experimenta lentitud temporal, no un crasheo.

### 2.4. Fatal System Failures (OOM / OS Limits)
* **Caso:** El Host consume los 2GB de RAM permitidos y el SO amenaza con matar el proceso.
* **Manejo:** El `Resource Manager` (SPEC-413) gatilla el protocolo `Critical Watermark`. Purgamos el 100% de la memoria prescindible, forzamos la liberación (Free) al SO, y abortamos todas las peticiones del IDE devolviendo errores temporales. Si a pesar de esto el SO mata el proceso (SIGKILL), dependemos del *Crash Recovery* del arranque.

## 3. Crash Recovery (Arranque de Emergencia)

Si el binario muere inesperadamente, dejará un archivo `.fenix-lock` en el `.fenix` folder.
Al reiniciar, el Lifecycle Manager (SPEC-414):
1. Detecta el `.lock` sucio.
2. Infiere que los cachés persistentes (AST, Hash) pueden estar corruptos o mal serializados.
3. Borra el directorio `.fenix/cache` entero de forma preventiva.
4. Ejecuta un Cold Boot limpio.

## 4. Retries y Backoff en Red

Para las fallas en el `Background Queue` relacionadas a la red (ej. descargando un módulo):
* No crashean el compilador.
* Siguen una política de *Exponential Backoff*. Si falla 3 veces, la tarea P2 muere y se emite un error estático en el archivo fuente: `Unresolved Import (Network Error)`.
