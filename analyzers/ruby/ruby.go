package ruby

import (
	"os"
	"path/filepath"

	"github.com/bitrise-io/github-license-collector/analyzers"
)

type Analyzer struct {
	repoURL, localSourcePath string
}

func (a Analyzer) String() string {
	return "ruby"
}

func (a *Analyzer) Detect(repoURL, localSourcePath string) (bool, error) {
	a.repoURL, a.localSourcePath = repoURL, localSourcePath

	files, err := getRubyDeps(a.localSourcePath)
	if err != nil {
		return false, err
	}

	if len(files) == 0 {
		return false, nil
	}

	return true, nil
}

func (a *Analyzer) AnalyzeRepository() (analyzers.RepositoryLicenseInfos, error) {
	_, depToLicenses, err := GetGemDeps(a.localSourcePath)
	if err != nil {
		return analyzers.RepositoryLicenseInfos{}, nil
	}

	licenses := []analyzers.LicenseInfo{}
	for dep, licensesArray := range depToLicenses {
		for _, licenseType := range licensesArray {
			licenses = append(licenses, analyzers.LicenseInfo{
				Dependency:  dep,
				LicenseType: licenseType,
			})
		}
	}

	return analyzers.RepositoryLicenseInfos{
		RepositoryURL: a.repoURL,
		Licenses:      licenses,
	}, nil
}

func getRubyDeps(repoPath string) ([]string, error) {
	gemFiles := []string{}
	err := filepath.Walk(repoPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() && (f.Name() == "vendor") {
			return filepath.SkipDir
		}
		if !f.IsDir() && (f.Name() == "Gemfile") {
			gemFiles = append(gemFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return gemFiles, nil
}
