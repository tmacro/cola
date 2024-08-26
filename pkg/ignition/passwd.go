package ignition

import (
	"errors"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/pkg/config"
)

var (
	ErrDuplicateUser = errors.New("duplicate user")
)

func generateUsers(cfg *config.ApplianceConfig, g *generator) error {
	for _, user := range cfg.Users {
		ignUser := ignitionTypes.PasswdUser{
			Name:              user.Username,
			Groups:            toGroup(user.Groups),
			SSHAuthorizedKeys: toSSHAuthorizedKeys(user.SSHAuthorizedKeys),
			HomeDir:           toPtr(user.HomeDir),
			Shell:             toPtr(user.Shell),
		}

		g.Users = append(g.Users, ignUser)
	}

	return nil
}

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
