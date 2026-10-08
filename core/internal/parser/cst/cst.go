package cst

import (
	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/lexer"
)

// GreenNode represents an immutable node in the Green Tree.
// Green nodes are purely immutable, hashable, and share memory across identical subtrees.
// They contain no parent pointers or absolute spans.
type GreenNode struct {
	Kind     NodeKind
	Children []GreenNode
	Trivia   []Trivia
	Value    string // For leaf nodes (identifiers, literals, etc.)
}

// NodeKind represents the kind of a CST node.
type NodeKind int

const (
	NodeKindNone NodeKind = iota
	NodeKindDocument
	NodeKindProduct
	NodeKindVision
	NodeKindUsers
	NodeKindConstraints
	NodeKindFeature
	NodeKindStory
	NodeKindRequirement
	NodeKindAcceptance
	NodeKindArchitecture
	NodeKindImplementation
	NodeKindRelease
	NodeKindWorkflow
	NodeKindOn
	NodeKindPipeline
	NodeKindStage
	NodeKindProcessor
	NodeKindPolicy
	NodeKindDepends
	NodeKindFanout
	NodeKindStrategy
	NodeKindGate
	NodeKindToken
)

// RedNode represents an ephemeral wrapper (facade) over a Green node.
// Red nodes provide absolute Span, Parent pointers, and navigation for the IDE.
// They are created on-demand and should not be cached.
type RedNode struct {
	Green   *GreenNode
	Parent  *RedNode
	Span    lexer.Span
	Index   int // Index in parent's children
}

// IsRoot returns true if this RedNode has no parent.
func (n *RedNode) IsRoot() bool {
	return n.Parent == nil
}

// Parent returns the parent RedNode.
func (n *RedNode) Parent() *RedNode {
	return n.Parent
}

// Children returns the children as RedNodes.
func (n *RedNode) Children() []*RedNode {
	children := make([]*RedNode, len(n.Green.Children))
	for i, child := range n.Green.Children {
		children[i] = &RedNode{
			Green:  &child,
			Parent: n,
			Index:  i,
		}
	}
	return children
}

// ChildAt returns the child at the given index as a RedNode.
func (n *RedNode) ChildAt(index int) *RedNode {
	if index < 0 || index >= len(n.Green.Children) {
		return nil
	}
	return &RedNode{
		Green:  &n.Green.Children[index],
		Parent: n,
		Index:  index,
	}
}

// FirstChild returns the first child as a RedNode.
func (n *RedNode) FirstChild() *RedNode {
	if len(n.Green.Children) == 0 {
		return nil
	}
	return n.ChildAt(0)
}

// LastChild returns the last child as a RedNode.
func (n *RedNode) LastChild() *RedNode {
	if len(n.Green.Children) == 0 {
		return nil
	}
	return n.ChildAt(len(n.Green.Children) - 1)
}

// NextSibling returns the next sibling as a RedNode.
func (n *RedNode) NextSibling() *RedNode {
	if n.Parent == nil || n.Index >= len(n.Parent.Green.Children)-1 {
		return nil
	}
	return n.Parent.ChildAt(n.Index + 1)
}

// PreviousSibling returns the previous sibling as a RedNode.
func (n *RedNode) PreviousSibling() *RedNode {
	if n.Parent == nil || n.Index <= 0 {
		return nil
	}
 return n.Parent.ChildAt(n.Index - 1)
}

// Trivia represents trivia (whitespace, comments) in the source.
type Trivia struct {
	Kind  TriviaKind
	Text  string
}

// TriviaKind represents the type of trivia.
type TriviaKind int

const (
	TriviaKindWhitespace TriviaKind = iota
	TriviaKindComment
	TriviaKindNewline
)

// CST represents the Concrete Syntax Tree with Green nodes.
type CST struct {
	Root GreenNode
}

// CreateCST creates a CST from a Green node.
func CreateCST(root GreenNode) *CST {
	return &CST{
		Root: root,
	}
}

// CreateRedNode creates a RedNode from a Green node.
func CreateRedNode(green *GreenNode) *RedNode {
	return &RedNode{
		Green: green,
	}
}

// WalkFunc is a function called for each node during traversal.
type WalkFunc func(node *RedNode) bool

// Walk traverses the CST in pre-order.
func Walk(node *RedNode, fn WalkFunc) {
	if node == nil {
		return
	}

	if !fn(node) {
		return
	}

	for _, child := range node.Children() {
		Walk(child, fn)
	}
}

// WalkPostOrder traverses the CST in post-order.
func WalkPostOrder(node *RedNode, fn WalkFunc) {
	if node == nil {
		return
	}

	for _, child := range node.Children() {
		WalkPostOrder(child, fn)
	}

	fn(node)
}

// Find finds the first node that matches the predicate.
func Find(node *RedNode, predicate func(*RedNode) bool) *RedNode {
	var result *RedNode

	Walk(node, func(n *RedNode) bool {
		if predicate(n) {
			result = n
			return false // Stop walking
		}
		return true
	})

	return result
}

// FindAll finds all nodes that match the predicate.
func FindAll(node *RedNode, predicate func(*RedNode) bool) []*RedNode {
	var results []*RedNode

	Walk(node, func(n *RedNode) bool {
		if predicate(n) {
			results = append(results, n)
		}
		return true
	})

	return results
}

// FindByKind finds all nodes of a specific kind.
func FindByKind(node *RedNode, kind NodeKind) []*RedNode {
	return FindAll(node, func(n *RedNode) bool {
		return n.Green.Kind == kind
	})
}

// StructuralSharing returns true if two Green nodes are structurally equal.
// This is used for deduplication and memory optimization.
func StructuralSharing(a, b *GreenNode) bool {
	if a == b {
		return true
	}

	if a.Kind != b.Kind {
		return false
	}

	if a.Value != b.Value {
		return false
	}

	if len(a.Children) != len(b.Children) {
		return false
	}

	for i := range a.Children {
		if !StructuralSharing(&a.Children[i], &b.Children[i]) {
			return false
		}
	}

	return true
}

// NodePool manages a pool of Green nodes for structural sharing.
type NodePool struct {
	nodes map[string]*GreenNode
}

// NewNodePool creates a new NodePool.
func NewNodePool() *NodePool {
	return &NodePool{
		nodes: make(map[string]*GreenNode),
	}
}

// GetOrCreate gets a node from the pool or creates a new one.
func (p *NodePool) GetOrCreate(node GreenNode) *GreenNode {
	key := p.hashNode(node)
	if existing, exists := p.nodes[key]; exists {
		return existing
	}
	p.nodes[key] = &node
	return &node
}

// hashNode creates a hash key for a Green node.
func (p *NodePool) hashNode(node GreenNode) string {
	// Simple hash implementation
	// In production, this would use a more sophisticated hashing
	return string(node.Kind) + node.Value + string(rune(len(node.Children)))
}
