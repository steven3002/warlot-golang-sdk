package devcli

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment keys for defaults.
const (
	EnvBaseURL     = "WARLOT_BASE_URL"
	EnvAPIKey      = "WARLOT_API_KEY"
	EnvHolderID    = "WARLOT_HOLDER"
	EnvProjectName = "WARLOT_PNAME"

	EnvTimeoutSec  = "WARLOT_TIMEOUT"         // seconds
	EnvRetries     = "WARLOT_RETRIES"         // int
	EnvBackoffInit = "WARLOT_BACKOFF_INIT_MS" // ms
	EnvBackoffMax  = "WARLOT_BACKOFF_MAX_MS"  // ms
)

// Reasonable defaults for production-grade operation.
const (
	DefaultTimeoutSec  = 90
	DefaultRetries     = 5
	DefaultBackoffInit = 1000 // ms
	DefaultBackoffMax  = 8000 // ms
)

// GlobalFlags captures CLI-wide settings and defaults.
type GlobalFlags struct {
	BaseURL     string
	APIKey      string
	HolderID    string
	ProjectName string

	Timeout     time.Duration
	Retries     int
	BackoffInit time.Duration
	BackoffMax  time.Duration
	Verbose     bool
}

// ParseGlobalFlagsArgs binds global flags to the provided FlagSet and parses args.
func ParseGlobalFlagsArgs(fs *flag.FlagSet, args []string) GlobalFlags {
	var g GlobalFlags

	// Defaults sourced from environment variables.
	defBase := getenvDefault(EnvBaseURL, "https://warlot-api.onrender.com")
	defKey := getenvDefault(EnvAPIKey, "")
	defHolder := getenvDefault(EnvHolderID, "")
	defPname := getenvDefault(EnvProjectName, "")

	defTO := time.Duration(atoiDefault(os.Getenv(EnvTimeoutSec), DefaultTimeoutSec)) * time.Second
	defRet := atoiDefault(os.Getenv(EnvRetries), DefaultRetries)
	defBInit := durMsDefault(os.Getenv(EnvBackoffInit), time.Duration(DefaultBackoffInit)*time.Millisecond)
	defBMax := durMsDefault(os.Getenv(EnvBackoffMax), time.Duration(DefaultBackoffMax)*time.Millisecond)

	fs.StringVar(&g.BaseURL, "base", defBase, "API base URL (env "+EnvBaseURL+")")
	fs.StringVar(&g.APIKey, "apikey", defKey, "API key (env "+EnvAPIKey+")")
	fs.StringVar(&g.HolderID, "holder", defHolder, "Holder ID (env "+EnvHolderID+")")
	fs.StringVar(&g.ProjectName, "pname", defPname, "Project name (env "+EnvProjectName+")")

	timeoutSec := fs.Int("timeout", int(defTO/time.Second), "Request timeout seconds (env "+EnvTimeoutSec+")")
	fs.IntVar(&g.Retries, "retries", defRet, "Max retries on 429/5xx (env "+EnvRetries+")")

	backoffInit := fs.Int("backoff-init", int(defBInit/time.Millisecond), "Initial backoff ms (env "+EnvBackoffInit+")")
	backoffMax := fs.Int("backoff-max", int(defBMax/time.Millisecond), "Max backoff ms (env "+EnvBackoffMax+")")

	fs.BoolVar(&g.Verbose, "v", false, "Verbose request/response logs (API key redacted)")

	// Parse now.
	fs.Parse(args)

	// Finalize computed durations.
	g.Timeout = time.Duration(*timeoutSec) * time.Second
	g.BackoffInit = time.Duration(*backoffInit) * time.Millisecond
	g.BackoffMax = time.Duration(*backoffMax) * time.Millisecond

	return g
}

// MustNonEmpty enforces required flag presence for better operator feedback.
func MustNonEmpty(val, name string) {
	if strings.TrimSpace(val) == "" {
		// Returning an error is less ergonomic for small commands; exiting is acceptable for CLI.
		// Errors are printed by the command runner for consistent formatting.
		panic("missing required " + name)
	}
}

// Helpers

func getenvDefault(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func atoiDefault(s string, d int) int {
	if s == "" {
		return d
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return d
	}
	return i
}

func durMsDefault(msStr string, d time.Duration) time.Duration {
	if msStr == "" {
		return d
	}
	ms, err := strconv.Atoi(msStr)
	if err != nil {
		return d
	}
	return time.Duration(ms) * time.Millisecond
}
