package checks

var jsonOutputFlagNames = []string{"output", "format", "json", "o"}
var stdinFlagNames = []string{"from-file", "input", "stdin"}
var stdinHelpTerms = []string{"stdin", "standard input", "read from -", "pipe"}
var dataInputFlagNames = []string{"file", "data", "body", "content", "payload", "path"}

var confirmBypassFlagNames = []string{"yes", "force", "confirm", "assume-yes", "no-confirm", "non-interactive"}

var dryRunFlagNames = []string{"dry-run", "dryrun", "whatif", "simulate", "dry_run"}

var timeoutFlagNames = []string{"timeout", "request-timeout"}

var paginationFlagNames = []string{
	"page-size", "page", "per-page", "limit", "cursor",
	"offset", "page-all", "paginate", "all",
}

var retryFlagNames = []string{"retry", "retry-count", "retry-max", "retry-after", "max-retries"}
var retryHelpTerms = []string{"retry", "rate-limit", "rate limit", "throttle"}
var networkIndicatorFlags = []string{"url", "endpoint", "host", "server", "api-url", "base-url"}
var networkHelpTerms = []string{"http://", "https://", "api endpoint", "rest api", "graphql", "webhook"}
var filterFlagNames = []string{"fields", "select", "columns", "filter", "jq", "query", "field"}
var exitCodeHelpTerms = []string{"exit code", "exit status", "return code", "exit codes"}

var authTokenFlagNames = []string{
	"token",
	"api-key",
	"api-token",
	"access-token",
	"credentials",
}

var authRelatedTerms = []string{
	"auth",
	"login",
	"token",
	"credential",
}

var nonInteractiveAuthFlagNames = []string{
	"token",
	"with-token",
	"service-account",
	"api-key",
	"api-token",
	"access-token",
	"credentials",
}

var loginCommandNames = map[string]bool{
	"login":   true,
	"signin":  true,
	"sign-in": true,
}
