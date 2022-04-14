package npm

import (
	"testing"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/stretchr/testify/require"
)

func Test_npmJSONsToLicenseInfos(t *testing.T) {
	t.Log("single line OK")
	{
		npmJSONs := `{"type":"table","data":{"head":["Name","Version","License","URL","VendorUrl","VendorName"],"body":[["@algolia/cache-browser-local-storage","4.11.0","MIT","git://github.com/algolia/algoliasearch-client-javascript.git","Unknown","Unknown"],["@algolia/cache-common","4.11.0","MIT","git://github.com/algolia/algoliasearch-client-js.git","Unknown","Unknown"]]}}`
		licenses, err := npmJSONsToLicenseInfos(npmJSONs)
		require.NoError(t, err)
		require.Equal(t,
			[]analyzers.LicenseInfo{
				{
					LicenseType: "MIT",
					Dependency:  "@algolia/cache-browser-local-storage",
				},
				{
					LicenseType: "MIT",
					Dependency:  "@algolia/cache-common",
				},
			}, licenses)
	}
	t.Log("multi line OK")
	{
		npmJSONs := `{"type":"warning","data":"@bitrise/bitkit > react-popper > popper.js@1.16.1: You can find the new Popper v2 at @popperjs/core, this package is dedicated to the legacy v1"}
{"type":"warning","data":"webpack > watchpack > watchpack-chokidar2 > chokidar@2.1.8: Chokidar 2 will break on node v14+. Upgrade to chokidar 3 with 15x less dependencies."}
{"type":"table","data":{"head":["Name","Version","License","URL","VendorUrl","VendorName"],"body":[["@algolia/cache-browser-local-storage","4.11.0","MIT","git://github.com/algolia/algoliasearch-client-javascript.git","Unknown","Unknown"],["@algolia/cache-common","4.11.0","MIT","git://github.com/algolia/algoliasearch-client-js.git","Unknown","Unknown"]]}}
`
		licenses, err := npmJSONsToLicenseInfos(npmJSONs)
		require.NoError(t, err)
		require.Equal(t,
			[]analyzers.LicenseInfo{
				{
					LicenseType: "MIT",
					Dependency:  "@algolia/cache-browser-local-storage",
				},
				{
					LicenseType: "MIT",
					Dependency:  "@algolia/cache-common",
				},
			}, licenses)
	}

	// --- Error tests ---
	{
		licenses, err := npmJSONsToLicenseInfos("")
		require.EqualError(t, err, "no license information provided by yarn")
		require.Equal(t, licenses, []analyzers.LicenseInfo{})
	}

	t.Log("invalid JSON")
	{
		licenses, err := npmJSONsToLicenseInfos(`{"not-a-valid-json":[}`)
		require.EqualError(t, err, "no license information provided by yarn")
		require.Equal(t, licenses, []analyzers.LicenseInfo{})
	}

	{
		npmJSONs := `{"type":"table","data":{"head":["Name","Version","License","URL","VendorUrl","VendorName"],"body":[]}}`
		licenses, err := npmJSONsToLicenseInfos(npmJSONs)
		require.EqualError(t, err, "0 license information found")
		require.Equal(t, licenses, []analyzers.LicenseInfo{})
	}
}
