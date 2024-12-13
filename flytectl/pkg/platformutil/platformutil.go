package platformutil

type Arch string

const (
	ArchAmd64 Arch = "amd64"
	ArchX86   Arch = "x86_64"
	Arch386   Arch = "386"
	Archi386  Arch = "i386"
)

func (a Arch) String() string {
	return string(a)
}

type Platform string

const (
	Windows Platform = "windows"
	Linux   Platform = "linux"
	Darwin  Platform = "darwin"
)

func (p Platform) String() string {
	return string(p)
}
