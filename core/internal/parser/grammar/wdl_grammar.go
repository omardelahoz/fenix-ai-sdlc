package grammar

import (
	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/lexer"
)

// WDLGrammar implements the Grammar interface for WDL (Workflow Definition Language).
type WDLGrammar struct{}

// NewWDLGrammar creates a new WDL grammar.
func NewWDLGrammar() *WDLGrammar {
	return &WDLGrammar{}
}

// Name returns the name of the grammar.
func (g *WDLGrammar) Name() string {
	return "WDL"
}

// ParseRoot parses the root of a WDL document.
func (g *WDLGrammar) ParseRoot(ctx *ParserContext) (Node, error) {
	// WDL document starts with "workflow <name>"
	if !g.expectKeyword(ctx, "workflow") {
		return nil, ctx.Errors[0]
	}

	// Workflow name
	if !g.expect(ctx, lexer.TokenIdentifier) {
		return nil, ctx.Errors[0]
	}

	// Create document node
	doc := &WorkflowDocumentNode{
		WorkflowName: ctx.CurrentToken.Value,
		Span:         ctx.CurrentToken.Span,
	}

	// Parse workflow triggers
	for g.checkKeyword(ctx, "on") {
		trigger, err := g.parseTrigger(ctx)
		if err != nil {
			return nil, err
		}
		doc.Triggers = append(doc.Triggers, trigger)
	}

	// Parse pipelines
	for !g.check(ctx, lexer.TokenEOF) {
		if g.checkKeyword(ctx, "pipeline") {
			pipeline, err := g.parsePipeline(ctx)
			if err != nil {
				return nil, err
			}
			doc.Pipelines = append(doc.Pipelines, pipeline)
		} else {
			// Skip unknown tokens
			ctx.CurrentToken = ctx.PeekToken
			ctx.PeekToken = ctx.Lexer.NextToken()
		}
	}

	return doc, nil
}

// IsKeyword returns true if the identifier is a WDL keyword.
func (g *WDLGrammar) IsKeyword(ident string) bool {
	switch ident {
	case "workflow", "on", "pipeline", "stage", "processor",
		"policy", "depends", "fanout", "strategy", "gate",
		"timeout", "approvers", "retry", "attempts", "backoff",
		"resources", "max-workers", "memory", "priority":
		return true
	default:
		return false
	}
}

// GetKeywordType returns the token type for a keyword.
func (g *WDLGrammar) GetKeywordType(ident string) lexer.TokenType {
	switch ident {
	case "workflow":
		return lexer.TokenWorkflow
	case "on":
		return lexer.TokenOn
	case "pipeline":
		return lexer.TokenPipeline
	case "stage":
		return lexer.TokenStage
	case "processor":
		return lexer.TokenProcessor
	case "policy":
		return lexer.TokenPolicy
	case "depends":
		return lexer.TokenDepends
	case "fanout":
		return lexer.TokenFanout
	case "strategy":
		return lexer.TokenStrategy
	case "gate":
		return lexer.TokenGate
	case "timeout":
		return lexer.TokenTimeout
	case "approvers":
		return lexer.TokenApprovers
	default:
		return lexer.TokenIdentifier
	}
}

// Helper methods for parsing WDL structures

func (g *WDLGrammar) expectKeyword(ctx *ParserContext, keyword string) bool {
	if ctx.CurrentToken.Type == lexer.TokenKeyword && ctx.CurrentToken.Value == keyword {
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
		return true
	}
	ctx.Errors = append(ctx.Errors, ParseError{
		Message: "expected keyword '" + keyword + "'",
		Span:    ctx.CurrentToken.Span,
	})
	return false
}

func (g *WDLGrammar) expect(ctx *ParserContext, tokenType lexer.TokenType) bool {
	if ctx.CurrentToken.Type == tokenType {
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
		return true
	}
	ctx.Errors = append(ctx.Errors, ParseError{
		Message: "expected " + tokenType.String(),
		Span:    ctx.CurrentToken.Span,
	})
	return false
}

func (g *WDLGrammar) check(ctx *ParserContext, tokenType lexer.TokenType) bool {
	return ctx.CurrentToken.Type == tokenType
}

func (g *WDLGrammar) checkKeyword(ctx *ParserContext, keyword string) bool {
	return ctx.CurrentToken.Type == lexer.TokenKeyword && ctx.CurrentToken.Value == keyword
}

func (g *WDLGrammar) parseTrigger(ctx *ParserContext) (*TriggerNode, error) {
	g.expectKeyword(ctx, "on")

	trigger := &TriggerNode{
		Span: ctx.CurrentToken.Span,
	}

	// Event name
	trigger.EventName = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()

	// Pipeline reference
	if g.checkKeyword(ctx, "pipeline") {
		g.expectKeyword(ctx, "pipeline")
		trigger.PipelineName = ctx.CurrentToken.Value
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}

	return trigger, nil
}

func (g *WDLGrammar) parsePipeline(ctx *ParserContext) (*PipelineNode, error) {
	g.expectKeyword(ctx, "pipeline")

	pipeline := &PipelineNode{
		Span: ctx.CurrentToken.Span,
	}

	// Pipeline name
	pipeline.Name = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()

	// Parse stages
	for g.checkKeyword(ctx, "stage") {
		stage, err := g.parseStage(ctx)
		if err != nil {
			return nil, err
		}
		pipeline.Stages = append(pipeline.Stages, stage)
	}

	return pipeline, nil
}

func (g *WDLGrammar) parseStage(ctx *ParserContext) (*StageNode, error) {
	g.expectKeyword(ctx, "stage")

	stage := &StageNode{
		Span: ctx.CurrentToken.Span,
	}

	// Stage name
	stage.Name = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()

	// Parse processor
	if g.checkKeyword(ctx, "processor") {
		processor, err := g.parseProcessor(ctx)
		if err != nil {
			return nil, err
		}
		stage.Processor = processor
	}

	// Parse dependencies
	if g.checkKeyword(ctx, "depends") {
		g.expectKeyword(ctx, "depends")
		for !g.check(ctx, lexer.TokenEOF) && !g.checkKeyword(ctx, "stage") && !g.checkKeyword(ctx, "processor") {
			depName := ctx.CurrentToken.Value
			stage.Dependencies = append(stage.Dependencies, depName)
			ctx.CurrentToken = ctx.PeekToken
			ctx.PeekToken = ctx.Lexer.NextToken()
		}
	}

	// Parse fanout
	if g.checkKeyword(ctx, "fanout") {
		fanout, err := g.parseFanout(ctx)
		if err != nil {
			return nil, err
		}
		stage.Fanout = fanout
	}

	// Parse gate
	if g.checkKeyword(ctx, "gate") {
		gate, err := g.parseGate(ctx)
		if err != nil {
			return nil, err
		}
		stage.Gate = gate
	}

	return stage, nil
}

func (g *WDLGrammar) parseProcessor(ctx *ParserContext) (*ProcessorNode, error) {
	g.expectKeyword(ctx, "processor")

	processor := &ProcessorNode{
		Span: ctx.CurrentToken.Span,
	}

	// Processor name
	processor.Name = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()

	// Parse policy
	if g.checkKeyword(ctx, "policy") {
		policy, err := g.parsePolicy(ctx)
		if err != nil {
			return nil, err
		}
		processor.Policy = policy
	}

	return processor, nil
}

func (g *WDLGrammar) parsePolicy(ctx *ParserContext) (*PolicyNode, error) {
	g.expectKeyword(ctx, "policy")

	policy := &PolicyNode{
		Span: ctx.CurrentToken.Span,
	}

	// Parse policy blocks
	for !g.check(ctx, lexer.TokenEOF) && !g.checkKeyword(ctx, "stage") && !g.checkKeyword(ctx, "processor") {
		if g.checkKeyword(ctx, "retry") {
			retry, err := g.parseRetryPolicy(ctx)
			if err != nil {
				return nil, err
			}
			policy.Retry = retry
		} else if g.checkKeyword(ctx, "resources") {
			resources, err := g.parseResources(ctx)
			if err != nil {
				return nil, err
			}
			policy.Resources = resources
		} else {
			ctx.CurrentToken = ctx.PeekToken
			ctx.PeekToken = ctx.Lexer.NextToken()
		}
	}

	return policy, nil
}

func (g *WDLGrammar) parseRetryPolicy(ctx *ParserContext) (*RetryPolicyNode, error) {
	g.expectKeyword(ctx, "retry")

	retry := &RetryPolicyNode{
		Span: ctx.CurrentToken.Span,
	}

	// Parse retry parameters
	for !g.check(ctx, lexer.TokenEOF) && !g.checkKeyword(ctx, "resources") && !g.checkKeyword(ctx, "stage") {
		paramName := ctx.CurrentToken.Value
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
		
		paramValue := ctx.CurrentToken.Value
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()

		if paramName == "attempts" {
			retry.Attempts = paramValue
		} else if paramName == "backoff" {
			retry.Backoff = paramValue
		}
	}

	return retry, nil
}

func (g *WDLGrammar) parseResources(ctx *ParserContext) (*ResourcesNode, error) {
	g.expectKeyword(ctx, "resources")

	resources := &ResourcesNode{
		Span: ctx.CurrentToken.Span,
	}

	// Parse resource parameters
	for !g.check(ctx, lexer.TokenEOF) && !g.checkKeyword(ctx, "stage") {
		paramName := ctx.CurrentToken.Value
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
		
		paramValue := ctx.CurrentToken.Value
		ctx.CurrentToken = ctx.PeekToken
		ctx.Lexer.NextToken()

		if paramName == "max-workers" {
			resources.MaxWorkers = paramValue
		} else if paramName == "memory" {
			resources.Memory = paramValue
		} else if paramName == "priority" {
			resources.Priority = paramValue
		}
	}

	return resources, nil
}

func (g *WDLGrammar) parseFanout(ctx *ParserContext) (*FanoutNode, error) {
	g.expectKeyword(ctx, "fanout")

	fanout := &FanoutNode{
		Span: ctx.CurrentToken.Span,
	}

	// Parse strategy
	if g.checkKeyword(ctx, "strategy") {
		g.expectKeyword(ctx, "strategy")
		fanout.Strategy = ctx.CurrentToken.Value
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}

	// Parse processors (simplified)
	for g.checkKeyword(ctx, "processor") {
		processor, err := g.parseProcessor(ctx)
		if err != nil {
			return nil, err
		}
		fanout.Processors = append(fanout.Processors, processor)
	}

	return fanout, nil
}

func (g *WDLGrammar) parseGate(ctx *ParserContext) (*GateNode, error) {
	g.expectKeyword(ctx, "gate")

	gate := &GateNode{
		Span: ctx.CurrentToken.Span,
	}

	// Parse timeout
	if g.checkKeyword(ctx, "timeout") {
		g.expectKeyword(ctx, "timeout")
		gate.Timeout = ctx.CurrentToken.Value
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}

	// Parse approvers
	if g.checkKeyword(ctx, "approvers") {
		g.expectKeyword(ctx, "approvers")
		for !g.check(ctx, lexer.TokenEOF) && !g.checkKeyword(ctx, "stage") {
			approver := ctx.CurrentToken.Value
			gate.Approvers = append(gate.Approvers, approver)
			ctx.CurrentToken = ctx.PeekToken
			ctx.PeekToken = ctx.Lexer.NextToken()
		}
	}

	return gate, nil
}

// Node implementations for WDL

type WorkflowDocumentNode struct {
	WorkflowName string
	Triggers     []*TriggerNode
	Pipelines    []*PipelineNode
	Span         lexer.Span
}

func (n *WorkflowDocumentNode) Type() NodeType { return NodeTypeDocument }
func (n *WorkflowDocumentNode) Span() lexer.Span { return n.Span }
func (n *WorkflowDocumentNode) Children() []Node {
	var children []Node
	for _, t := range n.Triggers {
		children = append(children, t)
	}
	for _, p := range n.Pipelines {
		children = append(children, p)
	}
	return children
}

type TriggerNode struct {
	EventName    string
	PipelineName string
	Span         lexer.Span
}

func (n *TriggerNode) Type() NodeType { return NodeType("Trigger") }
func (n *TriggerNode) Span() lexer.Span { return n.Span }
func (n *TriggerNode) Children() []Node { return nil }

type PipelineNode struct {
	Name   string
	Stages []*StageNode
	Span   lexer.Span
}

func (n *PipelineNode) Type() NodeType { return NodeType("Pipeline") }
func (n *PipelineNode) Span() lexer.Span { return n.Span }
func (n *PipelineNode) Children() []Node {
	var children []Node
	for _, s := range n.Stages {
		children = append(children, s)
	}
	return children
}

type StageNode struct {
	Name         string
	Processor    *ProcessorNode
	Dependencies []string
	Fanout       *FanoutNode
	Gate         *GateNode
	Span         lexer.Span
}

func (n *StageNode) Type() NodeType { return NodeType("Stage") }
func (n *StageNode) Span() lexer.Span { return n.Span }
func (n *StageNode) Children() []Node {
	var children []Node
	if n.Processor != nil {
		children = append(children, n.Processor)
	}
	if n.Fanout != nil {
		children = append(children, n.Fanout)
	}
	if n.Gate != nil {
		children = append(children, n.Gate)
	}
	return children
}

type ProcessorNode struct {
	Name   string
	Policy *PolicyNode
	Span   lexer.Span
}

func (n *ProcessorNode) Type() NodeType { return NodeType("Processor") }
func (n *ProcessorNode) Span() lexer.Span { return n.Span }
func (n *ProcessorNode) Children() []Node {
	if n.Policy != nil {
		return []Node{n.Policy}
	}
	return nil
}

type PolicyNode struct {
	Retry     *RetryPolicyNode
	Resources *ResourcesNode
	Span      lexer.Span
}

func (n *PolicyNode) Type() NodeType { return NodeType("Policy") }
func (n *PolicyNode) Span() lexer.Span { return n.Span }
func (n *PolicyNode) Children() []Node {
	var children []Node
	if n.Retry != nil {
		children = append(children, n.Retry)
	}
	if n.Resources != nil {
		children = append(children, n.Resources)
	}
	return children
}

type RetryPolicyNode struct {
	Attempts string
	Backoff  string
	Span     lexer.Span
}

func (n *RetryPolicyNode) Type() NodeType { return NodeType("RetryPolicy") }
func (n *RetryPolicyNode) Span() lexer.Span { return n.Span }
func (n *RetryPolicyNode) Children() []Node { return nil }

type ResourcesNode struct {
	MaxWorkers string
	Memory     string
	Priority   string
	Span       lexer.Span
}

func (n *ResourcesNode) Type() NodeType { return NodeType("Resources") }
func (n *ResourcesNode) Span() lexer.Span { return n.Span }
func (n *ResourcesNode) Children() []Node { return nil }

type FanoutNode struct {
	Strategy  string
	Processors []*ProcessorNode
	Span      lexer.Span
}

func (n *FanoutNode) Type() NodeType { return NodeType("Fanout") }
func (n *FanoutNode) Span() lexer.Span { return n.Span }
func (n *FanoutNode) Children() []Node {
	var children []Node
	for _, p := range n.Processors {
		children = append(children, p)
	}
	return children
}

type GateNode struct {
	Timeout   string
	Approvers []string
	Span      lexer.Span
}

func (n *GateNode) Type() NodeType { return NodeType("Gate") }
func (n *GateNode) Span() lexer.Span { return n.Span }
func (n *GateNode) Children() []Node { return nil }
