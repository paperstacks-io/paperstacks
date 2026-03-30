package build

var (
	// These variables are replaced by ldflags at build time
	Version   = "development"
	GitHash   = ""
	BuildTime = "1970-01-01T00:00:00Z" // build date in ISO8601 format
)
