package checks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
)

// ---------------------------------------------------------------------------
// AU-1: Env var auth support
// ---------------------------------------------------------------------------

type checkAU1 struct {
	BaseCheck
}

func newCheckAU1() *checkAU1 {
	return &checkAU1{
		BaseCheck: BaseCheck{
			CheckID:             "AU-1",
			CheckName:           "Env var auth support",
			CheckCategory:       CatAuth,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Support authentication via environment variables for headless/agent usage.",
		},
	}
}

var envVarSuffixRe = regexp.MustCompile(`[A-Z][A-Z0-9_]*_(TOKEN|API_KEY|CREDENTIALS|SECRET|PASSWORD)\b`)

func hasAuthEnvVarMention(text string) bool {
	return envVarSuffixRe.MatchString(text)
}

func containsAuthTerm(text string) bool {
	lower := strings.ToLower(text)
	for _, term := range authRelatedTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

func (c *checkAU1) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	if f, cmd := idx.FindFlag(authTokenFlagNames...); f != nil {
		return PassResult(c, fmt.Sprintf("found auth flag --%s on command %q", f.Name, strings.Join(cmd.FullPath, " ")))
	}

	for _, cmd := range idx.All() {
		if hasAuthEnvVarMention(cmd.RawHelp) {
			match := envVarSuffixRe.FindString(cmd.RawHelp)
			return PassResult(c, fmt.Sprintf("found auth env var %s in help for %q", match, strings.Join(cmd.FullPath, " ")))
		}
	}

	for _, name := range []string{"auth", "login"} {
		for _, cmd := range idx.CommandsByName(name) {
			h := idx.LowerHelp(cmd)
			if strings.Contains(h, "token") || strings.Contains(h, "api_key") ||
				strings.Contains(h, "api-key") || strings.Contains(h, "env") {
				return PassResult(c, fmt.Sprintf("found token/env var reference in %q subcommand help", cmd.Name))
			}
		}
	}

	anyAuthMention := false
	if _, ok := idx.HelpContainsAny(authRelatedTerms...); ok {
		anyAuthMention = true
	}
	if !anyAuthMention {
		if len(idx.CommandsByName("auth")) > 0 || len(idx.CommandsByName("login")) > 0 {
			anyAuthMention = true
		}
	}

	if !anyAuthMention {
		return PassResult(c, "no auth-related commands or flags detected; auth not applicable")
	}

	return FailResult(c, "auth-related content found but no env var or token flag for non-interactive auth")
}

// ---------------------------------------------------------------------------
// AU-2: No mandatory interactive auth
// ---------------------------------------------------------------------------

type checkAU2 struct {
	BaseCheck
}

func newCheckAU2() *checkAU2 {
	return &checkAU2{
		BaseCheck: BaseCheck{
			CheckID:             "AU-2",
			CheckName:           "No mandatory interactive auth",
			CheckCategory:       CatAuth,
			CheckSeverity:       Fail,
			CheckMethod:         Passive,
			CheckRecommendation: "Provide non-interactive auth paths (API keys, service account files, token env vars) alongside interactive flows.",
		},
	}
}

func findLoginCommand(idx *discovery.CommandIndex) *discovery.Command {
	for name := range loginCommandNames {
		if cmds := idx.CommandsByName(name); len(cmds) > 0 {
			return cmds[0]
		}
	}
	return nil
}

func hasNonInteractiveAlternative(idx *discovery.CommandIndex) (bool, string) {
	if f, cmd := idx.FindFlag(nonInteractiveAuthFlagNames...); f != nil {
		return true, fmt.Sprintf("found non-interactive auth flag --%s on command %q", f.Name, strings.Join(cmd.FullPath, " "))
	}

	for _, cmd := range idx.All() {
		if hasAuthEnvVarMention(cmd.RawHelp) {
			match := envVarSuffixRe.FindString(cmd.RawHelp)
			return true, fmt.Sprintf("found auth env var %s in help for %q", match, strings.Join(cmd.FullPath, " "))
		}
	}
	return false, ""
}

func (c *checkAU2) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	loginCmd := findLoginCommand(idx)
	if loginCmd == nil {
		return PassResult(c, "no login/signin/sign-in command found; no mandatory interactive auth")
	}

	found, detail := hasNonInteractiveAlternative(idx)
	if found {
		return PassResult(c, fmt.Sprintf("login command %q exists but non-interactive alternative found: %s",
			strings.Join(loginCmd.FullPath, " "), detail))
	}

	return FailResult(c, fmt.Sprintf("login command %q exists with no non-interactive auth alternative (no token flags or auth env vars found)",
		strings.Join(loginCmd.FullPath, " ")))
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func registerAuthChecks(r *Registry) {
	r.Register(newCheckAU1())
	r.Register(newCheckAU2())
}
