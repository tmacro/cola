package main

type ApplianceConfig struct {
	System      System      `hcl:"system,block"`
	Users       []User      `hcl:"user,block"`
	Extensions  []Extension `hcl:"extension,block"`
	Containers  []Container `hcl:"container,block"`
	Files       []File      `hcl:"file,block"`
	Directories []Directory `hcl:"directory,block"`
	Mounts      []Mount     `hcl:"mount,block"`
}

type System struct {
	Hostname string `hcl:"hostname"`
	Timezone string `hcl:"timezone,optional"`
}

type User struct {
	Username          string   `hcl:"username,label"`
	Groups            []string `hcl:"groups,optional"`
	HomeDir           string   `hcl:"home_dir,optional"`
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
}

type Volume struct {
	Source string `hcl:"source,label"`
	Target string `hcl:"target"`
}

type File struct {
	Path       string `hcl:"path,label"`
	Owner      string `hcl:"owner,optional"`
	Group      string `hcl:"group,optional"`
	Mode       string `hcl:"mode"`
	Inline     string `hcl:"inline,optional"`
	SourcePath string `hcl:"source_path,optional"`
	URL        string `hcl:"url,optional"`
}

type Directory struct {
	Path  string `hcl:"path,label"`
	Owner string `hcl:"owner,optional"`
	Group string `hcl:"group,optional"`
	Mode  string `hcl:"mode"`
}

type Mount struct {
	MountPoint string `hcl:"mount_point,label"`
	Type       string `hcl:"type"`
	What       string `hcl:"what"`
	Where      string `hcl:"where"`
	Options    string `hcl:"options,optional"`
}
