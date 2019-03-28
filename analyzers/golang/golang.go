package golang

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	lic "github.com/godrei/licenses/licenses"
)

type Analyzer struct {
	Name            string
	localSourcePath string
	repoURL         string
}

func (a Analyzer) String() string {
	return "go"
}

func (a *Analyzer) Detect(repoURL, localSourcePath string) (bool, error) {
	a.localSourcePath = localSourcePath
	a.repoURL = repoURL

	files, err := getGoDepDescrptors(localSourcePath)
	if err != nil {
		return false, err
	}
	return len(files) > 0, nil
}

func (a Analyzer) AnalyzeRepository() (analyzers.RepositoryLicenseInfos, error) {
	out, err := command.New("go", "get", "./...").SetDir(a.localSourcePath).AppendEnvs("GOPATH=" + os.Getenv("TMP_GOPATH")).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return analyzers.RepositoryLicenseInfos{}, errors.New(out)
		}
		return analyzers.RepositoryLicenseInfos{}, err
	}

	licensesByPackage, err := lic.Licenses(os.Getenv("TMP_GOPATH"), a.localSourcePath)
	if err != nil {
		return analyzers.RepositoryLicenseInfos{}, err
	}

	var licences []analyzers.LicenseInfo
	for p, l := range licensesByPackage {
		licence := analyzers.LicenseInfo{
			LicenseType: l,
			Dependency:  p,
		}
		licences = append(licences, licence)
	}

	return analyzers.RepositoryLicenseInfos{
		RepositoryURL: a.repoURL,
		Licenses:      licences,
	}, err
}

func getGoDepDescrptors(repoPath string) ([]string, error) {
	depDescriptors := []string{}
	err := filepath.Walk(repoPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() && f.Name() == "vendor" {
			return filepath.SkipDir
		}
		if !f.IsDir() && (f.Name() == "Gopkg.toml" || f.Name() == "Godeps.json") || f.Name() == "go.mod" {
			depDescriptors = append(depDescriptors, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return depDescriptors, nil
}
