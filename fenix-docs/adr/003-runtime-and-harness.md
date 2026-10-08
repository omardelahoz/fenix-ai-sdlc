# ADR 003: Runtime and Harness

## Status
Proposed

## Context
We need to orchestrate multiple agents in parallel.

## Decision
Use a Go-based Harness that manages goroutines for parallel agent execution, tracks state in a state manager, and manages the lifecycle using pipeline definitions.
