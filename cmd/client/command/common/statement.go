package common

import (
	"fmt"
	"os"
	"strings"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	"golang.org/x/term"
)

// ParseOptionalStatement returns one statement when actions or resources are set.
// Both must be provided together. If neither is set, provided is false.
func ParseOptionalStatement(effect string, actions, resources []string) (statements []policyTY.Statement, provided bool, err error) {
	if len(actions) == 0 && len(resources) == 0 {
		return nil, false, nil
	}
	if len(actions) == 0 || len(resources) == 0 {
		return nil, false, fmt.Errorf("--action and --resource must be used together")
	}
	normalized, err := ParseEffect(effect)
	if err != nil {
		return nil, false, err
	}
	return []policyTY.Statement{{
		Effect:    normalized,
		Actions:   actions,
		Resources: resources,
	}}, true, nil
}

func ParseEffect(effect string) (string, error) {
	effect = strings.TrimSpace(effect)
	if effect == "" {
		return policyTY.EffectAllow, nil
	}
	switch strings.ToLower(effect) {
	case "allow":
		return policyTY.EffectAllow, nil
	case "deny":
		return policyTY.EffectDeny, nil
	default:
		return "", fmt.Errorf("effect must be Allow or Deny")
	}
}

func PromptPassword() (string, error) {
	_, _ = fmt.Fprint(rootCmd.IOStreams.Out, "Password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, _ = fmt.Fprintln(rootCmd.IOStreams.Out)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}
