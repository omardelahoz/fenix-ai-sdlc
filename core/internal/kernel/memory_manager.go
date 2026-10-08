package kernel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// MemoryManagerImpl implements the MemoryManager interface with L0 cache.
type MemoryManagerImpl struct {
	mu            sync.RWMutex
	snapshots     map[string]*contracts.ExecutionSnapshot
	astCache      map[string]string // L0 cache for AST nodes
	metadataCache map[string]*contracts.PackageMetadata
	hits          int
	misses        int
	lastCleanup   time.Time
	config        *MemoryManagerConfig
}

// MemoryManagerConfig contains configuration for the Memory Manager.
type MemoryManagerConfig struct {
	MaxSnapshots        int           // Maximum number of snapshots to cache
	CacheTTL            time.Duration // Time-to-live for cached items
	AutoCleanupInterval time.Duration // Interval for automatic cleanup
}

// DefaultMemoryManagerConfig returns default configuration.
func DefaultMemoryManagerConfig() *MemoryManagerConfig {
	return &MemoryManagerConfig{
		MaxSnapshots:        1000,
		CacheTTL:            1 * time.Hour,
		AutoCleanupInterval: 5 * time.Minute,
	}
}

// NewMemoryManager creates a new MemoryManager with default config.
func NewMemoryManager() *MemoryManagerImpl {
	return NewMemoryManagerWithConfig(DefaultMemoryManagerConfig())
}

// NewMemoryManagerWithConfig creates a new MemoryManager with custom config.
func NewMemoryManagerWithConfig(config *MemoryManagerConfig) *MemoryManagerImpl {
	return &MemoryManagerImpl{
		snapshots:     make(map[string]*contracts.ExecutionSnapshot),
		astCache:      make(map[string]string),
		metadataCache: make(map[string]*contracts.PackageMetadata),
		lastCleanup:   time.Now(),
		config:        config,
	}
}

// Resolve resolves memory references in an ExecutionSnapshot.
// This is L0 cache - the fastest cache level for AST nodes.
func (m *MemoryManagerImpl) Resolve(ctx context.Context, snapshot *contracts.ExecutionSnapshot) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resolved := make([]string, 0, len(snapshot.MemoryReferences))

	for _, ref := range snapshot.MemoryReferences {
		// Check L0 cache first
		if astNode, exists := m.astCache[ref]; exists {
			resolved = append(resolved, astNode)
			m.hits++
		} else {
			// In a real implementation, this would load from disk or compute
			// For now, we'll cache the reference itself
			m.astCache[ref] = ref
			resolved = append(resolved, ref)
			m.misses++
		}
	}

	return resolved, nil
}

// Cache caches an ExecutionSnapshot for fast lookup.
func (m *MemoryManagerImpl) Cache(ctx context.Context, snapshotID string, snapshot *contracts.ExecutionSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we need to evict (LRU-like behavior)
	if len(m.snapshots) >= m.config.MaxSnapshots {
		m.evictOldest()
	}

	m.snapshots[snapshotID] = snapshot
	return nil
}

// CacheMetadata caches package metadata separately for faster access.
func (m *MemoryManagerImpl) CacheMetadata(snapshotID string, metadata *contracts.PackageMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metadataCache[snapshotID] = metadata
	return nil
}

// GetMetadata retrieves cached metadata.
func (m *MemoryManagerImpl) GetMetadata(snapshotID string) (*contracts.PackageMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metadata, exists := m.metadataCache[snapshotID]
	if !exists {
		return nil, fmt.Errorf("metadata not found for snapshot: %s", snapshotID)
	}

	return metadata, nil
}

// Invalidate invalidates cached snapshots and related metadata.
func (m *MemoryManagerImpl) Invalidate(ctx context.Context, snapshotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.snapshots[snapshotID]; !exists {
		return fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	delete(m.snapshots, snapshotID)
	delete(m.metadataCache, snapshotID)

	// Also invalidate AST cache entries that reference this snapshot
	for key := range m.astCache {
		// Simple invalidation - in production would be smarter
		if shouldInvalidate(key, snapshotID) {
			delete(m.astCache, key)
		}
	}

	return nil
}

// InvalidateASTNode invalidates a specific AST node from cache.
func (m *MemoryManagerImpl) InvalidateASTNode(nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.astCache, nodeID)
	return nil
}

// Get retrieves a cached snapshot.
func (m *MemoryManagerImpl) Get(ctx context.Context, snapshotID string) (*contracts.ExecutionSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot, exists := m.snapshots[snapshotID]
	if !exists {
		m.misses++
		return nil, fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	m.hits++
	return snapshot, nil
}

// Cleanup removes expired snapshots from cache.
func (m *MemoryManagerImpl) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expired := make([]string, 0)

	// Check snapshots for expiration
	for id, snapshot := range m.snapshots {
		// In a real implementation, snapshots would have creation timestamps
		// For now, we'll skip TTL-based expiration
		_ = snapshot
		_ = now
	}

	// Remove expired
	for _, id := range expired {
		delete(m.snapshots, id)
		delete(m.metadataCache, id)
	}

	m.lastCleanup = time.Now()
	return nil
}

// Stats returns cache statistics.
func (m *MemoryManagerImpl) Stats() MemoryManagerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MemoryManagerStats{
		SnapshotCount: len(m.snapshots),
		ASTCacheCount: len(m.astCache),
		Hits:          m.hits,
		Misses:        m.misses,
		HitRate:       calculateHitRate(m.hits, m.misses),
		LastCleanup:   m.lastCleanup,
	}
}

// MemoryManagerStats contains statistics about the Memory Manager.
type MemoryManagerStats struct {
	SnapshotCount int
	ASTCacheCount int
	Hits          int
	Misses        int
	HitRate       float64
	LastCleanup   time.Time
}

// evictOldest removes the oldest snapshot (simplified LRU).
func (m *MemoryManagerImpl) evictOldest() {
	// In a real implementation, this would track access times
	// For now, we'll just remove a random snapshot
	for id := range m.snapshots {
		delete(m.snapshots, id)
		delete(m.metadataCache, id)
		break
	}
}

// shouldInvalidate checks if an AST node should be invalidated.
func shouldInvalidate(nodeID, snapshotID string) bool {
	// In production, this would check if the node belongs to the snapshot
	// For now, always return false
	return false
}

// calculateHitRate calculates the cache hit rate.
func calculateHitRate(hits, misses int) float64 {
	total := hits + misses
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total) * 100
}
