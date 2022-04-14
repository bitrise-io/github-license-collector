package golang

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
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
	cmd := command.New("go-licenses", "csv", ".").SetDir(a.localSourcePath)
	out, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return analyzers.RepositoryLicenseInfos{}, errors.New(out)
		}
		return analyzers.RepositoryLicenseInfos{}, err
	}

	licences, err := csvToLicenseInfos(out)
	if err != nil {
		return analyzers.RepositoryLicenseInfos{}, err
	}

	return analyzers.RepositoryLicenseInfos{
		RepositoryURL: a.repoURL,
		Licenses:      licences,
	}, nil
}

func csvToLicenseInfos(csv string) ([]analyzers.LicenseInfo, error) {
	licenses := []analyzers.LicenseInfo{}
	for _, line := range strings.Split(csv, "\n") {
		csvParts := strings.Split(line, ",")
		// NOTE: a csv line should include 3 components (see https://github.com/google/go-licenses):
		// 1.: dependency/package name (URL) (e.g. "https://github.com/grpc/grpc-go/blob/master/LICENSE")
		// 2.: the license's URL (a remote URL to the license file, if any) (e.g. "google.golang.org/grpc")
		// 3.: the type ID of the license (e.g. "Apache-2.0")
		if len(csvParts) != 3 {
			return []analyzers.LicenseInfo{}, fmt.Errorf("invalid CSV line (number of csv parts != 3) : %s", csv)
		}
		licenses = append(licenses, analyzers.LicenseInfo{Dependency: csvParts[0], LicenseType: csvParts[2]})
	}
	return licenses, nil
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
