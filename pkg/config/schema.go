package config

import "github.com/hashicorp/hcl/v2"

const (
	VariableTypeString string = "string"
	VariableTypeNumber string = "number"
	VariableTypeBool   string = "bool"
)

type PartialConfig struct {
	Variables []Variable `hcl:"variable,block"`
	Remain    hcl.Body   `hcl:",remain"`
}

type VariableFile struct {
	Body hcl.Body `hcl:",remain"`
}

type ApplianceConfig struct {
	System      *System     `hcl:"system,block"`
	Etcd        *Etcd       `hcl:"etcd,block"`
	Users       []User      `hcl:"user,block"`
	Extensions  []Extension `hcl:"extension,block"`
	Containers  []Container `hcl:"container,block"`
	Files       []File      `hcl:"file,block"`
	Directories []Directory `hcl:"directory,block"`
	Symlinks    []Symlink   `hcl:"symlink,block"`
	Mounts      []Mount     `hcl:"mount,block"`
	Interfaces  []Interface `hcl:"interface,block"`
	Services    []Service   `hcl:"service,block"`
	Variables   []Variable  `hcl:"variable,block"`
}

type System struct {
	Hostname           string   `hcl:"hostname"`
	Timezone           string   `hcl:"timezone,optional"`
	EnableTTYAutoLogin bool     `hcl:"enable_tty_auto_login,optional"`
	Updates            *Updates `hcl:"updates,block"`
	PowerProfile       string   `hcl:"power-profile,optional"`
}

type Updates struct {
	RebootStrategy string `hcl:"reboot_strategy"`
}

type User struct {
	Username          string   `hcl:"username,label"`
	Uid               int      `hcl:"uid,optional"`
	Groups            []string `hcl:"groups,optional"`
	HomeDir           string   `hcl:"home_dir,optional"`
	NoCreateHome      bool     `hcl:"no_create_home,optional"`
	Shell             string   `hcl:"shell,optional"`
	SSHAuthorizedKeys []string `hcl:"ssh_authorized_keys,optional"`
}

type Extension struct {
	Name      string `hcl:"name,label"`
	Version   string `hcl:"version"`
	Arch      string `hcl:"arch,optional"`
	BakeryUrl string `hcl:"bakery_url"`
}

type Container struct {
	Name    string   `hcl:"name,label"`
	Image   string   `hcl:"image"`
	Args    []string `hcl:"args,optional"`
	Volumes []Volume `hcl:"volume,block"`
	Restart string   `hcl:"restart,optional"`
	CapAdd  []string `hcl:"cap_add,optional"`
}

type Volume struct {
	Source string `hcl:"source"`
	Target string `hcl:"target,label"`
}

type File struct {
	Path       string `hcl:"path,label"`
	Owner      string `hcl:"owner,optional"`
	Group      string `hcl:"group,optional"`
	Mode       string `hcl:"mode"`
	Inline     string `hcl:"inline,optional"`
	SourcePath string `hcl:"source_path,optional"`
	URL        string `hcl:"url,optional"`
	Overwrite  bool   `hcl:"overwrite,optional"`
}

type Directory struct {
	Path  string `hcl:"path,label"`
	Owner string `hcl:"owner,optional"`
	Group string `hcl:"group,optional"`
	Mode  string `hcl:"mode"`
}

type Symlink struct {
	Path      string `hcl:"path,label"`
	Target    string `hcl:"target"`
	Owner     string `hcl:"owner,optional"`
	Group     string `hcl:"group,optional"`
	Overwrite bool   `hcl:"overwrite,optional"`
}

type Mount struct {
	MountPoint string `hcl:"mount_point,label"`
	Type       string `hcl:"type"`
	What       string `hcl:"what"`
	Where      string `hcl:"where"`
	Options    string `hcl:"options,optional"`
}

type Interface struct {
	Name       string   `hcl:"name,optional"`
	MACAddress string   `hcl:"mac_address,optional"`
	Gateway    string   `hcl:"gateway,optional"`
	Address    string   `hcl:"address,optional"`
	Addresses  []string `hcl:"addresses,optional"`
	DNS        string   `hcl:"dns,optional"`
	DHCP       bool     `hcl:"dhcp,optional"`
	VLANs      []VLAN   `hcl:"vlan,block"`
}

type VLAN struct {
	Name    string `hcl:"name,label"`
	ID      int    `hcl:"id"`
	Address string `hcl:"address,optional"`
	Gateway string `hcl:"gateway,optional"`
	DNS     string `hcl:"dns,optional"`
	DHCP    bool   `hcl:"dhcp,optional"`
}

type Service struct {
	Name       string   `hcl:"name,label"`
	Inline     string   `hcl:"inline,optional"`
	SourcePath string   `hcl:"source_path,optional"`
	Enabled    bool     `hcl:"enabled,optional"`
	DropIns    []DropIn `hcl:"drop_in,block"`
}

type DropIn struct {
	Name       string `hcl:"name,label"`
	Inline     string `hcl:"inline,optional"`
	SourcePath string `hcl:"source_path,optional"`
}

type Etcd struct {
	Name          string `hcl:"name"`
	Server        bool   `hcl:"server,optional"`
	Gateway       bool   `hcl:"gateway,optional"`
	ListenAddress string `hcl:"listen-address,optional"`
	InitialToken  string `hcl:"initial-token,optional"`
	Peers         []Peer `hcl:"peer,block"`
}

type Peer struct {
	Name    string `hcl:"name,label"`
	Address string `hcl:"address"`
	Port    int    `hcl:"port"`
}

type Variable struct {
	Name   string   `hcl:"name,label"`
	Type   string   `hcl:"type"`
	Remain hcl.Body `hcl:",remain"`
}
