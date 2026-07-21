import config

import (
	"encondig/json"
	"os"
	"log"
	"path/filepath"
)

type Config struct {
	ConfigPath string
	SongsPath string
	ApStatePath string
}

func getConfigDir() (string, error) {
	// locate the user config diretory
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// appConfigDir = the walkmoon dir in user config dir 
	appConfigDir = filepath.join(userConfig, "walkmoon")

	// create .config/walkmoon with read only permission for all, write only for user
	err = MkdirAll(appConfigDir, 0755)
	if err != nil {
		return "", err
	}
	
	return appConfigDir, nil
}

func LoadState[T any]() (*T, string) {
	var state *T
	var defaultDir string

	configDir, err := GetConfigDir()
	if err != nil {
		log.Printf("Aviso: Não foi possível obter o diretório de configuração: %v", err)
		configDir = "."
	}

	statePath := filepath.Join(configDir, "playerState.json")

	jsonData, err := os.ReadFile(statePath)
	if err != nil {
		// Se não existe arquivo salvo, usa o diretório ~/Music do usuário
		homeDir, errHome := os.UserHomeDir()
		if errHome != nil {
			defaultDir = "."
		} else {
			defaultDir = filepath.Join(homeDir, "Music")
		}
		return nil, defaultDir
	}

	if err := json.Unmarshal(jsonData, &state); err != nil {
		log.Printf("Aviso: Falha ao decodificar %s: %v", statePath, err)
		
		homeDir, _ := os.UserHomeDir()
		return nil, filepath.Join(homeDir, "Music")
	}

	return state, defaultDir
}

// SaveState salva o estado atual do player em ~/.config/walkmoon/playerState.json
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

	return os.WriteFile(statePath, data, 0644)
}
