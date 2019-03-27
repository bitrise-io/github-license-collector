package npm

import (
	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/go-utils/log"
)

type Analyzer struct {
}

func (_ Analyzer) AnalyzeRepository(repoURL, localSourcePath string) (analyzers.RepositoryLicenseInfos, error) {
	log.Donef("analyze npm deps")
	return analyzers.RepositoryLicenseInfos{}, nil
}
