package manager

import (
	"context"
	_ "context"
	"fmt"
	"github.com/digitalocean/godo"
	_ "github.com/digitalocean/godo"
	"sync"
)

func Initialize(digitalOceanToken string, registry string, deleteCount int) RegistryManager {
	registryManager := RegistryManager{
		client: godo.NewFromToken(digitalOceanToken),
		ctx: context.TODO(),
		registry: registry,
		deleteCount: deleteCount,
	}

	return registryManager
}

func (registryManager RegistryManager) DeleteExtraTags(repositories []Repository, tags [][]RepositoryTag, deletedTags *int, waitGroup *sync.WaitGroup)  {
	defer waitGroup.Done()

	for i := 0; i < len(tags); i++ {
		if int(repositories[i].TagCount) > registryManager.deleteCount {
			extraTags := tags[i][registryManager.deleteCount:]
			for j := 0; j < len(extraTags); j++ {
				tagDeleted := registryManager.DeleteTag(extraTags[j].Repository, extraTags[j].Tag)
				if tagDeleted {
					*deletedTags += 1
				}
			}
		}
	}
}

func (registryManager RegistryManager) DeleteTag(repository string, tagName string) bool {
	resp, err := registryManager.client.Registry.DeleteTag(registryManager.ctx, registryManager.registry, repository, tagName)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if resp.Status == "204 No Content" {
		return true
	}
	return false
}

func (registryManager RegistryManager) GetAllGarbageCollection() {
	options := &godo.ListOptions{
		Page:    1,
		PerPage: 10,
	}

	gc, _, err := registryManager.client.Registry.ListGarbageCollections(registryManager.ctx, registryManager.registry, options)

	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < len(gc); i++ {
		collection := *gc[i]
		fmt.Println("Collection", collection.Status)
	}
}

func (registryManager RegistryManager) GetAllocatedSubscriptionMemory(subscriptionMemoryChannel chan float64, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	registrySubscription, _, err := registryManager.client.Registry.GetSubscription(registryManager.ctx)

	if err != nil {
		fmt.Println(err)
		subscriptionMemoryChannel <- 0
	}

	subscription := &registrySubscription
	tier := &((*subscription).Tier)

	subscriptionMemoryChannel <- float64((*tier).IncludedStorageBytes)

	close(subscriptionMemoryChannel)
}

func (registryManager RegistryManager) GetRepositories(repositoryChannel chan []Repository, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	options := &godo.ListOptions{
		Page:    1,
		PerPage: 10,
	}

	repositories, _, err := registryManager.client.Registry.ListRepositories(registryManager.ctx, registryManager.registry, options)

	if err != nil {
		fmt.Println(err)
		repositoryChannel <- nil
	}

	var repositoryList []Repository
	for i := 0; i < len(repositories); i++ {
		repository := *repositories[i]
		repositoryList = append(repositoryList, Repository{
			repository.Name,
			repository.RegistryName,
			repository.TagCount})
	}

	repositoryChannel <- repositoryList

	close(repositoryChannel)
}

func (registryManager RegistryManager) GetRepositoryTags(repositories []Repository, totalSpaceUsed *float64, tagsChannel chan [][]RepositoryTag) {
	options := &godo.ListOptions{
		Page:    1,
		PerPage: 10,
	}

	var TagList [][]RepositoryTag

	for i := 0; i < len(repositories); i++ {
		var TagRepository []RepositoryTag

		tags, _, err := registryManager.client.Registry.ListRepositoryTags(registryManager.ctx, registryManager.registry, repositories[i].RegistryName, options)

		if err != nil {
			fmt.Println(err)
			continue
		}

		for i := 0; i < len(tags); i++ {
			tag := *tags[i]
			*totalSpaceUsed = *totalSpaceUsed + float64(tag.SizeBytes)
			TagRepository = append(TagRepository, RepositoryTag{
				tag.RegistryName,
				tag.Repository,
				tag.Tag,
				tag.ManifestDigest,
				tag.CompressedSizeBytes,
				tag.SizeBytes,
				tag.UpdatedAt,
			})
		}
		TagList = append(TagList, TagRepository)
	}

	tagsChannel <- TagList
}

func (registryManager RegistryManager) StartGarbageCollection() string {
	request := &godo.StartGarbageCollectionRequest{
		Type: "untagged manifests and unreferenced blobs",
	}
	gc, _, err := registryManager.client.Registry.StartGarbageCollection(registryManager.ctx, registryManager.registry, request)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return gc.Status
}