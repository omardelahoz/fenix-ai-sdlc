# 🗺️ Fénix AI SDLC - Roadmap

> Plan de implementación por fases

## 📊 Visión General

Fénix es un sistema complejo con múltiples capas interdependientes. Este roadmap divide la implementación en fases lógicas, comenzando por el Kernel (fundación) y construyendo hacia arriba.

**Estrategia:** Bottom-up - Del Kernel hacia las capas superiores

```text
Fase 1: Kernel (SPEC-000) → Fase 2: Language System (SPEC-300) → 
Fase 3: Engineering Runtime (SPEC-100) → Fase 4: Compiler Host (SPEC-400) →
Fase 5: Integración & CLI
```

## 🎯 Fase 1: Kernel Foundation (SPEC-000) ✅ COMPLETADO

**Objetivo:** Establecer el núcleo del sistema operativo de ingeniería.

### Fase 1.1: Core Data Structures ✅ COMPLETADO
- [x] Implementar `ExecutionPackage` (PCB inmutable)
- [x] Implementar `ExecutionHandle` (estado mutable)
- [x] Implementar `PackageMetadata`
- [x] Implementar `TaskDescriptor`
- [x] Implementar `ExecutorDescriptor`
- [x] Implementar `ExecutionSnapshot`
- [x] Implementar `ExecutionPolicies`
- [x] Implementar `ResourceAllocation`
- [x] Implementar `Contracts`

**Archivos:** `core/pkg/contracts/execution.go`, `core/pkg/contracts/execution_handle.go`, `core/pkg/contracts/execution_test.go`

**Especificaciones:** [SPEC-001](fenix-docs/spec/series-000/001-execution-package.md)

### Fase 1.2: Kernel API Contracts ✅ COMPLETADO
- [x] Implementar interfaz `Kernel`
- [x] Implementar `MemoryManager`
- [x] Implementar `ExecutionManager`
- [x] Implementar `StateMachine`
- [x] Implementar `ContextModel`

**Archivos:** `core/pkg/contracts/kernel.go`

**Especificaciones:** [SPEC-003](fenix-docs/spec/series-000/003-kernel-api-contracts.md)

### Fase 1.3: Kernel State Machine ✅ COMPLETADO
- [x] Implementar estados del Kernel
- [x] Implementar transiciones de estado
- [x] Implementar validación de invariantes
- [x] Implementar recuperación de fallos

**Archivos:** `core/internal/kernel/state_machine.go`

**Especificaciones:** [SPEC-004](fenix-docs/spec/series-000/004-kernel-state-machine.md)

### Fase 1.4: Testing del Kernel ✅ COMPLETADO
- [x] Unit tests para todas las estructuras
- [x] Integration tests para state machine
- [x] Tests de inmutabilidad
- [x] Tests de concurrencia (race detector)

**Archivos:** `core/internal/kernel/kernel_test.go`, `core/internal/kernel/execution_manager_test.go`, `core/internal/kernel/concurrency_test.go`

**Hitos:**
- ✅ Kernel puede crear y validar ExecutionPackages
- ✅ Kernel puede gestionar ciclos de vida
- ✅ Kernel puede detectar invariant violations
- ✅ Kernel maneja concurrencia correctamente

### Fase 1.5: Implementaciones del Kernel ✅ COMPLETADO
- [x] Implementar `FenixKernel` (Kernel principal)
- [x] Implementar `ExecutionManagerImpl` (gestión de handles)
- [x] Implementar `MemoryManagerImpl` (gestión de cache L0)
- [x] Implementar `StateMachineImpl` (máquina de estados)
- [x] Implementar `Scheduler` (planificador de tareas)

**Archivos:** `core/internal/kernel/kernel.go`, `core/internal/kernel/execution_manager.go`, `core/internal/kernel/memory_manager.go`, `core/internal/kernel/state_machine.go`, `core/internal/kernel/scheduler.go`

### Fase 1.6: Integración con Providers ✅ COMPLETADO (BONUS)
- [x] Implementar `ClaudeAdapter` (Anthropic Claude)
- [x] Implementar `AntigravityAdapter` (Antigravity platform)
- [x] Implementar `OllamaAdapter` (local LLMs)
- [x] Implementar `ProviderFactory` (factory de providers)
- [x] Documentación de integración

**Archivos:** `core/internal/kernel/claude_adapter.go`, `core/internal/kernel/antigravity_adapter.go`, `core/internal/kernel/ollama_adapter.go`, `core/internal/kernel/provider_factory.go`, `fenix-docs/INTEGRATION.md`

### Fase 1.7: CLI Básica ✅ COMPLETADO
- [x] Implementar CLI `fenix init`
- [x] Implementar CLI `fenix execute`
- [x] Implementar CLI `fenix status`

**Archivos:** `cmd/fenix/main.go`

---

## 🔤 Fase 2: Language System (SPEC-300)

**Objetivo:** Implementar el pipeline de compilación para FDL/WDL/PMF.

### Fase 2.1: Lexical System
- [ ] Implementar `Lexer`
- [ ] Implementar `Token` types
- [ ] Implementar `Trivia` handling
- [ ] Implementar tokenización para FDL
- [ ] Implementar tokenización para WDL
- [ ] Implementar tokenización para PMF

**Especificaciones:** [SPEC-302](fenix-docs/spec/series-300/302-lexical-system.md)

### Fase 2.2: Parser Engine
- [ ] Implementar `ParserEngine` genérico
- [ ] Implementar `Grammar` interface
- [ ] Implementar `FDLGrammar`
- [ ] Implementar `WDLGrammar`
- [ ] Implementar `PMFGrammar`
- [ ] Implementar error recovery

**Especificaciones:** [SPEC-304](fenix-docs/spec/series-300/304-parser-engine.md)

### Fase 2.3: CST/AST (Green/Red Trees)
- [ ] Implementar `GreenNode` (inmutable)
- [ ] Implementar `RedNode` (facade)
- [ ] Implementar `SyntaxTree`
- [ ] Implementar `SyntaxCanonicalizer`
- [ ] Implementar conversion CST → AST

**Especificaciones:** [SPEC-303](fenix-docs/spec/series-300/303-green-red-trees.md), [SPEC-305](fenix-docs/spec/series-300/305-syntax-canonicalizer.md)

### Fase 2.4: Symbol System & Binder
- [ ] Implementar `Symbol` types
- [ ] Implementar `SymbolTable`
- [ ] Implementar `Binder`
- [ ] Implementar `BindingTable`
- [ ] Implementar resolución de referencias

**Especificaciones:** [SPEC-306](fenix-docs/spec/series-300/306-symbol-system.md), [SPEC-307](fenix-docs/spec/series-300/307-binder.md)

### Fase 2.5: Semantic Analyzer
- [ ] Implementar `SemanticAnalyzer`
- [ ] Implementar validación de tipos
- [ ] Implementar validación de constraints
- [ ] Implementar diagnóstico de errores

**Especificaciones:** [SPEC-308](fenix-docs/spec/series-300/308-semantic-analyzer.md)

### Fase 2.6: Language IR
- [ ] Implementar `CanonicalLanguageIR`
- [ ] Implementar conversión AST → IR
- [ ] Implementar validación de IR

**Especificaciones:** [SPEC-311](fenix-docs/spec/series-300/311-language-ir.md)

### Fase 2.7: Testing del Language System
- [ ] Tests de lexer para FDL/WDL/PMF
- [ ] Tests de parser con casos edge
- [ ] Tests de semantic analyzer
- [ ] Tests de conversión AST → IR

**Hitos:**
- ✅ FDL puede ser parseado a AST
- ✅ WDL puede ser parseado a AST
- ✅ PMF puede ser parseado a AST
- ✅ Semantic analysis detecta errores
- ✅ AST puede convertirse a Language IR

---

## ⚙️ Fase 3: Engineering Runtime (SPEC-100)

**Objetivo:** Implementar el pipeline de IRs y los Engineering Processors.

### Fase 3.1: Engineering Runtime IR
- [ ] Implementar `EngineeringRuntimeIR`
- [ ] Implementar interfaz para Processors
- [ ] Implementar registro de Processors

**Especificaciones:** [SPEC-100](fenix-docs/spec/series-100/100-engineering-runtime-ir.md)

### Fase 3.2: Engineering IR Builder
- [ ] Implementar `EngineeringIRBuilder`
- [ ] Implementar conversión Language IR → Engineering IR
- [ ] Implementar validación de Engineering IR

**Especificaciones:** [SPEC-102](fenix-docs/spec/series-100/102-engineering-ir.md)

### Fase 3.3: Core Processors
- [ ] Implementar `ProductDiscoveryProcessor`
- [ ] Implementar `RequirementsProcessor`
- [ ] Implementar `ArchitectureProcessor`
- [ ] Implementar `DevelopmentProcessor`

**Especificaciones:** [SPEC-100](fenix-docs/spec/series-100/100-engineering-runtime-ir.md)

### Fase 3.4: Pipeline Engine
- [ ] Implementar `RuntimePipelineEngine`
- [ ] Implementar ejecución secuencial de Processors
- [ ] Implementar fan-out/fan-in
- [ ] Implementar manejo de errores

**Especificaciones:** [SPEC-103](fenix-docs/spec/series-100/103-runtime-pipeline-engine.md)

### Fase 3.5: Engineering Artifact Model
- [ ] Implementar `EngineeringArtifact`
- [ ] Implementar almacenamiento de artefactos
- [ ] Implementar versión de artefactos

**Especificaciones:** [SPEC-110](fenix-docs/spec/series-100/110-engineering-artifact-model.md)

### Fase 3.6: Testing del Engineering Runtime
- [ ] Tests de processors individuales
- [ ] Tests de pipeline completo
- [ ] Tests de conversión IR → IR
- [ ] Tests de generación de artefactos

**Hitos:**
- ✅ FDL puede transformarse a Product IR
- ✅ Product IR puede transformarse a Requirements IR
- ✅ Requirements IR puede transformarse a Architecture IR
- ✅ Architecture IR puede transformarse a Development IR
- ✅ Processors pueden ejecutarse en pipeline

---

## 🏠 Fase 4: Compiler Host (SPEC-400)

**Objetivo:** Implementar el host que orquesta todo el sistema.

### Fase 4.1: Workspace Model
- [ ] Implementar `Workspace`
- [ ] Implementar `Project`
- [ ] Implementar `Module`
- [ ] Implementar `Document`
- [ ] Implementar `DependencyGraph`
- [ ] Implementar version tokens

**Especificaciones:** [SPEC-401](fenix-docs/spec/series-400/401-workspace-model.md)

### Fase 4.2: Snapshot Manager
- [ ] Implementar `SnapshotManager`
- [ ] Implementar `WorkspaceSnapshot`
- [ ] Implementar `SnapshotLease`
- [ ] Implementar Copy-on-Write
- [ ] Implementar GC de snapshots

**Especificaciones:** [SPEC-402](fenix-docs/spec/series-400/402-workspace-snapshot-manager.md)

### Fase 4.3: Scheduler
- [ ] Implementar `Scheduler`
- [ ] Implementar `ExecutionPolicy`
- [ ] Implementar `ExecutionPlan`
- [ ] Implementar admission control
- [ ] Implementar dependency resolution

**Especificaciones:** [SPEC-403](fenix-docs/spec/series-400/403-scheduler-architecture.md)

### Fase 4.4: Task Execution Engine
- [ ] Implementar `TaskExecutionEngine`
- [ ] Implementar `WorkerPool`
- [ ] Implementar `ExecutionSession`
- [ ] Implementar `ExecutionAttempt`
- [ ] Implementar fault isolation

**Especificaciones:** [SPEC-404](fenix-docs/spec/series-400/404-task-execution-engine.md)

### Fase 4.5: Event Bus
- [ ] Implementar `EventBus`
- [ ] Implementar pub/sub
- [ ] Implementar topic routing
- [ ] Implementar QoS policies

**Especificaciones:** [SPEC-405](fenix-docs/spec/series-400/405-event-bus.md)

### Fase 4.6: Background Services
- [ ] Implementar `BackgroundServicesOrchestrator`
- [ ] Implementar Daemon lifecycle
- [ ] Implementar throttling
- [ ] Implementar core daemons

**Especificaciones:** [SPEC-406](fenix-docs/spec/series-400/406-background-services.md)

### Fase 4.7: VFS & Cache
- [ ] Implementar `VFS` (Virtual File System)
- [ ] Implementar `CacheManager`
- [ ] Implementar estratos de cache
- [ ] Implementar persistencia

**Especificaciones:** [SPEC-408](fenix-docs/spec/series-400/408-file-system-abstraction.md), [SPEC-409](fenix-docs/spec/series-400/409-cache-manager.md)

### Fase 4.8: Supporting Services
- [ ] Implementar `DiagnosticsPipeline`
- [ ] Implementar `NotificationSystem`
- [ ] Implementar `ResourceManager`
- [ ] Implementar `LifecycleManager`
- [ ] Implementar `SessionModel`
- [ ] Implementar `Telemetry`

**Especificaciones:** SPEC-410 a SPEC-416

### Fase 4.9: Testing del Compiler Host
- [ ] Tests de workspace loading
- [ ] Tests de snapshot lifecycle
- [ ] Tests de scheduling
- [ ] Tests de event bus
- [ ] Tests de background services
- [ ] Integration tests completos

**Hitos:**
- ✅ Workspace puede cargar proyectos
- ✅ Snapshots pueden crearse y gestionarse
- ✅ Tasks pueden ser scheduladas
- ✅ Eventos pueden publicarse/suscribirse
- ✅ Sistema puede manejar concurrencia masiva

---

## 🔌 Fase 5: Integración & CLI

**Objetivo:** Crear interfaces de usuario y conectores externos.

### Fase 5.1: CLI
- [ ] Implementar comandos básicos CLI
- [ ] Implementar `fenix init`
- [ ] Implementar `fenix build`
- [ ] Implementar `fenix validate`
- [ ] Implementar `fenix run`

### Fase 5.2: AI Adapters
- [ ] Implementar `ClaudeAdapter`
- [ ] Implementar `OpenAIAdapter`
- [ ] Implementar `GeminiAdapter`
- [ ] Implementar interfaz `AIProvider` genérica

**Especificaciones:** Ver fenix-engine/pkg/provider/provider.go

### Fase 5.3: LSP Server
- [ ] Implementar LSP server básico
- [ ] Implementar `textDocument/completion`
- [ ] Implementar `textDocument/hover`
- [ ] Implementar `textDocument/definition`
- [ ] Implementar `textDocument/diagnostics`

### Fase 5.4: Examples & Templates
- [ ] Crear ejemplos completos de FDL
- [ ] Crear ejemplos completos de WDL
- [ ] Crear templates de proyectos
- [ ] Crear documentación de ejemplos

### Fase 5.5: Testing de Integración
- [ ] E2E tests del CLI
- [ ] Tests de adapters
- [ ] Tests de LSP
- [ ] Tests con ejemplos reales

**Hitos:**
- ✅ CLI puede ejecutar comandos básicos
- ✅ AI adapters pueden conectar con providers
- ✅ LSP server puede servir IDE
- ✅ Ejemplos funcionan end-to-end

---

## 🎯 Fase 6: Polish & Production Readiness

**Objetivo:** Optimizar, documentar y preparar para producción.

### Fase 6.1: Performance
- [ ] Benchmarking de componentes críticos
- [ ] Optimización de memory footprint
- [ ] Optimización de latencia
- [ ] Profile-guided optimization

### Fase 6.2: Security
- [ ] Security audit
- [ ] Validación de inputs
- [ ] Sanitización de outputs
- [ ] Rate limiting

### Fase 6.3: Documentation
- [ ] Completar documentación de API
- [ ] Crear tutoriales
- [ ] Crear videos demo
- [ ] Crear sitio web

### Fase 6.4: CI/CD
- [ ] Configurar GitHub Actions
- [ ] Automatizar tests
- [ ] Automatizar builds
- [ ] Automatizar releases

### Fase 6.5: Monitoring
- [ ] Implementar metrics collection
- [ ] Implementar logging estructurado
- [ ] Implementar tracing
- [ ] Implementar alerting

**Hitos:**
- ✅ Sistema es performante
- ✅ Sistema es seguro
- ✅ Documentación es completa
- ✅ CI/CD automatizado
- ✅ Monitoring operativo

---

## 📈 Progreso Actual

| Fase | Estado | Progreso |
|------|--------|----------|
| **Fase 1: Kernel** | � Completado | 100% |
| **Fase 2: Language System** | 🔴 No iniciado | 0% |
| **Fase 3: Engineering Runtime** | 🔴 No iniciado | 0% |
| **Fase 4: Compiler Host** | 🔴 No iniciado | 0% |
| **Fase 5: Integración & CLI** | � Parcial | 50% |
| **Fase 6: Polish** | 🔴 No iniciado | 0% |

**Estado General:** � **Fase 1 (Kernel) completada, listo para Fase 2**

---

## 🤝 Cómo Contribuir

1. Revisa las fases anteriores y elige una tarea
2. Lee las especificaciones relevantes (SPEC-XXX)
3. Crea una rama: `git checkout -b feat/implementation-task`
4. Implementa siguiendo [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)
5. Crea un PR con referencia a este roadmap

## 📅 Timeline Estimado

**Nota:** Estos son estimados aproximados. El timeline depende de recursos disponibles.

- **Fase 1:** 4-6 semanas
- **Fase 2:** 6-8 semanas
- **Fase 3:** 4-6 semanas
- **Fase 4:** 8-10 semanas
- **Fase 5:** 4-6 semanas
- **Fase 6:** 4-6 semanas

**Total estimado:** 30-42 semanas (7-10 meses)

## 🚀 Hitos Principales

| Hito | Descripción | Fecha Estimada |
|------|-------------|----------------|
| **M1: Kernel MVP** | Kernel funcional con ExecutionPackage | Semana 6 |
| **M2: FDL Parser** | FDL puede parsearse a AST | Semana 14 |
| **M3: IR Pipeline** | Pipeline de IRs funcional | Semana 20 |
| **M4: Host MVP** | Compiler Host básico funcional | Semana 30 |
| **M5: CLI MVP** | CLI puede ejecutar workflows | Semana 36 |
| **M6: v1.0** | Primer release estable | Semana 42 |

---

¿Listo para contribuir? Revisa [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) y elige una tarea de Fase 1 🚀
