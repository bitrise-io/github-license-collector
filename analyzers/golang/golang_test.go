package golang

import (
	"testing"

	"github.com/bitrise-io/github-license-collector/analyzers"
	"github.com/stretchr/testify/require"
)

func Test_csvToLicenseInfos(t *testing.T) {
	t.Log("single line csv")
	{
		licenses, err := csvToLicenseInfos(`google.golang.org/grpc,https://github.com/grpc/grpc-go/blob/master/LICENSE,Apache-2.0`)
		require.NoError(t, err)
		require.Equal(t, licenses,
			[]analyzers.LicenseInfo{
				{
					LicenseType: "Apache-2.0",
					Dependency:  "google.golang.org/grpc",
				},
			})
	}
	t.Log("multi line csv")
	{
		csv := `google.golang.org/grpc,https://github.com/grpc/grpc-go/blob/master/LICENSE,Apache-2.0
go.opencensus.io,https://github.com/census-instrumentation/opencensus-go/blob/master/LICENSE,Apache-2.0`
		licenses, err := csvToLicenseInfos(csv)
		require.NoError(t, err)
		require.Equal(t, licenses,
			[]analyzers.LicenseInfo{
				{
					LicenseType: "Apache-2.0",
					Dependency:  "google.golang.org/grpc",
				},
				{
					LicenseType: "Apache-2.0",
					Dependency:  "go.opencensus.io",
				},
			})
	}

	// --- Error tests ---
	{
		licenses, err := csvToLicenseInfos("")
		require.EqualError(t, err, "invalid CSV line (number of csv parts != 3) : ")
		require.Equal(t, licenses, []analyzers.LicenseInfo{})
	}

	{
		licenses, err := csvToLicenseInfos(`one,two`)
		require.EqualError(t, err, "invalid CSV line (number of csv parts != 3) : one,two")
		require.Equal(t, licenses, []analyzers.LicenseInfo{})
	}
}
