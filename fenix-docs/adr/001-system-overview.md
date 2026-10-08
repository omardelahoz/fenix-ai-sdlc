# ADR 001: System Overview

## Status
Proposed

## Context
We need an AI SDLC Engine that is model-agnostic, uses declarative configs, and has a persistent product memory.

## Decision
We will build a Go-based engine (fenix-engine) and an installer, and keep all definitions declarative (.nd using YAML structure).
