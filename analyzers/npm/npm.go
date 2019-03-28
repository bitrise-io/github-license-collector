package npm

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
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
	Name string
}

func (a Analyzer) String() string {
	return a.Name
}

func (_ Analyzer) AnalyzeRepository(repoURL, localSourcePath string) (analyzers.RepositoryLicenseInfos, error) {

	if err := os.Chdir(localSourcePath); err != nil {
		return analyzers.RepositoryLicenseInfos{}, fmt.Errorf("change to source dir %s: %s", localSourcePath, err)
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

	licenseInfos := analyzers.RepositoryLicenseInfos{RepositoryURL: repoURL}
    for _, lic := range licenses.Data.Body {
		licenseInfos.Licenses = append(licenseInfos.Licenses, analyzers.LicenseInfo{
			LicenseType: lic[headIndexes["license"]],
			Dependency: lic[headIndexes["name"]],
		})
	}
	log.Donef("analyze npm deps")

	return licenseInfos, nil
}
