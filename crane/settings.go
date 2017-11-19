package crane

import (
	"encoding/json"
	"errors"
	"fmt"
	uuid "github.com/hashicorp/go-uuid"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var settings *Settings

type Settings struct {
	filename        string
	UUID            string    `json:"uuid"`
	Version         string    `json:"version"`
	LatestVersion   string    `json:"latest_version"`
	LastUpdateCheck time.Time `json:"last_update_check"`
	CheckForUpdates bool      `json:"check_for_updates"`
}

// Determine crane settings base path.
// On windows, this is %APPDATA%\\crane
// On unix, this is ${XDG_CONFIG_HOME}/crane (which usually
// is ${HOME}/.config)
func settingsPath() (string, error) {
	settingsPath := os.Getenv("CRANE_SETTINGS_PATH")
	if len(settingsPath) > 0 {
		return settingsPath, nil
	}
	if runtime.GOOS == "windows" {
		settingsPath = os.Getenv("APPDATA")
		if len(settingsPath) > 0 {
			return fmt.Sprintf("%s/crane", settingsPath), nil
		}
		return "", errors.New("Cannot detect settings path!")
	}
	settingsPath = os.Getenv("XDG_CONFIG_HOME")
	if len(settingsPath) > 0 {
		return fmt.Sprintf("%s/crane", settingsPath), nil
	}
	homeDir := os.Getenv("HOME")
	if len(homeDir) > 0 {
		return fmt.Sprintf("%s/.config/crane", homeDir), nil
	}
	return "", errors.New("Cannot detect settings path!")
}

func createSettings(filename string) error {
	uuid, _ := uuid.GenerateUUID()
	settings = &Settings{
		filename:        filename,
		UUID:            uuid,
		Version:         Version,
		LatestVersion:   Version,
		LastUpdateCheck: time.Now(),
		CheckForUpdates: true,
	}
	msg := fmt.Sprintf("Writing settings file to %s\n", filename)
	printInfof(msg)
	return settings.Write(filename)
}

func readSettings() error {
	// create path if it does not exist yet
	sp, err := settingsPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(sp); err != nil {
		os.MkdirAll(sp, os.ModePerm)
		if _, err := os.Stat(sp); err != nil {
			return err
		}
	}

	// create file if it does not exist yet
	filename := filepath.Join(sp, "config.json")
	if _, err := os.Stat(filename); err != nil {
		return createSettings(filename)
	}

	// read settings of file
	settings = &Settings{filename: filename}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(StatusError{err, 74})
	}
	err = json.Unmarshal(data, settings)
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(StatusError{err, 65})
	}
	return nil
}

func (s *Settings) ShouldCheckForUpdates() bool {
	if !s.CheckForUpdates {
		return false
	}
	week := 7 * 24 * time.Hour
	dontCheckUntil := settings.LastUpdateCheck.Add(week)
	return time.Now().After(dontCheckUntil)
}

// If version in settings does not match version of binary,
// we assume that the binary was updated and update the
// settings file with the new information.
func (s *Settings) CorrectVersion() error {
	if Version != s.Version {
		s.Version = Version
		return s.Update(Version)
	}
	return nil
}

func (s *Settings) Update(latestVersion string) error {
	s.LastUpdateCheck = time.Now()
	s.LatestVersion = latestVersion
	return s.Write(s.filename)
}

func (s *Settings) Write(filename string) error {
	contents, _ := json.Marshal(s)
	return ioutil.WriteFile(filename, contents, 0644)
}
