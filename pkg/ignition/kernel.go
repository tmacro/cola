package ignition

import "github.com/tmacro/cola/pkg/config"

func generateKernelArguments(cfg *config.ApplianceConfig, g *generator) error {
	if cfg.System != nil && cfg.System.EnableTTYAutoLogin {
		g.KernelArguments.ShouldExist = append(g.KernelArguments.ShouldExist, "flatcar.autologin")
	} else {
		g.KernelArguments.ShouldNotExist = append(g.KernelArguments.ShouldNotExist, "flatcar.autologin")
	}

	return nil
}
