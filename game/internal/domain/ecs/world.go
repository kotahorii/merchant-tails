package ecs

import (
	"context"
	"sync"
	"time"
)

// World contains all ECS data and systems
type World struct {
	entityManager *EntityManager
	systemManager *SystemManager
	resources     map[string]interface{}
	mu            sync.RWMutex
}

// NewWorld creates a new ECS world
func NewWorld() *World {
	return &World{
		entityManager: NewEntityManager(),
		systemManager: NewSystemManager(),
		resources:     make(map[string]interface{}),
	}
}

// GetEntityManager returns the entity manager
func (w *World) GetEntityManager() *EntityManager {
	return w.entityManager
}

// GetSystemManager returns the system manager
func (w *World) GetSystemManager() *SystemManager {
	return w.systemManager
}

// AddResource adds a shared resource to the world
func (w *World) AddResource(name string, resource interface{}) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.resources[name] = resource
}

// GetResource retrieves a shared resource
func (w *World) GetResource(name string) interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.resources[name]
}

// RemoveResource removes a shared resource
func (w *World) RemoveResource(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.resources, name)
}

// Update runs all systems
func (w *World) Update(ctx context.Context, deltaTime time.Duration) error {
	return w.systemManager.Update(ctx, w, deltaTime)
}

// Clear removes all entities and resources
func (w *World) Clear() {
	w.entityManager.Clear()

	w.mu.Lock()
	defer w.mu.Unlock()
	w.resources = make(map[string]interface{})
}

// CreateEntity is a convenience method to create an entity
func (w *World) CreateEntity() *Entity {
	return w.entityManager.CreateEntity()
}

// DestroyEntity is a convenience method to destroy an entity
func (w *World) DestroyEntity(id EntityID) {
	w.entityManager.DestroyEntity(id)
}

// QueryEntities is a convenience method to query entities
func (w *World) QueryEntities(componentType ComponentType) []*Entity {
	return w.entityManager.QueryEntities(componentType)
}

// QueryEntitiesMultiple is a convenience method to query entities with multiple components
func (w *World) QueryEntitiesMultiple(componentTypes ...ComponentType) []*Entity {
	return w.entityManager.QueryEntitiesMultiple(componentTypes...)
}

// AddSystem is a convenience method to add a system
func (w *World) AddSystem(name string, system System, priority int) {
	w.systemManager.AddSystem(name, system, priority)
}

// RemoveSystem is a convenience method to remove a system
func (w *World) RemoveSystem(name string) {
	w.systemManager.RemoveSystem(name)
}

// GetSystem is a convenience method to get a system
func (w *World) GetSystem(name string) System {
	return w.systemManager.GetSystem(name)
}

// WorldBuilder provides a fluent interface for building worlds
type WorldBuilder struct {
	world   *World
	systems []struct {
		name     string
		system   System
		priority int
	}
}

// NewWorldBuilder creates a new world builder
func NewWorldBuilder() *WorldBuilder {
	return &WorldBuilder{
		world: NewWorld(),
		systems: make([]struct {
			name     string
			system   System
			priority int
		}, 0),
	}
}

// WithSystem adds a system to the world
func (wb *WorldBuilder) WithSystem(name string, system System, priority int) *WorldBuilder {
	wb.systems = append(wb.systems, struct {
		name     string
		system   System
		priority int
	}{name, system, priority})
	return wb
}

// WithResource adds a resource to the world
func (wb *WorldBuilder) WithResource(name string, resource interface{}) *WorldBuilder {
	wb.world.AddResource(name, resource)
	return wb
}

// Build creates the world with all configured systems and resources
func (wb *WorldBuilder) Build() *World {
	for _, sys := range wb.systems {
		wb.world.AddSystem(sys.name, sys.system, sys.priority)
	}
	return wb.world
}

// CreateDefaultWorld creates a world with standard game systems
func CreateDefaultWorld() *World {
	return NewWorldBuilder().
		WithSystem("physics", &PhysicsSystem{}, 1).
		WithSystem("movement", &MovementSystem{}, 2).
		WithSystem("ai", &AISystem{}, 3).
		WithSystem("inventory", &InventorySystem{}, 4).
		WithSystem("render", &RenderSystem{}, 99).
		Build()
}
