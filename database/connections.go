package database

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SavedConnection holds a saved database connection
type SavedConnection struct {
	Name     string           `json:"name"`
	Config   ConnectionConfig `json:"config"`
}

// ConnectionStore manages saved connections
type ConnectionStore struct {
	Connections []SavedConnection `json:"connections"`
	filePath    string
}

// NewConnectionStore creates a new connection store
func NewConnectionStore() (*ConnectionStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".syncforge")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	store := &ConnectionStore{
		filePath: filepath.Join(configDir, "connections.json"),
	}

	store.load()
	return store, nil
}

// load reads connections from file
func (s *ConnectionStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.Connections = []SavedConnection{}
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &s.Connections)
}

// save writes connections to file
func (s *ConnectionStore) save() error {
	data, err := json.MarshalIndent(s.Connections, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0600)
}

// GetAll returns all saved connections
func (s *ConnectionStore) GetAll() []SavedConnection {
	return s.Connections
}

// Save adds or updates a connection
func (s *ConnectionStore) Save(conn SavedConnection) error {
	// Check if connection with same name exists
	for i, c := range s.Connections {
		if c.Name == conn.Name {
			s.Connections[i] = conn
			return s.save()
		}
	}

	// Add new connection
	s.Connections = append(s.Connections, conn)
	return s.save()
}

// Delete removes a connection by name
func (s *ConnectionStore) Delete(name string) error {
	for i, c := range s.Connections {
		if c.Name == name {
			s.Connections = append(s.Connections[:i], s.Connections[i+1:]...)
			return s.save()
		}
	}
	return nil
}
