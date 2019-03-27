package npm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
)

func init() {
	cmd := command.New("npm", "install", "-g", "yarn")

	log.Printf("$ %s", cmd.PrintableCommandArgs())
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
	log.Donef("$ %s", cmd.PrintableCommandArgs())
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

	cmd := command.New("yarn", "install")

	log.Printf("$ %s", cmd.PrintableCommandArgs())
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return analyzers.RepositoryLicenseInfos{}, fmt.Errorf("run command: %s", out)
		} else {
			return analyzers.RepositoryLicenseInfos{}, fmt.Errorf("run command: %s", err)
		}
	}
	log.Donef("$ %s", cmd.PrintableCommandArgs())
		
	cmd = command.New("yarn", "licenses")

	log.Printf("$ %s", cmd.PrintableCommandArgs())
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return analyzers.RepositoryLicenseInfos{}, fmt.Errorf("run command: %s", out)
		} else {
			return analyzers.RepositoryLicenseInfos{}, fmt.Errorf("run command: %s", err)
		}
	}
	log.Printf(out)
	log.Donef("$ %s", cmd.PrintableCommandArgs())

	log.Donef("analyze npm deps")

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
