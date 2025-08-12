package rendering

import (
	"runtime"
	"sync"
	"time"
)

// RenderOptimizer manages rendering performance optimizations
type RenderOptimizer struct {
	mu              sync.RWMutex
	batchSize       int
	cullingEnabled  bool
	lodEnabled      bool
	frameTarget     int
	currentFPS      float64
	frameTime       time.Duration
	adaptiveQuality bool
	qualityLevel    QualityLevel
	stats           *RenderStats
}

// QualityLevel represents rendering quality settings
type QualityLevel int

const (
	QualityLow QualityLevel = iota
	QualityMedium
	QualityHigh
	QualityUltra
)

// RenderStats tracks rendering performance metrics
type RenderStats struct {
	FrameCount       uint64
	TotalFrameTime   time.Duration
	AverageFrameTime time.Duration
	MinFrameTime     time.Duration
	MaxFrameTime     time.Duration
	DroppedFrames    uint64
	BatchCount       uint64
	DrawCalls        uint64
	VerticesRendered uint64
	TexturesLoaded   uint64
	LastUpdate       time.Time
}

// NewRenderOptimizer creates a new render optimizer
func NewRenderOptimizer() *RenderOptimizer {
	return &RenderOptimizer{
		batchSize:       100,
		cullingEnabled:  true,
		lodEnabled:      true,
		frameTarget:     60,
		adaptiveQuality: true,
		qualityLevel:    QualityHigh,
		stats:           &RenderStats{},
	}
}

// BatchRenderer optimizes rendering by batching draw calls
type BatchRenderer struct {
	mu           sync.Mutex
	batches      map[string]*RenderBatch
	maxBatchSize int
	optimizer    *RenderOptimizer
}

// RenderBatch represents a batch of render operations
type RenderBatch struct {
	ID         string
	Objects    []RenderObject
	Shader     string
	Texture    string
	Transform  Transform
	Dirty      bool
	LastUpdate time.Time
}

// RenderObject represents an object to be rendered
type RenderObject struct {
	ID          string
	Vertices    []Vertex
	Indices     []uint32
	Transform   Transform
	Material    Material
	Visible     bool
	LODLevel    int
	BoundingBox BoundingBox
}

// Vertex represents a 3D vertex
type Vertex struct {
	Position [3]float32
	Normal   [3]float32
	TexCoord [2]float32
	Color    [4]float32
}

// Transform represents object transformation
type Transform struct {
	Position [3]float32
	Rotation [4]float32 // Quaternion
	Scale    [3]float32
}

// Material represents rendering material
type Material struct {
	DiffuseTexture  string
	NormalTexture   string
	SpecularTexture string
	Shininess       float32
	Opacity         float32
}

// BoundingBox for frustum culling
type BoundingBox struct {
	Min [3]float32
	Max [3]float32
}

// NewBatchRenderer creates a new batch renderer
func NewBatchRenderer(optimizer *RenderOptimizer) *BatchRenderer {
	return &BatchRenderer{
		batches:      make(map[string]*RenderBatch),
		maxBatchSize: optimizer.batchSize,
		optimizer:    optimizer,
	}
}

// AddObject adds an object to the appropriate batch
func (br *BatchRenderer) AddObject(obj RenderObject) {
	br.mu.Lock()
	defer br.mu.Unlock()

	// Create batch key based on shader and texture
	batchKey := obj.Material.DiffuseTexture + "_" + obj.Material.NormalTexture

	batch, exists := br.batches[batchKey]
	if !exists {
		batch = &RenderBatch{
			ID:         batchKey,
			Objects:    make([]RenderObject, 0, br.maxBatchSize),
			Texture:    obj.Material.DiffuseTexture,
			LastUpdate: time.Now(),
		}
		br.batches[batchKey] = batch
	}

	// Add object to batch
	batch.Objects = append(batch.Objects, obj)
	batch.Dirty = true

	// Flush if batch is full
	if len(batch.Objects) >= br.maxBatchSize {
		br.flushBatch(batch)
	}
}

// flushBatch renders and clears a batch
func (br *BatchRenderer) flushBatch(batch *RenderBatch) {
	if len(batch.Objects) == 0 {
		return
	}

	// Update stats
	br.optimizer.stats.BatchCount++
	br.optimizer.stats.DrawCalls++

	// Clear batch
	batch.Objects = batch.Objects[:0]
	batch.Dirty = false
}

// FlushAll renders all pending batches
func (br *BatchRenderer) FlushAll() {
	br.mu.Lock()
	defer br.mu.Unlock()

	for _, batch := range br.batches {
		if batch.Dirty {
			br.flushBatch(batch)
		}
	}
}

// FrustumCuller performs view frustum culling
type FrustumCuller struct {
	mu            sync.RWMutex
	viewMatrix    [16]float32
	projMatrix    [16]float32
	frustumPlanes [6]Plane
	enabled       bool
}

// Plane represents a frustum plane
type Plane struct {
	Normal   [3]float32
	Distance float32
}

// NewFrustumCuller creates a new frustum culler
func NewFrustumCuller() *FrustumCuller {
	return &FrustumCuller{
		enabled: true,
	}
}

// UpdateMatrices updates view and projection matrices
func (fc *FrustumCuller) UpdateMatrices(view, proj [16]float32) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.viewMatrix = view
	fc.projMatrix = proj
	fc.extractFrustumPlanes()
}

// extractFrustumPlanes extracts frustum planes from matrices
func (fc *FrustumCuller) extractFrustumPlanes() {
	// Combine view and projection matrices
	var vp [16]float32
	multiplyMatrices(&vp, &fc.viewMatrix, &fc.projMatrix)

	// Extract planes (simplified)
	// Left plane
	fc.frustumPlanes[0] = Plane{
		Normal:   [3]float32{vp[3] + vp[0], vp[7] + vp[4], vp[11] + vp[8]},
		Distance: vp[15] + vp[12],
	}
	// Right plane
	fc.frustumPlanes[1] = Plane{
		Normal:   [3]float32{vp[3] - vp[0], vp[7] - vp[4], vp[11] - vp[8]},
		Distance: vp[15] - vp[12],
	}
	// Bottom plane
	fc.frustumPlanes[2] = Plane{
		Normal:   [3]float32{vp[3] + vp[1], vp[7] + vp[5], vp[11] + vp[9]},
		Distance: vp[15] + vp[13],
	}
	// Top plane
	fc.frustumPlanes[3] = Plane{
		Normal:   [3]float32{vp[3] - vp[1], vp[7] - vp[5], vp[11] - vp[9]},
		Distance: vp[15] - vp[13],
	}
	// Near plane
	fc.frustumPlanes[4] = Plane{
		Normal:   [3]float32{vp[3] + vp[2], vp[7] + vp[6], vp[11] + vp[10]},
		Distance: vp[15] + vp[14],
	}
	// Far plane
	fc.frustumPlanes[5] = Plane{
		Normal:   [3]float32{vp[3] - vp[2], vp[7] - vp[6], vp[11] - vp[10]},
		Distance: vp[15] - vp[14],
	}
}

// IsVisible checks if a bounding box is visible
func (fc *FrustumCuller) IsVisible(box BoundingBox) bool {
	if !fc.enabled {
		return true
	}

	fc.mu.RLock()
	defer fc.mu.RUnlock()

	// Check each frustum plane
	for _, plane := range fc.frustumPlanes {
		// Get the positive vertex relative to the plane normal
		var pVertex [3]float32
		for i := 0; i < 3; i++ {
			if plane.Normal[i] >= 0 {
				pVertex[i] = box.Max[i]
			} else {
				pVertex[i] = box.Min[i]
			}
		}

		// Check if positive vertex is outside
		dot := plane.Normal[0]*pVertex[0] +
			plane.Normal[1]*pVertex[1] +
			plane.Normal[2]*pVertex[2]
		if dot+plane.Distance < 0 {
			return false // Outside frustum
		}
	}

	return true // Inside or intersecting frustum
}

// LODManager manages level of detail
type LODManager struct {
	mu        sync.RWMutex
	lodLevels []float32 // Distance thresholds
	cameraPos [3]float32
	enabled   bool
}

// NewLODManager creates a new LOD manager
func NewLODManager() *LODManager {
	return &LODManager{
		lodLevels: []float32{10.0, 25.0, 50.0, 100.0}, // Distance thresholds
		enabled:   true,
	}
}

// UpdateCameraPosition updates the camera position for LOD calculations
func (lm *LODManager) UpdateCameraPosition(pos [3]float32) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.cameraPos = pos
}

// GetLODLevel calculates the appropriate LOD level for an object
func (lm *LODManager) GetLODLevel(objPos [3]float32) int {
	if !lm.enabled {
		return 0 // Highest detail
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Calculate distance to camera
	dx := objPos[0] - lm.cameraPos[0]
	dy := objPos[1] - lm.cameraPos[1]
	dz := objPos[2] - lm.cameraPos[2]
	distance := float32(sqrt(float64(dx*dx + dy*dy + dz*dz)))

	// Determine LOD level
	for i, threshold := range lm.lodLevels {
		if distance < threshold {
			return i
		}
	}

	return len(lm.lodLevels) // Lowest detail
}

// AdaptiveQualityManager dynamically adjusts quality based on performance
type AdaptiveQualityManager struct {
	mu              sync.RWMutex
	targetFPS       float64
	currentFPS      float64
	qualityLevel    QualityLevel
	adjustmentDelay time.Duration
	lastAdjustment  time.Time
	enabled         bool
}

// NewAdaptiveQualityManager creates a new adaptive quality manager
func NewAdaptiveQualityManager(targetFPS float64) *AdaptiveQualityManager {
	return &AdaptiveQualityManager{
		targetFPS:       targetFPS,
		qualityLevel:    QualityHigh,
		adjustmentDelay: 2 * time.Second,
		enabled:         true,
	}
}

// UpdateFPS updates the current FPS and adjusts quality if needed
func (aqm *AdaptiveQualityManager) UpdateFPS(fps float64) {
	aqm.mu.Lock()
	defer aqm.mu.Unlock()

	aqm.currentFPS = fps

	if !aqm.enabled {
		return
	}

	// Check if enough time has passed since last adjustment
	if time.Since(aqm.lastAdjustment) < aqm.adjustmentDelay {
		return
	}

	// Adjust quality based on FPS
	const tolerance = 5.0
	if fps < aqm.targetFPS-tolerance && aqm.qualityLevel > QualityLow {
		// Decrease quality
		aqm.qualityLevel--
		aqm.lastAdjustment = time.Now()
	} else if fps > aqm.targetFPS+tolerance && aqm.qualityLevel < QualityUltra {
		// Increase quality
		aqm.qualityLevel++
		aqm.lastAdjustment = time.Now()
	}
}

// GetQualitySettings returns current quality settings
func (aqm *AdaptiveQualityManager) GetQualitySettings() QualitySettings {
	aqm.mu.RLock()
	defer aqm.mu.RUnlock()

	switch aqm.qualityLevel {
	case QualityLow:
		return QualitySettings{
			TextureResolution: 0.5,
			ShadowQuality:     0,
			AntiAliasing:      0,
			PostProcessing:    false,
			ParticleCount:     0.25,
		}
	case QualityMedium:
		return QualitySettings{
			TextureResolution: 0.75,
			ShadowQuality:     1,
			AntiAliasing:      2,
			PostProcessing:    false,
			ParticleCount:     0.5,
		}
	case QualityHigh:
		return QualitySettings{
			TextureResolution: 1.0,
			ShadowQuality:     2,
			AntiAliasing:      4,
			PostProcessing:    true,
			ParticleCount:     0.75,
		}
	case QualityUltra:
		return QualitySettings{
			TextureResolution: 1.0,
			ShadowQuality:     3,
			AntiAliasing:      8,
			PostProcessing:    true,
			ParticleCount:     1.0,
		}
	default:
		return QualitySettings{}
	}
}

// QualitySettings represents rendering quality parameters
type QualitySettings struct {
	TextureResolution float32
	ShadowQuality     int
	AntiAliasing      int
	PostProcessing    bool
	ParticleCount     float32
}

// Helper functions
func multiplyMatrices(result, a, b *[16]float32) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i*4+j] = 0
			for k := 0; k < 4; k++ {
				result[i*4+j] += a[i*4+k] * b[k*4+j]
			}
		}
	}
}

func sqrt(x float64) float64 {
	// Simple square root approximation
	if x == 0 {
		return 0
	}
	// Newton's method for square root
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// UpdateStats updates rendering statistics
func (ro *RenderOptimizer) UpdateStats(frameTime time.Duration) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	ro.stats.FrameCount++
	ro.stats.TotalFrameTime += frameTime
	ro.stats.AverageFrameTime = ro.stats.TotalFrameTime / time.Duration(ro.stats.FrameCount)

	if ro.stats.MinFrameTime == 0 || frameTime < ro.stats.MinFrameTime {
		ro.stats.MinFrameTime = frameTime
	}
	if frameTime > ro.stats.MaxFrameTime {
		ro.stats.MaxFrameTime = frameTime
	}

	// Calculate FPS
	if frameTime > 0 {
		ro.currentFPS = float64(time.Second) / float64(frameTime)
	}

	// Check for dropped frames
	targetFrameTime := time.Second / time.Duration(ro.frameTarget)
	if frameTime > targetFrameTime*2 {
		ro.stats.DroppedFrames++
	}

	ro.stats.LastUpdate = time.Now()
}

// GetStats returns current rendering statistics
func (ro *RenderOptimizer) GetStats() RenderStats {
	ro.mu.RLock()
	defer ro.mu.RUnlock()
	return *ro.stats
}

// SetQualityLevel manually sets the quality level
func (ro *RenderOptimizer) SetQualityLevel(level QualityLevel) {
	ro.mu.Lock()
	defer ro.mu.Unlock()
	ro.qualityLevel = level
}

// EnableFeature enables a specific optimization feature
func (ro *RenderOptimizer) EnableFeature(feature string, enabled bool) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	switch feature {
	case "culling":
		ro.cullingEnabled = enabled
	case "lod":
		ro.lodEnabled = enabled
	case "adaptive":
		ro.adaptiveQuality = enabled
	}
}

// Optimize runs all optimization passes
func (ro *RenderOptimizer) Optimize() {
	// Force GC to free unused GPU resources
	runtime.GC()

	// Additional optimization logic can be added here
}
