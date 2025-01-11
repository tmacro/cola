package templates

import (
	"strings"
	"text/template"

	"github.com/tmacro/cola/pkg/config"
)

func tplJoin(sep string, s []string) string {
	return strings.Join(s, sep)
}

func renderTemplate[T any](tpl *template.Template, data T) (string, error) {
	buf := new(strings.Builder)
	err := tpl.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

var systemdContainerTpl = template.Must(
	template.New("container").
		Funcs(template.FuncMap{"join": tplJoin}).
		Parse(mustGetEmbeddedFile("systemd.container.tpl")))

var systemdMountTpl = template.Must(
	template.New("mount").
		Parse(mustGetEmbeddedFile("systemd.mount.tpl")))

var systemdNetworkTpl = template.Must(
	template.New("network").
		Parse(mustGetEmbeddedFile("systemd.network.tpl")))

var systemdVlanNetworkTpl = template.Must(
	template.New("vlanNetwork").
		Parse(mustGetEmbeddedFile("systemd.vlan.network.tpl")))

var systemdVlanNetDevTpl = template.Must(
	template.New("vlanNetDev").
		Parse(mustGetEmbeddedFile("systemd.vlan.netdev.tpl")))

var systemdTmpfileConfigTpl = template.Must(
	template.New("tmpfileConfig").
		Parse(mustGetEmbeddedFile("systemd.tmpfile.tpl")))

func SystemdContainer(container config.Container) (string, error) {
	return renderTemplate(systemdContainerTpl, container)
}

func SystemdMount(mount config.Mount) (string, error) {
	return renderTemplate(systemdMountTpl, mount)
}

type networkConfig struct {
	Type       string
	Name       string
	MACAddress string
	Addresses  []string
	Gateway    string
	DNS        string
	DHCP       bool
	VLANs      []config.VLAN
	Options    map[string]string
}

func SystemdNetwork(network config.Interface) (string, error) {
	addresses := []string{}
	if network.Address != "" {
		addresses = append(addresses, network.Address)
	}

	addresses = append(addresses, network.Addresses...)

	cfg := networkConfig{
		Name:       network.Name,
		MACAddress: network.MACAddress,
		Addresses:  addresses,
		Gateway:    network.Gateway,
		DNS:        network.DNS,
		DHCP:       network.DHCP,
		VLANs:      network.VLANs,
	}

	return renderTemplate(systemdNetworkTpl, cfg)
}

func SystemdVlanNetwork(vlan config.VLAN) (string, error) {
	cfg := networkConfig{
		Name:      vlan.Name,
		Addresses: []string{vlan.Address},
		Gateway:   vlan.Gateway,
		DNS:       vlan.DNS,
		DHCP:      vlan.DHCP,
	}

	return renderTemplate(systemdVlanNetworkTpl, cfg)
}

type netdevConfig struct {
	Kind string
	Name string
	ID   int
}

func SystemdVlanNetDev(vlan config.VLAN) (string, error) {
	cfg := netdevConfig{
		Name: vlan.Name,
		ID:   vlan.ID,
	}

	return renderTemplate(systemdVlanNetDevTpl, cfg)
}

type Tmpfile struct {
	Mode   string
	Target string
	Source string
}

func SystemdTmpfileConfig(files ...Tmpfile) (string, error) {
	return renderTemplate(systemdTmpfileConfigTpl, files)
}
