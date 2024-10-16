package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Asset represents an asset with an ID, Name, and Value
type Asset struct {
    ID    string `json:"id"`
    Value []byte `json:"value"`
}

// SmartContract simulates a smart contract with basic asset management functionality
type SmartContract struct {
    ledger map[string]Asset
    mutex  sync.Mutex
    filepath   string // File path to store ledger
	shutdownCh chan struct{}
}

// NewSmartContract creates a new SmartContract instance and loads ledger from file
func NewSmartContract(filePath string) *SmartContract {
    contract := &SmartContract{
        ledger: make(map[string]Asset),
        filepath:   filePath,
		shutdownCh: make(chan struct{}),
    }
    contract.loadFromFile()
	go contract.persist()
    return contract
}

// Stop gracefully shuts down the SmartContract, ensuring all data is persisted
func (sc *SmartContract) Stop() {
    close(sc.shutdownCh)
}

// persist routine handles periodic persistence in the background
func (sc *SmartContract) persist() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            sc.saveToFile()
        case <-sc.shutdownCh:
            sc.saveToFile()
            return
        }
    }
}

// CreateAsset creates a new asset in the ledger
func (sc *SmartContract) CreateAsset(id string, value []byte) error {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()

    if _, exists := sc.ledger[id]; exists {
        err := fmt.Errorf("asset %s already exists", id)
        log.Error().Err(err).Msg("Failed to create asset")
        return err
    }

    sc.ledger[id] = Asset{
        ID:    id,
        Value: value,
    }

    log.Info().Str("id", id).Msg("Asset created")
    return nil
}

// ReadAsset retrieves an asset from the ledger by ID
func (sc *SmartContract) ReadAsset(id string) (Asset, error) {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()

    asset, exists := sc.ledger[id]
    if !exists {
		err := fmt.Errorf("asset %s does not exist", id)
        log.Error().Err(err).Msg("Failed to read asset")
		return Asset{}, err
    }

    log.Info().Str("id", id).Msg("Asset read")
    return asset, nil
}

// UpdateAsset updates an existing asset in the ledger
func (sc *SmartContract) UpdateAsset(id string, value []byte) error {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()

    if _, exists := sc.ledger[id]; !exists {
        err := fmt.Errorf("asset %s does not exist", id)
        log.Error().Err(err).Msg("Failed to update asset")
        return err
    }

    sc.ledger[id] = Asset{
        ID:    id,
        Value: value,
    }

    log.Info().Str("id", id).Msg("Asset updated")
    return nil
}

// DeleteAsset removes an asset from the ledger
func (sc *SmartContract) DeleteAsset(id string) error {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()

    if _, exists := sc.ledger[id]; !exists {
        err := fmt.Errorf("asset %s does not exist", id)
        log.Error().Err(err).Msg("Failed to delete asset")
        return err
    }

    delete(sc.ledger, id)

    log.Info().Str("id", id).Msg("Asset deleted")
    return nil
}

// Save the current ledger to file (for persistence)
func (sc *SmartContract) saveToFile() {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()

	// 确保目录存在
    if err := os.MkdirAll(sc.filepath, 0755); err != nil {
        log.Error().Err(err).Msg("Failed to create directory")
        return
    }

    data, err := json.MarshalIndent(sc.ledger, "", "  ")
    if err != nil {
        log.Error().Err(err).Msg("Error saving ledger")
        return
    }

	file := filepath.Join(sc.filepath, "ledger.json")

    err = os.WriteFile(file, data, 0644)
    if err != nil {
        log.Error().Err(err).Msg("Error writing to file")
    } else {
        log.Info().Str("file", sc.filepath).Msg("Ledger saved")
    }
}

// Load the ledger from file (for persistence)
func (sc *SmartContract) loadFromFile() {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()

    if _, err := os.Stat(sc.filepath); os.IsNotExist(err) {
        log.Info().Str("file", sc.filepath).Msg("Ledger file does not exist, skipping load")
        return // File doesn't exist, skip loading
    }

    data, err := os.ReadFile(sc.filepath)
    if err != nil {
        log.Error().Err(err).Msg("Error reading file")
        return
    }

    err = json.Unmarshal(data, &sc.ledger)
    if err != nil {
        log.Error().Err(err).Msg("Error parsing JSON")
    } else {
        log.Info().Str("file", sc.filepath).Msg("Ledger loaded")
    }
}

// PrintLedger prints the current state of the ledger
func (sc *SmartContract) PrintLedger() {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()

    ledgerJSON, _ := json.MarshalIndent(sc.ledger, "", "  ")
    log.Info().Str("ledger", string(ledgerJSON)).Msg("Current ledger state")
}

func main() {
    // Initialize logger
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

    // Create a new SmartContract instance with persistence to "ledger.json"
    contract := NewSmartContract("ledger.json")

    // Create assets
    contract.CreateAsset("1", []byte("10000"))
    contract.CreateAsset("2", []byte("250000"))

    // Print the current ledger state
    contract.PrintLedger()

    // Read an asset
    asset, err := contract.ReadAsset("1")
    if err != nil {
        log.Error().Err(err).Msg("Error reading asset")
    } else {
        log.Info().Interface("asset", asset).Msg("Read Asset")
    }

    // Update an asset
    err = contract.UpdateAsset("1", []byte("12000"))
    if err != nil {
        log.Error().Err(err).Msg("Error updating asset")
    }

    // Print the updated ledger state
    contract.PrintLedger()

    // Delete an asset
    err = contract.DeleteAsset("2")
    if err != nil {
        log.Error().Err(err).Msg("Error deleting asset")
    }

    // Print the final ledger state
    contract.PrintLedger()
}