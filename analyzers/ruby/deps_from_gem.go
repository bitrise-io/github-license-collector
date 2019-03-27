package ruby

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/command"
)

// GetGemDeps ...
func GetGemDeps(repoPath string) (map[string][]string, error) {
	lockFiles, err := getGemlockFiles(repoPath)
	if err != nil {
		return nil, err
	}
	allDeps := map[string][]string{}
	for _, lockFile := range lockFiles {
		deps, err := parseLockfile(lockFile)
		if err != nil {
			return nil, err
		}
		for _, dep := range deps {
			allDeps[dep] = nil
		}
	}

	for dep := range allDeps {
		licenses, err := getLicensesForGem(dep)
		if err != nil {
			return nil, err
		}
		allDeps[dep] = licenses
		time.Sleep(time.Millisecond * 500)
	}
	return allDeps, nil
}

func getGemlockFiles(repoPath string) ([]string, error) {
	gemFiles := []string{}
	err := filepath.Walk(repoPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() && f.Name() == "vendor" {
			return filepath.SkipDir
		}
		if !f.IsDir() && f.Name() == "Gemfile.lock" {
			gemFiles = append(gemFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return gemFiles, nil
}

func parseLockfile(gemlockPath string) ([]string, error) {
	cmd := command.New("ruby", "parseGemlock.rb", gemlockPath)

	output, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return nil, err
	}
	// fmt.Print(output)

	return strings.Split(output, "\n"), nil
}

func getLicensesForGem(gem string) ([]string, error) {
	url := fmt.Sprintf("https://rubygems.org/api/v1/versions/%s.json", gem)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var result []map[string]interface{}
	json.Unmarshal(body, &result)

	if len(result) == 0 {
		return nil, nil
	}
	licensesKey := result[0]["licenses"]
	fmt.Println(licensesKey)
	if licensesKey == nil {
		return nil, nil
	}
	licenses := []string{}
	for _, license := range licensesKey.([]interface{}) {
		licenses = append(licenses, license.(string))
	}
	return licenses, nil
}
