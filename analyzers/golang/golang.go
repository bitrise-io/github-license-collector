// this analyser utilizes: https://github.com/pmezard/licenses, which tool uses go list command to list the used dependencies
// licenses tool lists the licence of the root package, this is filtered out
// ? means that no licence found for the given repository
package golang

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
)

type Analyzer struct {
	Name string
}

func (a Analyzer) String() string {
	return a.Name
}

func (_ Analyzer) AnalyzeRepository(repoURL, localSourcePath string) (analyzers.RepositoryLicenseInfos, error) {
	files, err := getGoDeps(localSourcePath)
	if err != nil {
		return analyzers.RepositoryLicenseInfos{}, nil
	}

	if len(files) == 0 {
		return analyzers.RepositoryLicenseInfos{RepositoryURL: strings.Join(files, ";")}, nil
	}

	rootPackage, err := selfPackageName(localSourcePath)
	if err != nil {
		return analyzers.RepositoryLicenseInfos{}, err
	}

	licByDep, warnings, err := fetchLicences(localSourcePath, rootPackage)
	if err != nil {
		return analyzers.RepositoryLicenseInfos{}, err
	}

	var licences []analyzers.LicenseInfo
	for dep, lic := range licByDep {
		licence := analyzers.LicenseInfo{
			LicenseType: lic,
			Dependency:  dep,
		}
		licences = append(licences, licence)
	}

	err = nil
	if len(warnings) > 0 {
		err = errors.New(warnings)
	}
	return analyzers.RepositoryLicenseInfos{
		RepositoryURL: repoURL,
		Licenses:      licences,
	}, err
}

func getGoDeps(repoPath string) ([]string, error) {
	gemFiles := []string{}
	err := filepath.Walk(repoPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() && f.Name() == "vendor" {
			return filepath.SkipDir
		}
		if !f.IsDir() && (f.Name() == "Gopkg.toml" || f.Name() == "Godeps.json") {
			gemFiles = append(gemFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return gemFiles, nil
}

func fetchLicences(rootDir, rootPackage string) (map[string]string, string, error) {
	pth, err := exec.LookPath("licenses")
	if pth == "" || err != nil {
		out, err := command.
			New("go", "get", "github.com/pmezard/licenses").
			AppendEnvs("GOPATH=" + os.Getenv("TMP_GOPATH")).
			RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			if errorutil.IsExitStatusError(err) {
				return nil, "", errors.New(out)
			}
			return nil, "", err
		}
	}

	out, err := command.
		New("licenses", "./...").
		AppendEnvs("GOPATH=" + os.Getenv("TMP_GOPATH")).
		SetDir(rootDir).
		RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return nil, "", errors.New(out)
		}
		return nil, "", err
	}

	licenceByDep := map[string]string{}
	var warnings []string
	for _, line := range strings.Split(out, "\n") {
		var dep string
		var licElems []string
		for _, p := range strings.Split(line, " ") {
			if len(p) == 0 {
				continue
			}
			if len(dep) == 0 {
				dep = p
				continue
			}
			licElems = append(licElems, p)
		}
		if len(dep) == 0 {
			return nil, "", fmt.Errorf("failed to parse: %s", line)
		}
		lic := strings.Join(licElems, " ")
		if len(lic) == 0 {
			return nil, "", fmt.Errorf("no licence found for dep: %s", dep)
		}
		if dep == rootPackage {
			continue
		}
		if strings.Contains(lic, "cannot find package") {
			warnings = append(warnings, lic)
			continue
		}
		licenceByDep[dep] = lic
	}
	return licenceByDep, strings.Join(warnings, "\n"), nil
}

func selfPackageName(rootDir string) (string, error) {
	out, err := command.
		New("go", "list").
		AppendEnvs("GOPATH=" + os.Getenv("TMP_GOPATH")).
		SetDir(rootDir).
		RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return "", errors.New(out)
		}
		return "", err
	}
	packages := strings.Split(out, "\n")

	var rootPackage string
	for _, p := range packages {
		if rootPackage == "" {
			rootPackage = p
			continue
		}
		if len(strings.Split(p, "/")) < len(strings.Split(rootPackage, "/")) {
			rootPackage = p
		}
	}
	return strings.TrimSpace(rootPackage), nil
}
