package manager

import (
	"github.com/digitalocean/godo"
	"time"
)

type RegistryManager struct {
	client *godo.Client
	registry string
	deleteCount int
}

type Repository struct {
	RegistryName string
	Name         string
	TagCount     uint64
}

type RepositoryTag struct {
	RegistryName        string
	Repository          string
	Tag                 string
	ManifestDigest      string
	CompressedSizeBytes uint64
	SizeBytes           uint64
	UpdatedAt           time.Time
}