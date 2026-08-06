package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	PlayerState    PlayerState    `json:"player_state"`
	PlayerKeybinds PlayerKeybinds `json:"player_keybinds"`
}

type PlayerState struct {
	LastTrackPath string  `json:"path"`
	Volume        float64 `json:"volume"`
	CurrentIndex  int     `json:"index"`
	LoopPlaylist  bool    `json:"loop_playlist"`
	LoopSong      bool    `json:"loop_song"`
	PlayShuffle   bool    `json:"play_shuffle"`
}

type PlayerKeybinds struct {
	NextSong       string `json:"next_song"`
	PreviousSong   string `json:"previous_song"`
	PausePlayer    string `json:"pause_player"`
	MutePlayer     string `json:"mute_player"`
	AdvanceSong    string `json:"advance_song"`   // avança 5 segundos na música
	ComeBackSong   string `json:"come_back_song"` // volta 5 segundos na música
	FilePicker     string `json:"file_picker"`    // permite o usuário escolher uma música dentre suas pastas
	ShuffleMode    string `json:"shuffle"`
	LoopSongMode   string `json:"loop_song"`
	LoopPlaylist   string `json:"loop_playlist"`
	IncreaseVolume string `json:"increase_volume"`
	DecreaseVolume string `json:"decrease_volume"`
	Quit           string `json:"quit"`
}

func GetConfigDir() (string, error) {
	// locate the user config diretory
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// appConfigDir = the walkmoon dir in user config dir
	walkmoonDir := filepath.Join(userConfig, "walkmoon")

	// it makes sure that .config/walkmoon with read only permission, exists
	err = os.MkdirAll(walkmoonDir, 0o755)
	if err != nil {
		return "", err
	}

	return walkmoonDir, nil
}

func LoadConfig[T any](configType string) (*T, error) {
	var target *T

	configDir, err := GetConfigDir()
	if err != nil {
		return target, err
	}

	configFilePath := filepath.Join(configDir, configType+".json")

	jsonData, err := os.ReadFile(configFilePath)
	if err != nil {
		return target, err
	}

	if err := json.Unmarshal(jsonData, &target); err != nil {
		return target, err
	}

	return target, nil
}

func SaveState(configType string, state any) error {
	walkmoonDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(walkmoonDir, configType+".json")

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFilePath, data, 0o644)
}
