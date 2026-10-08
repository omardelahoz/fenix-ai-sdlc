# 🚀 Fénix AI SDLC - Quick Start

> Entiende Fénix en 5 minutos

## ¿Qué es Fénix?

Fénix es un **Sistema Operativo de Ingeniería con IA** que transforma cómo desarrollamos software. En lugar de usar prompts ambiguos en Markdown, Fénix usa **DSLs declarativos** (FDL, WDL, PMF) que son compilados por un pipeline industrial antes de que la IA nunca los vea.

**La diferencia fundamental:**

| Enfoque Tradicional | Fénix |
|---------------------|-------|
| Prompt: "Haz un login con Azure AD" | FDL: Estructura formal compilada |
| IA interpreta libremente | IA recibe AST validado |
| Resultado impredecible | Resultado determinista |
| Contexto perdido | Knowledge Graph persistente |

## 🏗️ Arquitectura en 30 segundos

```text
FDL (Producto) → Lexer → Parser → AST → Engineering Processors → IRs → AI Execution
```

1. **Escribes FDL** - Lenguaje declarativo que describe QUÉ quieres construir
2. **Fénix Compila** - Lexer/Parser en Go valida la estructura y genera AST
3. **Processors Transforman** - Múltiples IRs progresivos (Product → Requirements → Architecture → Code)
4. **AI Ejecuta** - Claude/Gemini recibe contratos formales, no prompts libres

## 📝 Tu Primer FDL

Crea un archivo `booking.fdl`:

```fdl
product BookingAutomation

vision
    Automate reservation process for small clinics.

users
    Administrator
    Customer

constraints
    OAuth2
    AzureAD
    GDPR
    MaxLatency: 200ms

feature Authentication
    story Login
        requirement
            User authenticates using Azure AD.
            Session expires after 15 minutes of inactivity.
        acceptance
            Given a valid Azure AD token
            When the user attempts to login
            Then the system issues a JWT and redirects to Dashboard
        architecture
            API
                AuthenticationService
                JwtProvider
                UserRepository
        tests
            Integration: Verify AzureAD handshake
            Unit: JWT decoding and validation

implementation
    Backend: Go (Chi router)
    Frontend: React (Next.js)

release
    v1.0
```

## 🔄 Tu Primer WDL (Workflow)

Crea `workflow.wdl`:

```wdl
workflow StandardSaaSDevelopment

on ProductUpdated
    pipeline MainSaaSPipeline

pipeline MainSaaSPipeline

stage Discovery
    processor DiscoveryProcessor

stage Requirements
    depends Discovery
    processor RequirementsProcessor

stage Development
    depends Requirements
    fanout
        strategy wait-all
        processors
            BackendProcessor
            FrontendProcessor
            DatabaseProcessor
```

## 🎯 Conceptos Clave

### Los 3 DSLs de Fénix

| DSL | Propósito | Describe |
|-----|-----------|----------|
| **FDL** | Fénix Definition Language | QUÉ es el producto (Entidades, UI, Arquitectura) |
| **WDL** | Workflow Definition Language | CÓMO se ejecutan los workflows (Pipelines, CI/CD) |
| **PMF** | Processor Manifest Format | CÓMO se integran plugins/processors |

### El Pipeline de Compilación

```text
FDL/WDL/PMF (Texto)
    ↓
Lexer (Tokens)
    ↓
Parser (AST)
    ↓
Binder (Resolución de símbolos)
    ↓
Semantic Analyzer (Validación)
    ↓
Canonical Language IR (Modelo de dominio)
    ↓
Engineering IR (Artefactos físicos)
    ↓
Ejecución (Claude/Gemini/etc)
```

### El Kernel (ExecutionPackage)

El Kernel de Fénix usa un **Process Control Block (PCB)** inmutable:

```go
type ExecutionPackage struct {
    Metadata  PackageMetadata
    Task      TaskDescriptor
    Executor  ExecutorDescriptor
    Snapshot  ExecutionSnapshot
    Policies  ExecutionPolicies
    Resources ResourceAllocation
    Contracts Contracts
}
```

**Regla de oro:** El ExecutionPackage es **completamente inmutable**. Cualquier cambio genera una nueva versión.

## 🚦 Estado Actual del Proyecto

- ✅ **Documentación**: Completa (SPEC-000 a SPEC-419)
- ✅ **Arquitectura**: Definida en detalle
- 🚧 **Implementación**: En fase inicial (solo interfaces básicas)
- 🎯 **Próximo paso**: Implementación del Kernel (Serie 000)

## 📚 Documentación Completa

- **[README.md](README.md)** - Visión general del proyecto
- **[SPEC-000](fenix-docs/spec/series-000/)** - Kernel API Contracts
- **[SPEC-100](fenix-docs/spec/series-100/)** - Engineering Runtime & IR
- **[SPEC-200](fenix-docs/spec/series-200/)** - Repository Layout & Build
- **[SPEC-300](fenix-docs/spec/series-300/)** - Language System Architecture
- **[SPEC-400](fenix-docs/spec/series-400/)** - Compiler Host Architecture
- **[ADRs](fenix-docs/adr/)** - Architecture Decision Records

## 🛠️ Próximos Pasos

1. Lee [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) para configurar tu entorno
2. Revisa [ROADMAP.md](ROADMAP.md) para entender el plan de implementación
3. Explora los ejemplos en [fenix-examples/](fenix-examples/)
4. Contribuye siguiendo el roadmap priorizando el Kernel (Serie 000)

## 💡 Por qué Fénix es diferente

**A diferencia de Cursor, Claude Code, Aider:**

- ✅ Compilador industrial con Lexer/Parser/AST
- ✅ DSLs formales en lugar de Markdown
- ✅ Knowledge Graph persistente (Product Memory)
- ✅ IRs progresivos en lugar de prompts directos
- ✅ Modelo agnóstico (Claude, Gemini, local LLMs)
- ✅ Determinismo matemático en lugar de "intento"

**Fénix no es un "AI coding framework" - es un Sistema Operativo de Ingeniería.**

---

¿Listo para construir el futuro del desarrollo de software? 🚀
