package analyzers

type LicenseInfo struct {
	LicenseType string
	Dependency  string
}

type RepositoryLicenseInfos struct {
	RepositoryURL string
	Licenses      []LicenseInfo
}
