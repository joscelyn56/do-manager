package registrymanager

import (
	"context"
	_ "context"
	"github.com/digitalocean/godo"
	_ "github.com/digitalocean/godo"
	"time"
)

type RegistryManager struct {
	client *godo.Client
	ctx context.Context
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