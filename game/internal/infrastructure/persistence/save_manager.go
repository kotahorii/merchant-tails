package persistence

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/domain/gamestate"
	"github.com/yourusername/merchant-tails/game/internal/domain/inventory"
	"github.com/yourusername/merchant-tails/game/internal/domain/market"
	"github.com/yourusername/merchant-tails/game/internal/domain/progression"
)

// SaveManager handles game save/load operations
type SaveManager struct {
	saveDir string
}

// NewSaveManager creates a new save manager
func NewSaveManager() (*SaveManager, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create save directory
	saveDir := filepath.Join(homeDir, ".merchant-tails", "saves")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, err
	}

	return &SaveManager{
		saveDir: saveDir,
	}, nil
}

// SaveGame saves the current game state to a slot
func (sm *SaveManager) SaveGame(
	slot int,
	state *gamestate.GameState,
	marketData *market.Market,
	inv *inventory.InventoryManager,
	prog *progression.ProgressionManager,
) error {
	// Create save data structure
	saveData := map[string]interface{}{
		"player": map[string]interface{}{
			"name":       "Player", // TODO: Add GetPlayerName to GameState
			"gold":       state.GetGold(),
			"rank":       "Apprentice", // TODO: Add GetRank to GameState
			"reputation": state.GetReputation(),
		},
		"game": map[string]interface{}{
			"currentDay":    1,        // TODO: Add GetCurrentDay to GameState
			"currentSeason": "SPRING", // TODO: Get from time manager
			"saveTimestamp": time.Now().Unix(),
			"saveVersion":   "1.0.0",
		},
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	// Write to file
	filename := sm.getSaveFilename(slot)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	// Also write metadata JSON for quick access
	metadata := SaveMetadata{
		Slot:       slot,
		Timestamp:  time.Now(),
		PlayerName: "Player", // TODO: Add GetPlayerName to GameState
		Gold:       state.GetGold(),
		Day:        1,            // TODO: Add GetCurrentDay to GameState
		Rank:       "Apprentice", // TODO: Add GetRank to GameState
	}

	metadataFile := sm.getMetadataFilename(slot)
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataFile, metadataJSON, 0644)
}

// LoadGame loads a saved game from a slot
func (sm *SaveManager) LoadGame(slot int) (map[string]interface{}, error) {
	filename := sm.getSaveFilename(slot)

	// Read save file
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("save slot %d is empty", slot)
		}
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	// Deserialize JSON
	var saveData map[string]interface{}
	if err := json.Unmarshal(data, &saveData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal save data: %w", err)
	}

	return saveData, nil
}

// GetSaveSlots returns information about all save slots
func (sm *SaveManager) GetSaveSlots() ([]SaveSlotInfo, error) {
	slots := make([]SaveSlotInfo, 3) // Support 3 save slots

	for i := 0; i < 3; i++ {
		info := SaveSlotInfo{
			Slot:   i,
			Exists: false,
		}

		// Check if metadata exists
		metadataFile := sm.getMetadataFilename(i)
		if data, err := os.ReadFile(metadataFile); err == nil {
			var metadata SaveMetadata
			if err := json.Unmarshal(data, &metadata); err == nil {
				info.Exists = true
				info.Metadata = &metadata
			}
		}

		slots[i] = info
	}

	return slots, nil
}

// DeleteSave deletes a save file
func (sm *SaveManager) DeleteSave(slot int) error {
	saveFile := sm.getSaveFilename(slot)
	metadataFile := sm.getMetadataFilename(slot)

	// Remove both files
	_ = os.Remove(saveFile)
	_ = os.Remove(metadataFile)

	return nil
}

// ExportSave exports a save to a writer
func (sm *SaveManager) ExportSave(slot int, w io.Writer) error {
	filename := sm.getSaveFilename(slot)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// ImportSave imports a save from a reader
func (sm *SaveManager) ImportSave(slot int, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Validate it's a valid save
	var saveData map[string]interface{}
	if err := json.Unmarshal(data, &saveData); err != nil {
		return fmt.Errorf("invalid save data: %w", err)
	}

	// Write to slot
	filename := sm.getSaveFilename(slot)
	return os.WriteFile(filename, data, 0644)
}

// Helper methods

func (sm *SaveManager) getSaveFilename(slot int) string {
	return filepath.Join(sm.saveDir, fmt.Sprintf("save_%d.dat", slot))
}

func (sm *SaveManager) getMetadataFilename(slot int) string {
	return filepath.Join(sm.saveDir, fmt.Sprintf("save_%d.json", slot))
}

// Helper method to get save directory path
func (sm *SaveManager) GetSaveDirectory() string {
	return sm.saveDir
}

// SaveMetadata contains quick-access save information
type SaveMetadata struct {
	Slot       int       `json:"slot"`
	Timestamp  time.Time `json:"timestamp"`
	PlayerName string    `json:"playerName"`
	Gold       int       `json:"gold"`
	Day        int       `json:"day"`
	Rank       string    `json:"rank"`
}

// SaveSlotInfo contains information about a save slot
type SaveSlotInfo struct {
	Slot     int           `json:"slot"`
	Exists   bool          `json:"exists"`
	Metadata *SaveMetadata `json:"metadata,omitempty"`
}
