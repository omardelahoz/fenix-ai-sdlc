package parser

import (
	"fmt"
	"testing"

	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/grammar"
	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/lexer"
)

// ExampleParseFDL demonstrates how to parse an FDL document.
func ExampleParseFDL() {
	source := `
product BookingSystem

vision
    A modern booking system for healthcare appointments.

users
    Patient
    Doctor
    Administrator

constraints
    MaxLatency: 100ms
    OfflineSupport

feature AppointmentBooking
    story BookAppointment
        requirement
            Patient can book appointments
`

	// Create parser with FDL grammar
	fdlGrammar := grammar.NewFDLGrammar()
	parser := NewParser(fdlGrammar, source, "example.fdl")

	// Parse
	doc, err := parser.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	// Access parsed document
	if docNode, ok := doc.(*grammar.DocumentNode); ok {
		fmt.Printf("Product: %s\n", docNode.ProductName)
		if docNode.Vision != nil {
			fmt.Printf("Vision: %s\n", docNode.Vision.Text)
		}
		fmt.Printf("Features: %d\n", len(docNode.Features))
	}
}

func TestLexer(t *testing.T) {
	source := `product BookingSystem`

	lex := lexer.NewLexer(source, "test.fdl")
	tokens := lex.Tokenize()

	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != lexer.TokenProduct {
		t.Errorf("Expected TokenProduct, got %s", tokens[0].Type)
	}

	if tokens[1].Type != lexer.TokenIdentifier {
		t.Errorf("Expected TokenIdentifier, got %s", tokens[1].Type)
	}

	if tokens[2].Type != lexer.TokenEOF {
		t.Errorf("Expected TokenEOF, got %s", tokens[2].Type)
	}
}

func TestParser(t *testing.T) {
	source := `product BookingSystem`

	fdlGrammar := grammar.NewFDLGrammar()
	parser := NewParser(fdlGrammar, source, "test.fdl")

	doc, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected non-nil document")
	}

	docNode, ok := doc.(*grammar.DocumentNode)
	if !ok {
		t.Fatal("Expected DocumentNode")
	}

	if docNode.ProductName != "BookingSystem" {
		t.Errorf("Expected ProductName 'BookingSystem', got '%s'", docNode.ProductName)
	}
}

func TestParserWithVision(t *testing.T) {
	source := `
product BookingSystem

vision
    A modern booking system.
`

	fdlGrammar := grammar.NewFDLGrammar()
	parser := NewParser(fdlGrammar, source, "test.fdl")

	doc, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	docNode, ok := doc.(*grammar.DocumentNode)
	if !ok {
		t.Fatal("Expected DocumentNode")
	}

	if docNode.Vision == nil {
		t.Error("Expected Vision node")
	}
}

func TestParserWithFeature(t *testing.T) {
	source := `
product BookingSystem

feature AppointmentBooking
    story BookAppointment
`

	fdlGrammar := grammar.NewFDLGrammar()
	parser := NewParser(fdlGrammar, source, "test.fdl")

	doc, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	docNode, ok := doc.(*grammar.DocumentNode)
	if !ok {
		t.Fatal("Expected DocumentNode")
	}

	if len(docNode.Features) != 1 {
		t.Errorf("Expected 1 feature, got %d", len(docNode.Features))
	}

	if docNode.Features[0].Name != "AppointmentBooking" {
		t.Errorf("Expected feature name 'AppointmentBooking', got '%s'", docNode.Features[0].Name)
	}
}

func TestWDLParser(t *testing.T) {
	source := `workflow DevelopmentPipeline
    pipeline MainPipeline
        stage Discovery
            processor DiscoveryProcessor`

	wdlGrammar := grammar.NewWDLGrammar()
	parser := NewParser(wdlGrammar, source, "test.wdl")

	doc, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	docNode, ok := doc.(*grammar.WorkflowDocumentNode)
	if !ok {
		t.Fatal("Expected WorkflowDocumentNode")
	}

	if docNode.WorkflowName != "DevelopmentPipeline" {
		t.Errorf("Expected workflow name 'DevelopmentPipeline', got '%s'", docNode.WorkflowName)
	}

	if len(docNode.Pipelines) != 1 {
		t.Errorf("Expected 1 pipeline, got %d", len(docNode.Pipelines))
	}
}

func TestWDLParserWithTriggers(t *testing.T) {
	source := `workflow MainWorkflow
    on ProductUpdated
        pipeline MainPipeline`

	wdlGrammar := grammar.NewWDLGrammar()
	parser := NewParser(wdlGrammar, source, "test.wdl")

	doc, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	docNode, ok := doc.(*grammar.WorkflowDocumentNode)
	if !ok {
		t.Fatal("Expected WorkflowDocumentNode")
	}

	if len(docNode.Triggers) != 1 {
		t.Errorf("Expected 1 trigger, got %d", len(docNode.Triggers))
	}

	if docNode.Triggers[0].EventName != "ProductUpdated" {
		t.Errorf("Expected trigger event 'ProductUpdated', got '%s'", docNode.Triggers[0].EventName)
	}
}

func TestPMFParser(t *testing.T) {
	source := `processor_manifest CodeGenerator
    version 1.0
    author John Doe
    capability code_generation`

	pmfGrammar := grammar.NewPMFGrammar()
	parser := NewParser(pmfGrammar, source, "test.pmf")

	doc, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	docNode, ok := doc.(*grammar.ProcessorManifestNode)
	if !ok {
		t.Fatal("Expected ProcessorManifestNode")
	}

	if docNode.Name != "CodeGenerator" {
		t.Errorf("Expected manifest name 'CodeGenerator', got '%s'", docNode.Name)
	}

	if docNode.Version == nil {
		t.Error("Expected Version node")
	}

	if docNode.Version.Value != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", docNode.Version.Value)
	}

	if len(docNode.Capabilities) != 1 {
		t.Errorf("Expected 1 capability, got %d", len(docNode.Capabilities))
	}
}
