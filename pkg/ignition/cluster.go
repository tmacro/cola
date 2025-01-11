package ignition

import (
	"errors"
	"strconv"
	"strings"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/internal/files"
	"github.com/tmacro/cola/pkg/config"
)

func encodeEnvVars(vars map[string]string) string {
	var envVars []string
	for k, v := range vars {
		envVars = append(envVars, k+"=\""+v+"\"")
	}
	return strings.Join(envVars, "\n")
}

func generateEtcdConfig(cfg *config.ApplianceConfig, g *generator) error {
	if cfg.Etcd == nil {
		return nil
	}

	dropins := []ignitionTypes.Dropin{}
	env := make(map[string]string)

	if cfg.Etcd.Server {
		dropins = append(dropins, ignitionTypes.Dropin{
			Name:     "20-member.conf",
			Contents: toPtr(files.MustGetEmbeddedFile("etcd.member.conf")),
		})

		env["ETCD_NAME"] = cfg.Etcd.Name
		env["LISTEN_ADDR"] = cfg.Etcd.ListenAddress
		env["INITIAL_CLUSTER_TOKEN"] = cfg.Etcd.InitialToken
		peers := []string{
			cfg.Etcd.Name + "=http://" + cfg.Etcd.ListenAddress + ":2380",
		}

		for _, peer := range cfg.Etcd.Peers {
			peers = append(peers, peer.Name+"=http://"+peer.Address+":"+strconv.Itoa(peer.Port))
		}

		env["INITIAL_CLUSTER"] = strings.Join(peers, ",")
	}

	if cfg.Etcd.Gateway {
		dropins = append(dropins, ignitionTypes.Dropin{
			Name:     "20-gateway.conf",
			Contents: toPtr(files.MustGetEmbeddedFile("etcd.gateway.conf")),
		})

		endpoints := make([]string, 0, len(cfg.Etcd.Peers))

		for _, peer := range cfg.Etcd.Peers {
			endpoints = append(endpoints, peer.Address+":"+strconv.Itoa(peer.Port))
		}

		env["CLUSTER_ENDPOINTS"] = strings.Join(endpoints, ",")
	}

	if len(dropins) == 0 {
		return errors.New("no etcd configuration specified")
	}

	g.Units = append(g.Units, ignitionTypes.Unit{
		Enabled: toPtr(true),
		Name:    "etcd-member.service",
		Dropins: dropins,
	})

	g.Files = append(g.Files, ignitionTypes.File{
		Node: ignitionTypes.Node{
			Path:      "/etc/default/etcd",
			Overwrite: toPtr(true),
		},
		FileEmbedded1: ignitionTypes.FileEmbedded1{
			Mode: toPtr(0640),
			Contents: ignitionTypes.Resource{
				Source: toPtr(toDataUrl(encodeEnvVars(env))),
			},
		},
	})

	return nil
}
