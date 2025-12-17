package keymanager

import (
	"fmt"
	"math/rand"
	"os"
	"sync"

	"github.com/gocarina/gocsv"
)

type Key struct {
	Value    string `csv:"key"`
	Provider string `csv:"provider"`
	Active   bool   `csv:"active"`
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
			activeKeys = append(activeKeys, key)
		}
	}

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

	idx := rand.Intn(len(km.keys))
	return km.keys[idx].Value
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

	idx := rand.Intn(len(km.keys))
	return km.keys[idx].Value
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
