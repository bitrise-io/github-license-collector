package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/github-license-collector/analyzers/golang"
	"github.com/bitrise-io/github-license-collector/analyzers/npm"
	"github.com/bitrise-io/github-license-collector/analyzers/ruby"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type analyzer interface {
	AnalyzeRepository(repoURL, localSourcePath string) (analyzers.RepositoryLicenseInfos, error)
}

var analyzerTools = []analyzer{
	golang.Analyzer{},
	npm.Analyzer{},
	ruby.Analyzer{},
}

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: collect,
}

var (
	flagOrg string
)

func init() {
	RootCmd.AddCommand(collectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// collectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	collectCmd.Flags().StringVar(&flagOrg, "org", "", "GitHub Org")
}

func collect(cmd *cobra.Command, args []string) error {
	if len(flagOrg) < 1 {
		return errors.New("Organization not specified")
	}
	githubPersonalAccessToken := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if len(githubPersonalAccessToken) < 1 {
		return errors.New("GITHUB_PERSONAL_ACCESS_TOKEN env var isn't specified")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubPersonalAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opt := &github.RepositoryListByOrgOptions{Type: "all"}
	// get all pages of results
	log.Infof("fetching list of repos")
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(context.Background(), flagOrg, opt)
		if err != nil {
			return errors.WithStack(err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	log.Donef("done")
	fmt.Println()

	log.Infof("cloning repos")
	type repo struct {
		url, path string
	}

	tempPath, err := pathutil.NormalizedOSTempDirPath("temp")
	if err != nil {
		log.Errorf("Failed to create temp path, error: %s", err)
		os.Exit(1)
	}

	repoChan := make(chan repo)
	for _, aRepo := range allRepos {
		go func(r *github.Repository) {
			log.Infof("Start clone: %s", r.GetSSHURL())

			goPath := filepath.Join("github.com", r.GetFullName())
			fullPath := filepath.Join(tempPath, "src", goPath)

			if err := cloneRepo(r.GetSSHURL(), fullPath); err != nil {
				log.Errorf("Failed to clone repo(%s), error: %s", r.GetSSHURL(), err)
				os.Exit(1)
			}

			repoChan <- repo{r.GetURL(), fullPath}
			log.Donef("Cloned: %s - %s", r.GetSSHURL(), fullPath)
		}(aRepo)
	}

	processedRepos := 0
	for {
		r := <-repoChan

		for _, a := range analyzerTools {
			infos, err := a.AnalyzeRepository(r.url, r.path)
			if err != nil {
				log.Errorf("failed to analyze repo: %s, error: %s", r.url, err)
				continue
			}
			_ = infos
		}

		processedRepos++
		if len(allRepos) == processedRepos {
			break
		}
	}

	log.Donef("repos scanned: %d", len(allRepos))
	return nil
}

func cloneRepo(url, path string) error {
	out, err := command.New("git", "clone", "--depth", "1", url, path).RunAndReturnTrimmedCombinedOutput()
	return errors.Wrap(err, out)
}
