package keymanager

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
)

// model priority - lower index = higher priority (newest/best first)
// image models commented out for now
var modelPriority = []string{
	"gemini-3-pro-preview",
	// "gemini-3-pro-image-preview",
	// "gemini-2.5-flash-image",
	"gemini-2.5-pro",
	"gemini-flash-latest",
	"gemini-flash-lite-latest",
	"gemini-2.5-flash",
	"gemini-2.5-flash-lite",
	"gemini-2.0-flash",
	"gemini-2.0-flash-lite",
}

type Key struct {
	Value         string    `csv:"key"`
	Provider      string    `csv:"provider"`
	Active        bool      `csv:"active"`
	WorkingModels string    `csv:"working_models"`
	CheckedAt     time.Time `csv:"checked_at"`
	priority      int       // computed, not from csv
}

type KeyManager struct {
	keys    []Key
	mu      sync.RWMutex
	csvPath string
}

func New(csvPath string) (*KeyManager, error) {
	km := &KeyManager{
		csvPath: csvPath,
	}

	if err := km.loadKeys(); err != nil {
		return km, fmt.Errorf("failed to load keys: %w", err)
	}

	return km, nil
}

// getBestPriority returns the best (lowest) priority for a key's working models
func getBestPriority(workingModels string) int {
	models := strings.Split(workingModels, "|")
	best := len(modelPriority) + 1 // worst possible

	for _, model := range models {
		// strip "models/" prefix if present
		model = strings.TrimPrefix(model, "models/")
		for i, pm := range modelPriority {
			if model == pm {
				if i < best {
					best = i
				}
				break
			}
		}
	}

	return best
}

func (km *KeyManager) loadKeys() error {
	file, err := os.Open(km.csvPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var allKeys []Key
	if err := gocsv.UnmarshalFile(file, &allKeys); err != nil {
		return err
	}

	var activeKeys []Key
	for _, key := range allKeys {
		if key.Active {
			key.priority = getBestPriority(key.WorkingModels)
			activeKeys = append(activeKeys, key)
		}
	}

	// sort by priority (lower = better)
	sort.Slice(activeKeys, func(i, j int) bool {
		return activeKeys[i].priority < activeKeys[j].priority
	})

	km.mu.Lock()
	km.keys = activeKeys
	km.mu.Unlock()

	return nil
}

func (km *KeyManager) GetKey() string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if len(km.keys) == 0 {
		return ""
	}

	// find all keys with the best (lowest) priority
	bestPriority := km.keys[0].priority
	var bestKeys []Key
	for _, k := range km.keys {
		if k.priority == bestPriority {
			bestKeys = append(bestKeys, k)
		} else {
			break // already sorted, no need to continue
		}
	}

	// random among best keys
	idx := rand.Intn(len(bestKeys))
	return bestKeys[idx].Value
}

func (km *KeyManager) RotateKey(failedKey string) string {
	km.mu.Lock()
	defer km.mu.Unlock()

	filtered := make([]Key, 0, len(km.keys))
	for _, key := range km.keys {
		if key.Value != failedKey {
			filtered = append(filtered, key)
		}
	}

	km.keys = filtered

	if len(km.keys) == 0 {
		return ""
	}

	// return best available key
	bestPriority := km.keys[0].priority
	var bestKeys []Key
	for _, k := range km.keys {
		if k.priority == bestPriority {
			bestKeys = append(bestKeys, k)
		} else {
			break
		}
	}

	idx := rand.Intn(len(bestKeys))
	return bestKeys[idx].Value
}

func (km *KeyManager) ActiveCount() int {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return len(km.keys)
}

func (km *KeyManager) MarkInactive(key string) error {
	file, err := os.Open(km.csvPath)
	if err != nil {
		return err
	}

	var allKeys []Key
	if err := gocsv.UnmarshalFile(file, &allKeys); err != nil {
		file.Close()
		return err
	}
	file.Close()

	for i, k := range allKeys {
		if k.Value == key {
			allKeys[i].Active = false
			break
		}
	}

	outFile, err := os.Create(km.csvPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := gocsv.MarshalFile(&allKeys, outFile); err != nil {
		return err
	}

	km.mu.Lock()
	filtered := make([]Key, 0, len(km.keys))
	for _, k := range km.keys {
		if k.Value != key {
			filtered = append(filtered, k)
		}
	}
	km.keys = filtered
	km.mu.Unlock()

	return nil
}
