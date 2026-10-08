# 🛠️ Fénix AI SDLC - Developer Guide

> Guía completa para desarrolladores que contribuyen a Fénix

## 📋 Tabla de Contenidos

- [Requisitos Previos](#requisitos-previos)
- [Configuración del Entorno](#configuración-del-entorno)
- [Arquitectura del Proyecto](#arquitectura-del-proyecto)
- [Estructura del Repositorio](#estructura-del-repositorio)
- [Guía de Contribución](#guía-de-contribución)
- [Testing](#testing)
- [Workflow de Desarrollo](#workflow-de-desarrollo)
- [Recursos Adicionales](#recursos-adicionales)

## 🔧 Requisitos Previos

### Software Necesario

- **Go 1.22+** - Lenguaje principal de Fénix
- **Git** - Control de versiones
- **Make o Task** - Automatización de build (usamos Taskfile.yml)
- **Editor** - VSCode, GoLand, o tu editor preferido

### Opcional pero Recomendado

- **Docker** - Para pruebas de integración
- **gRPC Tools** - Para generar código desde protobuf
- **protoc** - Compilador de Protocol Buffers

## 🚀 Configuración del Entorno

### 1. Clonar el Repositorio

```bash
git clone https://github.com/omardelahoz/fenix-ai-sdlc.git
cd fenix-ai-sdlc
```

### 2. Configurar Go Workspace

Fénix usa Go Workspaces (go.work) para manejar múltiples módulos:

```bash
# Verificar go.work existe
cat go.work

# Si no existe, crearlo
go work init

# Agregar módulos
go work use ./core
go work use ./cmd/fenix
go work use ./plugins/claude
go work use ./plugins/openai
go work use ./sdk
```

### 3. Instalar Dependencias

```bash
# Descargar dependencias de todos los módulos
go work download

# O módulo por módulo
cd core && go mod download
cd cmd/fenix && go mod download
```

### 4. Instalar Herramientas de Desarrollo

```bash
# Instalar Task (si prefieres Task sobre Make)
go install github.com/go-task/task/v3/cmd/task@latest

# Instalar herramientas de Go
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/segmentio/gohorse@latest  # Para testing
```

### 5. Configurar VSCode (Opcional)

Crear `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v"],
  "go.testTimeout": "30s"
}
```

## 🏗️ Arquitectura del Proyecto

### Visión de Alto Nivel

Fénix está organizado en capas concéntricas:

```text
┌─────────────────────────────────────────┐
│         CLI / IDE / Dashboard           │  (Interfaces de usuario)
├─────────────────────────────────────────┤
│      Compiler Host (SPEC-400)           │  (Workspace, Snapshots, Scheduler)
├─────────────────────────────────────────┤
│      Language System (SPEC-300)         │  (Lexer, Parser, AST, Binder)
├─────────────────────────────────────────┤
│      Engineering Runtime (SPEC-100)      │  (IR Pipeline, Processors)
├─────────────────────────────────────────┤
│      Kernel (SPEC-000)                  │  (ExecutionPackage, Memory Manager)
├─────────────────────────────────────────┤
│      AI Drivers (Adapters)              │  (Claude, Gemini, OpenAI)
└─────────────────────────────────────────┘
```

### Series de Especificaciones

| Serie | Tema | Responsabilidad |
|-------|------|-----------------|
| **SPEC-000** | Kernel API Contracts | ExecutionPackage, Memory Manager, State Machine |
| **SPEC-100** | Engineering Runtime & IR | FDL/WDL/PMF Semantics, IR Pipeline, Processors |
| **SPEC-200** | Repository Layout | Estructura física, build system, dependencias |
| **SPEC-300** | Language System | Lexer, Parser, AST, Binder, Semantic Analyzer |
| **SPEC-400** | Compiler Host | Workspace, Snapshots, Scheduler, Event Bus |

## 📁 Estructura del Repositorio

Fénix sigue el layout consolidado de monorepo definido en [SPEC-200](fenix-docs/spec/series-200/200-repository-layout.md):

```text
/fenix-ai-sdlc
├── go.work                 # Go workspace definition
├── Taskfile.yml            # Automation runner
├── README.md
├── QUICKSTART.md
├── DEVELOPER_GUIDE.md
├── ROADMAP.md
│
├── fenix-docs/             # Documentación técnica
│   ├── spec/
│   │   ├── series-000/    # Kernel specs
│   │   ├── series-100/    # Runtime specs
│   │   ├── series-200/    # Repository specs
│   │   ├── series-300/    # Language specs
│   │   └── series-400/    # Host specs
│   └── adr/               # Architecture Decision Records
│
├── fenix-engine/          # Core engine (módulo Go existente)
│   ├── cmd/               # CLI commands
│   ├── pkg/
│   │   ├── clarification/
│   │   ├── contextbuilder/
│   │   └── provider/
│   └── go.mod
│
├── fenix-system/          # Sistema de alto nivel (por implementar)
│   ├── agents/
│   ├── managers/
│   ├── pipelines/
│   ├── profiles/
│   ├── skills/
│   ├── standards/
│   └── templates/
│
├── fenix-adapters/        # Integraciones con AI providers
│
├── fenix-examples/        # Ejemplos de FDL/WDL
│
└── fenix-installer/       # Instalador y setup
```

**Nota:** La estructura actual será reorganizada según SPEC-200 para consolidar el core en un solo módulo `/core`.

## 🤝 Guía de Contribución

### Flujo de Trabajo

1. **Fork y Clonar**
   ```bash
   git clone https://github.com/tu-usuario/fenix-ai-sdlc.git
   cd fenix-ai-sdlc
   git remote add upstream https://github.com/omardelahoz/fenix-ai-sdlc.git
   ```

2. **Crear Rama**
   ```bash
   git checkout -b feature/tu-feature
   # o
   git checkout -b fix/tu-fix
   ```

3. **Hacer Cambios**
   - Sigue las convenciones de código de Go
   - Escribe tests para tu código
   - Actualiza la documentación si es necesario

4. **Commits**
   ```bash
   git add .
   git commit -m "feat: add kernel execution package structure"
   ```

   **Convención de Commits:**
   - `feat:` Nueva funcionalidad
   - `fix:` Bug fix
   - `docs:` Cambios en documentación
   - `refactor:` Refactorización
   - `test:` Agregar/actualizar tests
   - `chore:` Mantenimiento

5. **Push y PR**
   ```bash
   git push origin feature/tu-feature
   ```
   Luego crea un Pull Request en GitHub.

### Estándares de Código

#### Go Code Style

- Sigue [Effective Go](https://golang.org/doc/effective_go)
- Usa `gofmt` para formatear
- Usa `goimports` para gestionar imports
- Ejecuta `golangci-lint` antes de commitear

```bash
go fmt ./...
goimports -w .
golangci-lint run
```

#### Nombres de Paquetes

- Usa nombres cortos, en minúsculas, una palabra
- Evita guiones bajos o mezcla de mayúsculas
- El nombre del paquete debe ser el último componente del import path

```go
// ✅ Bueno
package kernel
package scheduler
package ast

// ❌ Malo
package kernelPackage
package Kernel_Scheduler
```

#### Interfaces

- Define interfaces donde las consumas, no donde las implementas
- Usa interfaces pequeñas y enfocadas
- Nombra interfaces con `-er` suffix cuando sea apropiado

```go
// ✅ Bueno
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Scheduler interface {
    Schedule(ctx context.Context, task Task) error
}
```

#### Documentación

- Documenta todos los paquetes exportados
- Documenta funciones exportadas con ejemplos
- Usa godoc format

```go
// Package kernel provides the core execution engine for Fénix.
//
// The kernel manages ExecutionPackages, which are immutable
// process control blocks that travel through the system.
package kernel

// Schedule adds a task to the scheduler queue.
//
// Example:
//
//   handle, err := scheduler.Schedule(ctx, task)
//   if err != nil {
//       return err
//   }
//   defer handle.Cancel()
func (s *Scheduler) Schedule(ctx context.Context, task Task) (*Handle, error) {
    // ...
}
```

### Testing

#### Escribir Tests

- Usa tabla-driven tests para casos múltiples
- Usa `t.Run()` para subtests
- Mock dependencias externas

```go
func TestExecutionPackage_Validate(t *testing.T) {
    tests := []struct {
        name    string
        pkg     *ExecutionPackage
        wantErr bool
    }{
        {
            name: "valid package",
            pkg:  &ExecutionPackage{...},
            wantErr: false,
        },
        {
            name: "missing metadata",
            pkg:  &ExecutionPackage{Metadata: PackageMetadata{}},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.pkg.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### Ejecutar Tests

```bash
# Todos los tests
go test ./...

# Tests con cobertura
go test -cover ./...

# Tests con race detector
go test -race ./...

# Tests específicos
go test ./internal/kernel

# Tests con verbosidad
go test -v ./internal/kernel
```

### Revisión de Código

Antes de crear un PR:

1. ✅ Código formateado (`go fmt`)
2. ✅ Imports ordenados (`goimports`)
3. ✅ Linting pasa (`golangci-lint`)
4. ✅ Tests pasan (`go test`)
5. ✅ Documentación actualizada
6. ✅ Especificaciones referenciadas (SPEC-XXX)

## 🧪 Testing

### Tipos de Tests

Fénix usa una pirámide de testing:

```text
          E2E Tests
         /          \
    Integration Tests
     /              \
  Unit Tests
```

#### Unit Tests

- Prueban funciones/paquetes individuales
- Rápidos (< 1ms)
- Sin dependencias externas

```bash
go test ./internal/kernel -run TestExecutionPackage
```

#### Integration Tests

- Prueban interacción entre componentes
- Usan mocks o servicios reales en memoria
- Moderadamente rápidos (< 100ms)

```bash
go test ./internal/kernel -run TestSchedulerIntegration
```

#### E2E Tests

- Prueban el sistema completo
- Usan dependencias reales (DB, AI APIs)
- Lentos (> 1s)

```bash
go test ./test/e2e -run TestFullWorkflow
```

### Mocking

Usa interfaces para mockear dependencias:

```go
// Mock implementation
type MockExecutor struct {
    ExecuteFunc func(ctx context.Context, pkg *ExecutionPackage) error
}

func (m *MockExecutor) Execute(ctx context.Context, pkg *ExecutionPackage) error {
    return m.ExecuteFunc(ctx, pkg)
}

// Usage in test
mock := &MockExecutor{
    ExecuteFunc: func(ctx context.Context, pkg *ExecutionPackage) error {
        return nil
    },
}
```

## 🔄 Workflow de Desarrollo

### Feature Development

1. **Leer Especificaciones**
   - Lee las specs relevantes (SPEC-XXX)
   - Entiende los invariants y contratos
   - Revisa ADRs relacionados

2. **Diseñar**
   - Diseña interfaces primero
   - Piensa en inmutabilidad y concurrencia
   - Considera el modelo de memoria de Go

3. **Implementar**
   - Escribe la estructura de datos
   - Implementa la lógica
   - Escribe tests paralelamente

4. **Validar**
   - Ejecuta tests
   - Ejecuta linting
   - Verifica que cumple las specs

5. **Documentar**
   - Actualiza godoc
   - Agrega ejemplos
   - Actualiza ADRs si es necesario

### Debugging

#### Logging

Usa structured logging:

```go
import "log/slog"

logger := slog.Default()
logger.Info("scheduling task",
    "task_id", task.ID,
    "priority", task.Priority,
    "snapshot_id", snapshot.ID,
)
```

#### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

## 📚 Recursos Adicionales

### Documentación Interna

- [QUICKSTART.md](QUICKSTART.md) - Introducción rápida
- [ROADMAP.md](ROADMAP.md) - Plan de implementación
- [fenix-docs/spec/](fenix-docs/spec/) - Especificaciones técnicas completas
- [fenix-docs/adr/](fenix-docs/adr/) - Architecture Decision Records

### Recursos Externos

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [The Go Blog](https://blog.golang.org/)

### Comunidad

- GitHub Issues: Reportar bugs y feature requests
- GitHub Discussions: Preguntas y discusiones
- Documentación: Actualizar y mejorar

## 🎯 Próximos Pasos

1. Configura tu entorno local
2. Lee las especificaciones del Kernel (SPEC-000)
3. Revisa el roadmap (ROADMAP.md)
4. Elige una tarea del roadmap
5. ¡Empieza a contribuir!

---

¿Tienes preguntas? Revisa las [especificaciones](fenix-docs/spec/) o abre una [discusión](https://github.com/omardelahoz/fenix-ai-sdlc/discussions).
