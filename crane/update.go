package crane

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

type UpdateRequestParams struct {
	UUID    string `json:"uuid"`
	Arch    string `json:"arch"`
	OS      string `json:"os"`
	Version string `json:"version"`
	Pro     bool   `json:"pro"`
}

type UpdateResponseBody struct {
	Outdated              bool   `json:"outdated"`
	LatestVersion         string `json:"latest_version"`
	LatestReleaseDate     string `json:"latest_release_date"`
	LatestInstallationUrl string `json:"latest_installation_url"`
	LatestChangelogUrl    string `json:"latest_changelog_url"`
}

func checkForUpdates(manual bool) error {
	client := &http.Client{Timeout: 3 * time.Second}
	params := UpdateRequestParams{
		UUID:    settings.UUID,
		Arch:    runtime.GOARCH,
		OS:      runtime.GOOS,
		Version: Version,
		Pro:     Pro,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(params)

	response := &UpdateResponseBody{}
	printInfof("Checking for updates ...\n")
	res, err := client.Post(
		"https://www.craneup.tech/update-checks",
		"application/json; charset=utf-8",
		b,
	)
	if err == nil && res.StatusCode != 200 {
		msg := fmt.Sprintf("Wrong status code %s", res.Status)
		err = errors.New(msg)
	}
	if err != nil {
		if manual {
			printErrorf("ERROR: %s\n", err)
		} else {
			verboseMsg(fmt.Sprintf("Update check failed: %s", err.Error()))
		}
		settings.DelayNextUpdateCheck()
		return nil
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(response)

	if response.Outdated {
		printNoticef("Newer version %s is available!\n\n", response.LatestVersion)
		fmt.Printf("\tRelease Date: %s\n", response.LatestReleaseDate)
		fmt.Printf("\tChangelog: %s\n", response.LatestChangelogUrl)
		fmt.Printf("\nUpdate now: %s\n\n", response.LatestInstallationUrl)
		return settings.Update(response.LatestVersion)
	}
	printSuccessf("Version %s is the latest version!\n\n", Version)
	return settings.Update(Version)
}

func autoUpdateCheckInterval() time.Duration {
	return 7 * 24 * time.Hour
}
