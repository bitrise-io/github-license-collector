package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/bitrise-io/github-license-collector/analyzers/golang"
	"github.com/bitrise-io/github-license-collector/analyzers/npm"
	"github.com/bitrise-io/github-license-collector/analyzers/ruby"
	"github.com/bitrise-io/github-license-collector/utils"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type repo struct {
	url, path string
}

type analyzer interface {
	Detect(repoURL, localSourcePath string) (bool, error)
	AnalyzeRepository() (analyzers.RepositoryLicenseInfos, error)
	String() string
}

var analyzerTools = []analyzer{
	&golang.Analyzer{},
	&npm.Analyzer{},
	&ruby.Analyzer{},
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
	flagOrg        string
	reposCachePath = "repos-cache"
	outputFilePath = "output.txt"
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

	absReposCachePath, err := pathutil.AbsPath(reposCachePath)
	if err != nil {
		return errors.WithStack(err)
	}

	tempPath := absReposCachePath
	totalReposCount := len(allRepos)
	reposList := make([]repo, totalReposCount, totalReposCount)

	os.Setenv("TMP_GOPATH", tempPath)

	var wg sync.WaitGroup
	wg.Add(totalReposCount)
	counter := utils.NewThreadSafeCounter()

	for _, aRepo := range allRepos {
		go gitCloneRepoAsync(aRepo, &wg, tempPath, counter, reposList, totalReposCount)
	}

	// Wait for all git clones to finish
	wg.Wait()
	log.Infof("Cloning repos finished, starting analyzing...")

	processedRepos := 0
	var others []string
	typeMap := map[string]int{}
	typeURLs := map[string][]string{}
	var allInfos []analyzers.RepositoryLicenseInfos
	for _, a := range analyzerTools {
		typeMap[a.String()] = 0
		typeURLs[a.String()] = []string{}
	}
	failedAnalyzes := []repo{}
	for _, r := range reposList {
		other := true
		for _, a := range analyzerTools {
			log.Printf("Check repo: %s", r.url)
			if detected, err := a.Detect(r.url, r.path); err != nil {
				log.Errorf("failed to detect analyzer(%s) for: %s, error: %s", a.String(), r.url, err)
				failedAnalyzes = append(failedAnalyzes, r)
				continue
			} else if !detected {
				continue
			}

			log.Printf("- running %s analyzer", a.String())

			info, err := a.AnalyzeRepository()
			if err != nil {
				log.Errorf("failed to analyze repo: %s, error: %s", r.url, err)
				failedAnalyzes = append(failedAnalyzes, r)
				continue
			}

			if info.RepositoryURL != "" {
				allInfos = append(allInfos, info)
				typeMap[a.String()]++
				typeURLs[a.String()] = append(typeURLs[a.String()], info.RepositoryURL)
				other = false
			}
		}

		if other {
			others = append(others, r.url)
		}

		processedRepos++
		if len(allRepos) == processedRepos {
			break
		}
	}

	log.Donef("repos scanned: %d", len(allRepos))
	for _, a := range analyzerTools {
		log.Infof("%s: %d", a.String(), typeMap[a.String()])
		log.Printf("- %s", strings.Join(typeURLs[a.String()], "\n- "))
	}
	log.Infof("other: %d", len(others))
	log.Printf("- %s", strings.Join(others, "\n- "))

	fmt.Println()
	log.Infof("failed: %d", len(failedAnalyzes))
	for _, aFailed := range failedAnalyzes {
		log.Printf("- %s: %s", aFailed.url, aFailed.path)
	}

	fmt.Println()

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to create output file")
	}
	log.Infof("Repository dependency licenses:")
	fmt.Fprintln(outputFile, "# Repository dependency licenses:")

	licenceTypes := map[string]int{}
	for _, info := range allInfos {
		log.Infof("\n%s:", info.RepositoryURL)
		fmt.Fprintf(outputFile, "\n## %s:\n", info.RepositoryURL)
		if len(info.Licenses) == 0 {
			log.Infof("No license info found")
			fmt.Fprintf(outputFile, "No license info found\n")
			continue
		}

		for _, dep := range info.Licenses {
			log.Printf("- %s: %s", dep.Dependency, dep.LicenseType)
			fmt.Fprintf(outputFile, "- %s: %s\n", dep.Dependency, dep.LicenseType)
			licenceTypes[dep.LicenseType]++
		}
	}

	fmt.Println()

	log.Infof("licence types (# of dependency uses):")
	allLicenceUsage := 0
	for lType, used := range licenceTypes {
		log.Printf("- %s: %d", lType, used)
		allLicenceUsage += used
	}
	log.Printf("%d licences used")
	return nil
}

func cloneRepo(url, path string) error {
	out, err := command.New("git", "clone", "--depth", "1", url, path).RunAndReturnTrimmedCombinedOutput()
	return errors.Wrap(err, out)
}

func gitCloneRepoAsync(r *github.Repository, wg *sync.WaitGroup, tempPath string, counter *utils.ThreadSafeCounter, reposList []repo, totalReposCount int) {
	defer wg.Done()
	log.Infof("Start clone: %s", r.GetSSHURL())

	goPath := filepath.Join("github.com", r.GetFullName())
	fullPath := filepath.Join(tempPath, "src", goPath)

	if exists, err := pathutil.IsPathExists(fullPath); err != nil {
		log.Errorf("Failed to check if path (%s) exists: %+v", fullPath, err)
	} else if exists {
		log.Warnf("Path (%s) already exists - Skipping git clone", fullPath)
	} else {
		if err := cloneRepo(r.GetSSHURL(), fullPath); err != nil {
			log.Errorf("Failed to clone repo(%s), error: %s", r.GetSSHURL(), err)
			os.Exit(1)
		}
	}

	counerVal := counter.Increment()
	reposList[counerVal-1] = repo{r.GetCloneURL(), fullPath}
	log.Donef("Cloned [%d/%d]: %s - %s", counerVal, totalReposCount, r.GetSSHURL(), fullPath)
}
