package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	ConfigPath  string
	SongsPath   string
	ApStatePath string
}

func GetConfigDir() (string, error) {
	// locate the user config diretory
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// appConfigDir = the walkmoon dir in user config dir
	appConfigDir := filepath.Join(userConfig, "walkmoon")

	// create .config/walkmoon with read only permission for all, write only for user
	err = os.MkdirAll(appConfigDir, 0o755)
	if err != nil {
		return "", err
	}

	return appConfigDir, nil
}

func LoadState[T any]() (*T, string, error) {
	var state *T
	var defaultDir string

	// get the config path for any kind of SO
	configDir, err := GetConfigDir()
	if err != nil {
		log.Printf("Aviso: Não foi possível obter o diretório de configuração: %v", err)
		configDir = "."
	}

	// statePath points for playerState file in configDir
	statePath := filepath.Join(configDir, "playerState.json")

	// read file in statePath
	jsonData, err := os.ReadFile(statePath)
	if err != nil {
		// if statePlayer.json is not created, try to find home diretocry in any SO
		homeDir, errHome := os.UserHomeDir()

		// if can't find Defaul Music dir, panic path
		if errHome != nil {
			defaultDir = "."
		} else {
			// if homeDir, then puts Music dir in default
			defaultDir = filepath.Join(homeDir, "Músicas")
		}

		return nil, defaultDir, nil
	}

	if err := json.Unmarshal(jsonData, &state); err != nil {
		log.Printf("Aviso: Falha ao decodificar %s: %v", statePath, err)

		homeDir, _ := os.UserHomeDir()
		return nil, filepath.Join(homeDir, "Music"), nil
	}

	return state, defaultDir, nil
}

func SaveState(state any) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	statePath := filepath.Join(configDir, "playerState.json")

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0o644)
}
