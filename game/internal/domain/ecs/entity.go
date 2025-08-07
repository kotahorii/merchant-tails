// Package ecs provides an Entity-Component-System architecture
package ecs

import (
	"sync"
	"sync/atomic"
)

// EntityID is a unique identifier for an entity
type EntityID uint64

// ComponentType identifies the type of a component
type ComponentType int

// Component is the interface that all components must implement
type Component interface {
	Type() ComponentType
}

// Entity represents a game object with components
type Entity struct {
	id         EntityID
	components map[ComponentType]Component
	alive      bool
	mu         sync.RWMutex
}

// Global entity ID counter
var entityIDCounter uint64

// NewEntity creates a new entity with a unique ID
func NewEntity() *Entity {
	return &Entity{
		id:         EntityID(atomic.AddUint64(&entityIDCounter, 1)),
		components: make(map[ComponentType]Component),
		alive:      true,
	}
}

// ID returns the entity's unique identifier
func (e *Entity) ID() EntityID {
	return e.id
}

// IsAlive returns whether the entity is still active
func (e *Entity) IsAlive() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.alive
}

// AddComponent adds a component to the entity
func (e *Entity) AddComponent(component Component) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.alive {
		e.components[component.Type()] = component
	}
}

// GetComponent retrieves a component by type
func (e *Entity) GetComponent(componentType ComponentType) Component {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.components[componentType]
}

// HasComponent checks if the entity has a component of the given type
func (e *Entity) HasComponent(componentType ComponentType) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, exists := e.components[componentType]
	return exists
}

// RemoveComponent removes a component from the entity
func (e *Entity) RemoveComponent(componentType ComponentType) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.components, componentType)
}

// GetComponentTypes returns all component types attached to this entity
func (e *Entity) GetComponentTypes() []ComponentType {
	e.mu.RLock()
	defer e.mu.RUnlock()

	types := make([]ComponentType, 0, len(e.components))
	for t := range e.components {
		types = append(types, t)
	}
	return types
}

// Destroy marks the entity as dead and removes all components
func (e *Entity) Destroy() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.alive = false
	e.components = make(map[ComponentType]Component)
}

// EntityManager manages all entities in the game
type EntityManager struct {
	entities map[EntityID]*Entity
	mu       sync.RWMutex
}

// NewEntityManager creates a new entity manager
func NewEntityManager() *EntityManager {
	return &EntityManager{
		entities: make(map[EntityID]*Entity),
	}
}

// CreateEntity creates a new entity and adds it to the manager
func (em *EntityManager) CreateEntity() *Entity {
	entity := NewEntity()

	em.mu.Lock()
	defer em.mu.Unlock()

	em.entities[entity.ID()] = entity
	return entity
}

// GetEntity retrieves an entity by ID
func (em *EntityManager) GetEntity(id EntityID) *Entity {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return em.entities[id]
}

// DestroyEntity removes an entity from the manager
func (em *EntityManager) DestroyEntity(id EntityID) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if entity, exists := em.entities[id]; exists {
		entity.Destroy()
		delete(em.entities, id)
	}
}

// GetAllEntities returns all entities
func (em *EntityManager) GetAllEntities() []*Entity {
	em.mu.RLock()
	defer em.mu.RUnlock()

	entities := make([]*Entity, 0, len(em.entities))
	for _, entity := range em.entities {
		if entity.IsAlive() {
			entities = append(entities, entity)
		}
	}
	return entities
}

// QueryEntities returns all entities that have the specified component type
func (em *EntityManager) QueryEntities(componentType ComponentType) []*Entity {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var result []*Entity
	for _, entity := range em.entities {
		if entity.IsAlive() && entity.HasComponent(componentType) {
			result = append(result, entity)
		}
	}
	return result
}

// QueryEntitiesMultiple returns all entities that have ALL of the specified component types
func (em *EntityManager) QueryEntitiesMultiple(componentTypes ...ComponentType) []*Entity {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var result []*Entity
	for _, entity := range em.entities {
		if !entity.IsAlive() {
			continue
		}

		hasAll := true
		for _, ct := range componentTypes {
			if !entity.HasComponent(ct) {
				hasAll = false
				break
			}
		}

		if hasAll {
			result = append(result, entity)
		}
	}
	return result
}

// Clear removes all entities
func (em *EntityManager) Clear() {
	em.mu.Lock()
	defer em.mu.Unlock()

	for id, entity := range em.entities {
		entity.Destroy()
		delete(em.entities, id)
	}
}
