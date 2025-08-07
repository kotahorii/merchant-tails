package ecs

import (
	"context"
	"sort"
	"sync"
	"time"
)

// System is the interface that all systems must implement
type System interface {
	// Name returns the name of the system
	Name() string

	// Update processes entities with the required components
	Update(ctx context.Context, world *World, deltaTime time.Duration) error

	// GetRequiredComponents returns the component types this system requires
	GetRequiredComponents() []ComponentType
}

// SystemManager manages all systems
type SystemManager struct {
	systems    map[string]System
	priorities map[string]int
	mu         sync.RWMutex
}

// NewSystemManager creates a new system manager
func NewSystemManager() *SystemManager {
	return &SystemManager{
		systems:    make(map[string]System),
		priorities: make(map[string]int),
	}
}

// AddSystem adds a system to the manager with a priority
// Lower priority numbers execute first
func (sm *SystemManager) AddSystem(name string, system System, priority int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.systems[name] = system
	sm.priorities[name] = priority
}

// RemoveSystem removes a system from the manager
func (sm *SystemManager) RemoveSystem(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.systems, name)
	delete(sm.priorities, name)
}

// GetSystem retrieves a system by name
func (sm *SystemManager) GetSystem(name string) System {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.systems[name]
}

// Update runs all systems in priority order
func (sm *SystemManager) Update(ctx context.Context, world *World, deltaTime time.Duration) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create a slice of system names sorted by priority
	type systemPriority struct {
		name     string
		priority int
	}

	sortedSystems := make([]systemPriority, 0, len(sm.systems))
	for name, priority := range sm.priorities {
		sortedSystems = append(sortedSystems, systemPriority{name, priority})
	}

	sort.Slice(sortedSystems, func(i, j int) bool {
		return sortedSystems[i].priority < sortedSystems[j].priority
	})

	// Update systems in order
	for _, sp := range sortedSystems {
		system := sm.systems[sp.name]
		if err := system.Update(ctx, world, deltaTime); err != nil {
			return err
		}
	}

	return nil
}

// MovementSystem handles entity movement
type MovementSystem struct{}

func (ms *MovementSystem) Name() string {
	return "Movement"
}

func (ms *MovementSystem) Update(ctx context.Context, world *World, deltaTime time.Duration) error {
	entities := world.GetEntityManager().QueryEntitiesMultiple(
		ComponentTypeTransform,
		ComponentTypePhysics,
	)

	dt := deltaTime.Seconds()

	for _, entity := range entities {
		transform := entity.GetComponent(ComponentTypeTransform).(*TransformComponent)
		physics := entity.GetComponent(ComponentTypePhysics).(*PhysicsComponent)

		// Update position based on velocity
		transform.X += physics.VelocityX * dt
		transform.Y += physics.VelocityY * dt
	}

	return nil
}

func (ms *MovementSystem) GetRequiredComponents() []ComponentType {
	return []ComponentType{ComponentTypeTransform, ComponentTypePhysics}
}

// PhysicsSystem handles physics simulation
type PhysicsSystem struct{}

func (ps *PhysicsSystem) Name() string {
	return "Physics"
}

func (ps *PhysicsSystem) Update(ctx context.Context, world *World, deltaTime time.Duration) error {
	entities := world.GetEntityManager().QueryEntities(ComponentTypePhysics)

	dt := deltaTime.Seconds()

	for _, entity := range entities {
		physics := entity.GetComponent(ComponentTypePhysics).(*PhysicsComponent)

		// Apply acceleration
		physics.VelocityX += physics.AccelerationX * dt
		physics.VelocityY += physics.AccelerationY * dt

		// Apply friction
		physics.VelocityX -= physics.VelocityX * physics.Friction * dt
		physics.VelocityY -= physics.VelocityY * physics.Friction * dt
	}

	return nil
}

func (ps *PhysicsSystem) GetRequiredComponents() []ComponentType {
	return []ComponentType{ComponentTypePhysics}
}

// RenderSystem handles rendering (placeholder for now)
type RenderSystem struct{}

func (rs *RenderSystem) Name() string {
	return "Render"
}

func (rs *RenderSystem) Update(ctx context.Context, world *World, deltaTime time.Duration) error {
	// Placeholder - actual rendering would interface with Godot
	entities := world.GetEntityManager().QueryEntitiesMultiple(
		ComponentTypeTransform,
		ComponentTypeRender,
	)

	// Sort by layer for proper rendering order
	sort.Slice(entities, func(i, j int) bool {
		renderI := entities[i].GetComponent(ComponentTypeRender).(*RenderComponent)
		renderJ := entities[j].GetComponent(ComponentTypeRender).(*RenderComponent)
		return renderI.Layer < renderJ.Layer
	})

	// In a real implementation, this would send render commands to Godot
	_ = entities

	return nil
}

func (rs *RenderSystem) GetRequiredComponents() []ComponentType {
	return []ComponentType{ComponentTypeTransform, ComponentTypeRender}
}

// InventorySystem manages inventory operations
type InventorySystem struct{}

func (is *InventorySystem) Name() string {
	return "Inventory"
}

func (is *InventorySystem) Update(ctx context.Context, world *World, deltaTime time.Duration) error {
	entities := world.GetEntityManager().QueryEntities(ComponentTypeInventory)

	for _, entity := range entities {
		inventory := entity.GetComponent(ComponentTypeInventory).(*InventoryComponent)

		// Remove references to destroyed items
		validItems := make([]EntityID, 0, len(inventory.Items))
		for _, itemID := range inventory.Items {
			if item := world.GetEntityManager().GetEntity(itemID); item != nil && item.IsAlive() {
				validItems = append(validItems, itemID)
			}
		}
		inventory.Items = validItems
	}

	return nil
}

func (is *InventorySystem) GetRequiredComponents() []ComponentType {
	return []ComponentType{ComponentTypeInventory}
}

// AISystem handles AI behavior
type AISystem struct{}

func (as *AISystem) Name() string {
	return "AI"
}

func (as *AISystem) Update(ctx context.Context, world *World, deltaTime time.Duration) error {
	entities := world.GetEntityManager().QueryEntities(ComponentTypeAI)

	for _, entity := range entities {
		ai := entity.GetComponent(ComponentTypeAI).(*AIComponent)

		// Simple state machine placeholder
		switch ai.State {
		case "idle":
			// Look for targets
			if ai.Target == 0 {
				// Find a target based on behavior type
				as.findTarget(entity, ai, world)
			}
		case "pursuing":
			// Move towards target
			if ai.Target != 0 {
				as.moveTowardsTarget(entity, ai, world)
			}
		case "trading":
			// Execute trade logic
			as.executeTrade(entity, ai, world)
		}
	}

	return nil
}

func (as *AISystem) GetRequiredComponents() []ComponentType {
	return []ComponentType{ComponentTypeAI}
}

func (as *AISystem) findTarget(entity *Entity, ai *AIComponent, world *World) {
	// Placeholder for target finding logic
	// Would search for appropriate targets based on AI behavior type
}

func (as *AISystem) moveTowardsTarget(entity *Entity, ai *AIComponent, world *World) {
	// Placeholder for movement logic
	// Would calculate path and update physics component
}

func (as *AISystem) executeTrade(entity *Entity, ai *AIComponent, world *World) {
	// Placeholder for trade execution
	// Would interact with merchant and trading systems
}
