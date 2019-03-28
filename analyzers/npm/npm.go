package npm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
)

func init() {
	cmd := command.New("npm", "install", "-g", "yarn")

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			log.Errorf("run command: %s", out)
			os.Exit(1)
		} else {
			log.Errorf("run command: %s", err)
			os.Exit(1)
		}
	}
}

type npmLicensesList struct {
	Type string `json:"type"`
	Data Data   `json:"data"`
}

type Data struct {
	Head []string   `json:"head"`
	Body [][]string `json:"body"`
}

type Analyzer struct {
	repoURL, localSourcePath string
}

func (a Analyzer) String() string {
	return "npm"
}

func (a *Analyzer) Detect(repoURL, localSourcePath string) (bool, error) {
	a.repoURL, a.localSourcePath = repoURL, localSourcePath

	files, err := getNpmDeps(a.localSourcePath)
	if err != nil {
		return false, err
	}

	if len(files) == 0 {
		return false, nil
	}

	return true, nil
}

func (a *Analyzer) AnalyzeRepository() (analyzers.RepositoryLicenseInfos, error) {
	if err := os.Chdir(a.localSourcePath); err != nil {
		return analyzers.RepositoryLicenseInfos{}, fmt.Errorf("change to source dir %s: %s", a.localSourcePath, err)
	}

	cmd := command.New("yarn", "licenses", "list", "--json", "--no-progress")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return analyzers.RepositoryLicenseInfos{}, fmt.Errorf("run command: %s", out)
		} else {
			return analyzers.RepositoryLicenseInfos{}, fmt.Errorf("run command: %s", err)
		}
	}

	var licenses npmLicensesList
	if err := json.Unmarshal([]byte(out), &licenses); err != nil {
		return analyzers.RepositoryLicenseInfos{}, fmt.Errorf("unmarshal yarn licenses output: %s", err)
	}

	headIndexes := map[string]int{}
	for i, header := range licenses.Data.Head {
		headIndexes[strings.ToLower(header)] = i
	}

	licenseInfos := analyzers.RepositoryLicenseInfos{}
	for _, lic := range licenses.Data.Body {
		licenseInfos.Licenses = append(licenseInfos.Licenses, analyzers.LicenseInfo{
			LicenseType: lic[headIndexes["license"]],
			Dependency:  lic[headIndexes["name"]],
		})
	}

	if len(licenseInfos.Licenses) > 0 {
		licenseInfos.RepositoryURL = a.repoURL
	}

	return licenseInfos, nil
}

func getNpmDeps(repoPath string) ([]string, error) {
	gemFiles := []string{}
	err := filepath.Walk(repoPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() && (f.Name() == "vendor" || f.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if !f.IsDir() && (f.Name() == "package.json" || f.Name() == "yarn.lock") {
			gemFiles = append(gemFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return gemFiles, nil
}
