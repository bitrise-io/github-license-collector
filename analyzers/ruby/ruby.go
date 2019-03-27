package ruby

import (
	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/go-utils/log"
)

type Analyzer struct {
}

func (_ Analyzer) AnalyzeRepository(repoURL, localSourcePath string) (analyzers.RepositoryLicenseInfos, error) {
	log.Warnf("analyze ruby deps")
	return analyzers.RepositoryLicenseInfos{}, nil
}
