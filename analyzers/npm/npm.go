package npm

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/github-license-collector/analyzers"
)

type Analyzer struct {
	Name string
}

func (a Analyzer) String() string {
	return a.Name
}

func (_ Analyzer) AnalyzeRepository(repoURL, localSourcePath string) (analyzers.RepositoryLicenseInfos, error) {
	files, err := getNpmDeps(localSourcePath)
	if err != nil {
		return analyzers.RepositoryLicenseInfos{}, nil
	}

	if len(files) > 0 {
		return analyzers.RepositoryLicenseInfos{RepositoryURL: strings.Join(files, ";")}, nil
	}

	return analyzers.RepositoryLicenseInfos{}, nil
}

func getNpmDeps(repoPath string) ([]string, error) {
	gemFiles := []string{}
	err := filepath.Walk(repoPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() && f.Name() == "vendor" {
			return filepath.SkipDir
		}
		if !f.IsDir() && f.Name() == "package.json" {
			gemFiles = append(gemFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return gemFiles, nil
}
