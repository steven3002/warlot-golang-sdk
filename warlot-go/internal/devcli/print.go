package devcli

import (
	"encoding/json"
	"fmt"
	"os"
)

// PrintJSON prints a value as pretty-printed JSON.
func PrintJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

// PrintGlobalUsage renders the top-level usage text.
func PrintGlobalUsage(bin string) {
	// Environment defaults echoed inline for transparency.
	fmt.Println(bin + ` - official CLI for the Warlot SQL API

USAGE:
  ` + bin + ` <command> [flags]

GLOBAL FLAGS (env defaults shown in []):
  -base          	API base URL [` + getenvDefault(EnvBaseURL, "https://warlot-api.onrender.com") + `]
  -apikey        	API key [` + getenvDefault(EnvAPIKey, "") + `]
  -holder        	Holder ID [` + getenvDefault(EnvHolderID, "") + `]
  -pname         	Project name [` + getenvDefault(EnvProjectName, "") + `]
  -timeout       	Request timeout seconds [` + getenvDefault(EnvTimeoutSec, "90") + `]
  -retries       	Retries on 429/5xx [` + getenvDefault(EnvRetries, "5") + `]
  -backoff-init  	Initial backoff ms [` + getenvDefault(EnvBackoffInit, "1000") + `]
  -backoff-max   	Max backoff ms [` + getenvDefault(EnvBackoffMax, "8000") + `]
  -v             	Verbose logs

COMMANDS:
  resolve             	                          Resolve project by holder + name
  init        		-owner 0x...                  Initialize new project
  issue-key   		-project <id> -user 0x...     Issue a project API key

  sql         		-project <id> -q "SQL ..." [-params '[...]' -idempotency key -stream]
  tables list   	-project <id>
  tables browse 	-project <id> -table products [-limit 10 -offset 0]
  tables schema 	-project <id> -table products
  tables count  	-project <id>
  status      		-project <id>
  commit      		-project <id>

EXAMPLES:
  ` + bin + ` resolve -holder 0xH -pname myproj
  ` + bin + ` init -holder 0xH -pname myproj -owner 0xUSER
  ` + bin + ` issue-key -project <id> -holder 0xH -pname myproj -user 0xUSER
  ` + bin + ` sql -project <id> -q 'SELECT * FROM products ORDER BY id DESC LIMIT 5'
  ` + bin + ` sql -project <id> -q 'INSERT INTO t (name) VALUES (?)' -params '["alice"]' -idempotency one
  ` + bin + ` tables browse -project <id> -table products -limit 10
`)
}

// Panicf is a small helper for required flag validation in subcommands.
func Panicf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	os.Exit(2)
}
