package ruby

import (
	"log"

	"github.com/bitrise-io/github-license-collector/analyzers"
)

type Analyzer struct {
	Name string
}

func (a Analyzer) String() string {
	return a.Name
}

func (_ Analyzer) AnalyzeRepository(repoURL, localSourcePath string) (analyzers.RepositoryLicenseInfos, error) {
	log.Printf("AnalyzeRepository: %s", localSourcePath)

	lockFiles, depToLicenses, err := GetGemDeps(localSourcePath)
	if err != nil {
		log.Printf("Ruby error: %s", err)
		return analyzers.RepositoryLicenseInfos{}, nil
	}

	log.Printf("depToLicense: %s", depToLicenses)

	licenses := []analyzers.LicenseInfo{}
	for dep, licensesArray := range depToLicenses {
		for _, licenseType := range licensesArray {
			licenses = append(licenses, analyzers.LicenseInfo{
				Dependency:  dep,
				LicenseType: licenseType,
			})
		}
	}

	log.Printf("Licenses: %s", licenses)

	if len(lockFiles) > 0 {
		return analyzers.RepositoryLicenseInfos{
			RepositoryURL: lockFiles[0],
			Licenses:      licenses,
		}, nil
	}

	return analyzers.RepositoryLicenseInfos{}, nil
}
