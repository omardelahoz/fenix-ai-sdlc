# 🔌 Integración con Providers de IA

## Visión General

Fénix utiliza una arquitectura de **adapters** para integrarse con diferentes proveedores de IA. Esto permite cambiar entre Claude, Antigravity, OpenAI, Gemini, o modelos locales sin modificar el código de Fénix ni tus archivos FDL/WDL.

## Arquitectura de Integración

```text
Fénix Runtime (Kernel)
         │
         ▼
  AIProvider Interface
         │
    ┌────┴────┬────────┬────────┐
    │         │        │        │
Claude Adapter  Antigravity  OpenAI  Local LLM
    │         │        │        │
    ▼         ▼        ▼        ▼
  Claude API  Antigravity  OpenAI  Ollama/llama.cpp
```

## Interfaces

### AIProvider

La interfaz `AIProvider` define el contrato que todos los adapters deben implementar:

```go
type AIProvider interface {
    Chat(ctx context.Context, prompt Prompt) (Response, error)
    Stream(ctx context.Context, prompt Prompt) (<-chan string, error)
    ProviderName() string
    AvailableModels() []string
}
```

### Configuración

Cada provider tiene su propia configuración:

```go
// Configuración genérica
type ProviderConfig struct {
    APIKey      string
    BaseURL     string
    Model       string
    MaxTokens   int
    Temperature float64
    Timeout     int
}

// Configuración específica de Claude
type ClaudeConfig struct {
    ProviderConfig
    Region string // e.g., "us-east-1"
}

// Configuración específica de Antigravity
type AntigravityConfig struct {
    ProviderConfig
    WorkspaceID string
    ProjectID    string
}
```

## Integración con Claude

### Configuración

```go
import "github.com/omardelahoz/fenix-ai-sdlc/core/internal/kernel"

config := &kernel.ClaudeConfig{
    ProviderConfig: kernel.ProviderConfig{
        APIKey:      "sk-ant-xxx",
        Model:       "claude-3-5-sonnet-20241022",
        MaxTokens:   100000,
        Temperature: 0.7,
        Timeout:     30,
    },
    Region: "us-east-1",
}

adapter := kernel.NewClaudeAdapter(config)
```

### Uso

```go
prompt := contracts.Prompt{
    SystemMessage: "You are a software engineer.",
    UserMessage:   "Generate a REST API for user authentication.",
}

// Non-streaming (wait for complete response)
response, err := adapter.Chat(context.Background(), prompt)
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Text)

// Streaming (real-time chunks)
stream, err := adapter.Stream(context.Background(), prompt)
if err != nil {
    log.Fatal(err)
}

for chunk := range stream {
    fmt.Print(chunk) // Print each chunk as it arrives
}
fmt.Println()
```

### Modelos Disponibles

- `claude-3-5-sonnet-20241022` - Modelo más avanzado
- `claude-3-5-haiku-20241022` - Modelo rápido y económico
- `claude-3-opus-20240229` - Modelo de alta capacidad
- `claude-3-sonnet-20240229` - Modelo balanceado
- `claude-3-haiku-20240307` - Modelo compacto

### Ventajas de Claude

- **Alta capacidad de razonamiento** - Excelente para arquitectura y diseño
- **Contexto largo** - Hasta 200K tokens
- **Seguridad** - Capacidades avanzadas de seguridad
- **Streaming completo** - Respuestas en tiempo real con eventos SSE

### Streaming con Claude

Claude soporta streaming nativo a través de su API de eventos:

```go
stream, err := adapter.Stream(ctx, prompt)
if err != nil {
    log.Fatal(err)
}

for chunk := range stream {
    fmt.Print(chunk) // Print chunks as they arrive
}
```

**Características del streaming de Claude:**
- Eventos tipo SSE (Server-Sent Events)
- Soporta múltiples tipos de eventos (message_start, content_block_delta, message_stop)
- Manejo de errores en tiempo real
- Detección de finalización automática

## Integración con Antigravity

### Configuración

```go
config := &kernel.AntigravityConfig{
    ProviderConfig: kernel.ProviderConfig{
        APIKey:      "ag-xxx",
        BaseURL:     "https://api.antigravity.com",
        Model:       "antigravity-coding-v1",
        MaxTokens:   100000,
        Temperature: 0.7,
        Timeout:     30,
    },
    WorkspaceID: "ws-123",
    ProjectID:   "proj-456",
}

adapter := kernel.NewAntigravityAdapter(config)
```

### Uso

```go
prompt := contracts.Prompt{
    SystemMessage: "You are a senior software architect.",
    UserMessage:   "Design a microservices architecture for an e-commerce platform.",
}

// Non-streaming (wait for complete response)
response, err := adapter.Chat(context.Background(), prompt)
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Text)

// Streaming (real-time chunks via SSE)
stream, err := adapter.Stream(context.Background(), prompt)
if err != nil {
    log.Fatal(err)
}

for chunk := range stream {
    fmt.Print(chunk) // Print each chunk as it arrives
}
fmt.Println()
```

### Modelos Disponibles

- `antigravity-coding-v1` - Especializado en generación de código
- `antigravity-architecture-v1` - Especializado en arquitectura
- `antigravity-analysis-v1` - Especializado en análisis de código

### Ventajas de Antigravity

- **Especializado en ingeniería** - Modelos entrenados específicamente para desarrollo de software
- **Integración con herramientas** - Conexión nativa con IDEs y herramientas de desarrollo
- **Workspace/Project** - Organización jerárquica de proyectos
- **Colaboración** - Soporte para equipos
- **Streaming completo** - Respuestas en tiempo real vía SSE

### Streaming con Antigravity

Antigravity soporta streaming a través de Server-Sent Events (SSE):

```go
stream, err := adapter.Stream(ctx, prompt)
if err != nil {
    log.Fatal(err)
}

for chunk := range stream {
    fmt.Print(chunk) // Print chunks as they arrive
}
```

**Características del streaming de Antigravity:**
- Formato SSE estándar (data: {...})
- Marcador [DONE] para finalización
- Manejo de errores en tiempo real
- Soporte para múltiples choices en streaming

## Integración con Otros Providers

### OpenAI

```go
// Implementación similar usando OpenAI API
config := &OpenAIConfig{
    ProviderConfig: ProviderConfig{
        APIKey: "sk-xxx",
        BaseURL: "https://api.openai.com/v1",
        Model:  "gpt-4",
    },
}

adapter := NewOpenAIAdapter(config)
```

### Google Gemini

```go
config := &GeminiConfig{
    ProviderConfig: ProviderConfig{
        APIKey: "xxx",
        BaseURL: "https://generativelanguage.googleapis.com/v1beta",
        Model:  "gemini-pro",
    },
}

adapter := NewGeminiAdapter(config)
```

### Local LLMs (Ollama)

```go
config := &OllamaConfig{
    ProviderConfig: ProviderConfig{
        BaseURL: "http://localhost:11434/api",
        Model:  "codellama:13b",
    },
}

adapter := NewOllamaAdapter(config)
```

## Cómo Fénix Usa los Adapters

### En el Kernel

El Kernel selecciona el provider apropiado basado en la configuración del ExecutionPackage:

```go
type ExecutionPackage struct {
    Resources ResourceAllocation
    // ...
}

type ResourceAllocation struct {
    Provider string // "claude-3.5-sonnet", "antigravity-coding-v1", etc.
    // ...
}
```

### Selector de Provider

```go
func GetProvider(providerName string, config ProviderConfig) (AIProvider, error) {
    switch providerName {
    case "claude":
        return NewClaudeAdapter(config.(*ClaudeConfig))
    case "antigravity":
        return NewAntigravityAdapter(config.(*AntigravityConfig))
    case "openai":
        return NewOpenAIAdapter(config.(*OpenAIConfig))
    default:
        return nil, fmt.Errorf("unknown provider: %s", providerName)
    }
}
```

### Ejemplo de Ejecución

```go
// Crear ExecutionPackage con provider Claude
pkg := &contracts.ExecutionPackage{
    Resources: contracts.ResourceAllocation{
        Provider: "claude-3.5-sonnet",
    },
    // ...
}

// Kernel selecciona el adapter apropiado
provider, err := GetProvider(pkg.Resources.Provider, claudeConfig)
if err != nil {
    return err
}

// Ejecutar prompt
response, err := provider.Chat(ctx, prompt)
```

## Estrategias de Uso

### 1. Provider Único

Usar un solo provider para todo el proyecto:

```yaml
# En configuración del proyecto
default_provider: claude-3.5-sonnet
```

### 2. Provider por Tarea

Usar diferentes providers para diferentes tipos de tareas:

```yaml
providers:
  architecture: claude-3-5-sonnet  # Mejor razonamiento
  coding: antigravity-coding-v1     # Especializado en código
  review: gpt-4                    # Equilibrado
  testing: claude-3-haiku          # Rápido y económico
```

### 3. Fallback Strategy

Usar un provider principal con fallbacks:

```yaml
fallback_chain:
  - claude-3-5-sonnet
  - gpt-4
  - local-llama-codellama
```

### 4. Cost Optimization

Usar modelos más económicos para tareas simples:

```yaml
cost_optimization:
  enabled: true
  simple_tasks: claude-3-haiku
  complex_tasks: claude-3-5-sonnet
```

## Configuración en FDL

Puedes especificar el provider en tu archivo FDL:

```fdl
product MyProduct

implementation
    Backend: Go (Chi router)
    
    ai_config
        default_provider: claude-3-5-sonnet
        architecture_provider: claude-3-5-sonnet
        coding_provider: antigravity-coding-v1
```

## Configuración en WDL

En workflows, puedes especificar providers por stage:

```wdl
pipeline DevelopmentPipeline

stage Architecture
    processor ArchitectureProcessor
        provider claude-3-5-sonnet

stage Coding
    processor BackendProcessor
        provider antigravity-coding-v1
```

## Seguridad y Privacidad

### API Keys

Las API keys deben:
- **Nunca** ser committed al repositorio
- Usarse desde variables de entorno
- O desde un archivo `.env` local (en `.gitignore`)

```bash
# .env.example
CLAUDE_API_KEY=sk-ant-xxx
ANTIGRAVITY_API_KEY=ag-xxx
OPENAI_API_KEY=sk-xxx
```

### Local LLMs

Para máxima privacidad, usa modelos locales:

```yaml
provider: local
local_config:
    base_url: http://localhost:11434
    model: codellama:13b
```

Ventajas:
- Datos nunca salen de tu máquina
- Sin costos de API
- Control total del modelo

## Métricas y Monitoreo

Fénix rastrea métricas por provider:

```go
type ProviderMetrics struct {
    TotalRequests      int
    TotalTokens       int
    TotalCost         float64
    AverageLatency    time.Duration
    ErrorRate         float64
    ProviderName      string
}
```

Esto permite:
- Comparar performance entre providers
- Optimizar costos
- Detectar problemas específicos de un provider

## Próximos Pasos

Para implementar un nuevo provider, sigue estos pasos detallados:

### Paso 1: Definir la Configuración

Agrega la configuración específica en `core/pkg/contracts/provider.go`:

```go
// [Provider]Config contains [Provider]-specific configuration.
type [Provider]Config struct {
    ProviderConfig
    // Provider-specific fields
    CustomField string
}
```

### Paso 2: Implementar el Adapter

Crea un nuevo archivo en `core/internal/kernel/[provider]_adapter.go`:

```go
package kernel

import (
    "context"
    "github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// [Provider]Adapter implements AIProvider for [Provider].
type [Provider]Adapter struct {
    config *contracts.[Provider]Config
    client *http.Client
}

func New[Provider]Adapter(config *contracts.[Provider]Config) *[Provider]Adapter {
    // Initialize adapter
    return &[Provider]Adapter{
        config: config,
        client: &http.Client{},
    }
}

func (a *[Provider]Adapter) Chat(ctx context.Context, prompt contracts.Prompt) (contracts.Response, error) {
    // Implement chat logic
    return contracts.Response{}, nil
}

func (a *[Provider]Adapter) Stream(ctx context.Context, prompt contracts.Prompt) (<-chan string, error) {
    // Implement streaming logic
    return nil, nil
}

func (a *[Provider]Adapter) ProviderName() string {
    return "[provider]"
}

func (a *[Provider]Adapter) AvailableModels() []string {
    return []string{"model-1", "model-2"}
}
```

### Paso 3: Agregar al Factory

Agrega el provider al `ProviderFactory` en `core/internal/kernel/provider_factory.go`:

```go
case "[provider]":
    [provider]Config, ok := config.(*contracts.[Provider]Config)
    if !ok {
        return nil, fmt.Errorf("invalid config type for [provider] provider")
    }
    return New[Provider]Adapter([provider]Config), nil
```

### Paso 4: Escribir Tests

Crea tests en `core/internal/kernel/[provider]_adapter_test.go`:

```go
package kernel

import (
    "context"
    "testing"
    "github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

func Test[Provider]Adapter_Chat(t *testing.T) {
    config := &contracts.[Provider]Config{
        ProviderConfig: contracts.ProviderConfig{
            Model: "test-model",
        },
    }
    
    adapter := New[Provider]Adapter(config)
    
    prompt := contracts.Prompt{
        UserMessage: "Test",
    }
    
    response, err := adapter.Chat(context.Background(), prompt)
    if err != nil {
        t.Fatalf("Chat failed: %v", err)
    }
    
    if response.Text == "" {
        t.Error("Expected non-empty response")
    }
}
```

### Paso 5: Documentar

Actualiza esta guía con la documentación del nuevo provider.

## Ejemplo Completo: Ollama

He implementado **Ollama** como ejemplo completo de un nuevo provider:

- **Configuración:** `OllamaConfig` en `core/pkg/contracts/provider.go`
- **Adapter:** `OllamaAdapter` en `core/internal/kernel/ollama_adapter.go`
- **Factory:** Agregado a `ProviderFactory`
- **Streaming:** Implementado completamente
- **Model Discovery:** Consulta `/api/tags` para obtener modelos disponibles

### Uso de Ollama

```go
config := &contracts.OllamaConfig{
    ProviderConfig: contracts.ProviderConfig{
        BaseURL:     "http://localhost:11434",
        Model:       "codellama:13b",
        MaxTokens:   8192,
        Temperature: 0.7,
    },
    Host: "localhost:11434",
}

factory := NewProviderFactory()
provider, err := factory.CreateProvider("ollama", config)

response, err := provider.Chat(ctx, prompt)
```

### Ventajas de Ollama

- **Privacidad total** - Datos nunca salen de tu máquina
- **Sin costos de API** - Gratis una vez instalado
- **Modelos variados** - Llama, Mistral, CodeLlama, etc.
- **Offline** - Funciona sin internet
- **Customizable** - Fine-tuning de modelos

## Referencias

- [Claude API Documentation](https://docs.anthropic.com/claude/reference/messages)
- [Antigravity Documentation](https://docs.antigravity.com)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Ollama Documentation](https://ollama.com/docs)
