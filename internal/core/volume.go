package core

// Volume describes one visible disk — internal, external/USB, mounted disk
// image, or network share — shown in the pinned rail (spec §4.3) alongside
// pinned folders. Ejectable is false for the boot/system disk, which can't
// be unmounted while running from it. IsNetwork splits already-mounted
// network shares (SMB/NFS/AFP/CIFS) into their own "Network" section of the
// rail, distinct from local/removable "Volumes" — it does not mean live
// network-neighborhood discovery of unmounted shares.
type Volume struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Ejectable  bool   `json:"ejectable"`
	TotalBytes uint64 `json:"totalBytes"`
	FreeBytes  uint64 `json:"freeBytes"`
	IsNetwork  bool   `json:"isNetwork"`
}

// VolumeLister lists currently mounted volumes, e.g. external drives that
// should surface the moment they're plugged in. Each OS provides its own
// implementation under internal/platform/<os>.
type VolumeLister interface {
	List() ([]Volume, error)
}

// VolumeEjector unmounts/ejects a volume by its mount path — the same
// operation as Finder's or Explorer's "Eject" on a USB drive.
type VolumeEjector interface {
	Eject(path string) error
}

// VolumeAdapter lists and ejects mounted volumes. Each OS provides one
// concrete implementation under internal/platform/<os> satisfying both
// narrower ports.
type VolumeAdapter interface {
	VolumeLister
	VolumeEjector
}
