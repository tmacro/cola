package ignition

import (
	"fmt"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/pkg/config"
)

func toGroup(groups []string) []ignitionTypes.Group {
	if len(groups) == 0 {
		return nil
	}

	ignGroups := make([]ignitionTypes.Group, len(groups))
	for i, group := range groups {
		ignGroups[i] = ignitionTypes.Group(group)
	}

	return ignGroups
}

func toSSHAuthorizedKeys(keys []string) []ignitionTypes.SSHAuthorizedKey {
	if len(keys) == 0 {
		return nil
	}

	ignKeys := make([]ignitionTypes.SSHAuthorizedKey, len(keys))
	for i, key := range keys {
		ignKeys[i] = ignitionTypes.SSHAuthorizedKey(key)
	}

	return ignKeys
}

func generateUsers(cfg *config.ApplianceConfig, g *generator) error {
	for _, user := range cfg.Users {
		ignUser := ignitionTypes.PasswdUser{
			Name:              user.Username,
			Groups:            toGroup(user.Groups),
			SSHAuthorizedKeys: toSSHAuthorizedKeys(user.SSHAuthorizedKeys),
		}

		if user.Uid != 0 {
			ignUser.UID = toPtr(user.Uid)
		}

		if user.NoCreateHome {
			ignUser.NoCreateHome = toPtr(true)
		}

		if user.HomeDir != "" {
			ignUser.HomeDir = toPtr(user.HomeDir)
		}

		if user.Shell != "" {
			ignUser.Shell = toPtr(user.Shell)
		}

		g.Users = append(g.Users, ignUser)
	}

	return nil
}

func validateUsers(g *generator) error {
	users := make(map[string]struct{})
	for _, user := range g.Users {
		if _, ok := users[user.Name]; ok {
			return fmt.Errorf("%w: %s", ErrDuplicateUser, user.Name)
		}
		users[user.Name] = struct{}{}
	}

	return nil
}
