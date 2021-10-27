package npm

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
)

type npmLicensesListTypeOnly struct {
	Type string `json:"type"`
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
	cmd := command.New("yarn", "licenses", "list", "--json", "--no-progress").SetDir(a.localSourcePath)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return analyzers.RepositoryLicenseInfos{}, errors.New(out)
		}
		return analyzers.RepositoryLicenseInfos{}, err
	}

	var licenses npmLicensesList
	for _, line := range strings.Split(out, "\n") {
		var lType npmLicensesListTypeOnly
		if err := json.Unmarshal([]byte(line), &lType); err != nil {
			log.Warnf("unmarshal yarn licenses type output: %s | line: %s", err, line)
		}
		if lType.Type == "table" {
			var l npmLicensesList
			if err := json.Unmarshal([]byte(line), &l); err != nil {
				log.Warnf("unmarshal yarn licenses data output: %s | line: %s", err, line)
			}
			licenses = l
			break
		}
	}
	if licenses.Type == "" {
		return analyzers.RepositoryLicenseInfos{}, nil
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
