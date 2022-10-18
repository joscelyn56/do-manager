package manager

import (
	"context"
	"github.com/digitalocean/godo"
	"sync"
)

func Initialize(digitalOceanToken string, registry string, deleteCount int) RegistryManager {
	registryManager := RegistryManager{
		client:      godo.NewFromToken(digitalOceanToken),
		registry:    registry,
		deleteCount: deleteCount,
	}

	return registryManager
}

func (registryManager RegistryManager) DeleteExtraTags(ctx context.Context, repositories []Repository, tags [][]RepositoryTag) int {
	deletedTags := 0

	for i := 0; i < len(tags); i++ {
		if int(repositories[i].TagCount) > registryManager.deleteCount {
			extraTags := tags[i][registryManager.deleteCount:]
			for j := 0; j < len(extraTags); j++ {
				tagDeleted, _ := registryManager.DeleteTag(ctx, extraTags[j].Repository, extraTags[j].Tag)
				if tagDeleted {
					deletedTags += 1
				}
			}
		}
	}

	return deletedTags
}

func (registryManager RegistryManager) DeleteTag(ctx context.Context, repository string, tagName string) (bool, error) {
	resp, err := registryManager.client.Registry.DeleteTag(ctx, registryManager.registry, repository, tagName)

	if err != nil {
		return false, err
	}

	if resp.Status == "204 No Content" {
		return true, nil
	}

	return false, nil
}

func (registryManager RegistryManager) GetActiveGarbageCollection(ctx context.Context) string {
	_, _, err := registryManager.client.Registry.GetGarbageCollection(ctx, registryManager.registry)

	if err != nil {
		return "Inactive"
	}

	return "Active"
}

func (registryManager RegistryManager) GetAllocatedSubscriptionMemory(ctx context.Context, subscriptionMemoryChannel chan float64, errorChannel chan error, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	registrySubscription, _, err := registryManager.client.Registry.GetSubscription(ctx)

	if err != nil {
		errorChannel <- err
		close(subscriptionMemoryChannel)
	}

	subscription := &registrySubscription
	tier := &((*subscription).Tier)

	subscriptionMemoryChannel <- float64((*tier).IncludedStorageBytes)

	close(subscriptionMemoryChannel)
}

func (registryManager RegistryManager) GetRepositories(ctx context.Context, repositoryChannel chan []Repository, errorChannel chan error, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	options := &godo.ListOptions{
		Page:    1,
		PerPage: 10,
	}

	var repositoryList []Repository

	for {
		repositories, resp, err := registryManager.client.Registry.ListRepositories(ctx, registryManager.registry, options)

		if err != nil {
			errorChannel <- err
			close(repositoryChannel)
		}

		for i := 0; i < len(repositories); i++ {
			repository := *repositories[i]
			repositoryList = append(repositoryList, Repository{
				repository.Name,
				repository.RegistryName,
				repository.TagCount})
		}

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			errorChannel <- err
			continue
		}

		// set the page we want for the next request
		options.Page = page + 1
	}

	repositoryChannel <- repositoryList

	close(repositoryChannel)
}

func (registryManager RegistryManager) GetRepositoryTags(ctx context.Context, repositories []Repository, totalSpaceUsed *float64, tagsChannel chan [][]RepositoryTag, errorChannel chan error, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	options := &godo.ListOptions{
		Page:    1,
		PerPage: 10,
	}

	var TagList [][]RepositoryTag

	for i := 0; i < len(repositories); i++ {
		var TagRepository []RepositoryTag

		for {
			tags, resp, err := registryManager.client.Registry.ListRepositoryTags(ctx, registryManager.registry, repositories[i].RegistryName, options)

			if err != nil {
				errorChannel <- err
				continue
			}

			for i := 0; i < len(tags); i++ {
				tag := *tags[i]
				*totalSpaceUsed = *totalSpaceUsed + float64(tag.CompressedSizeBytes)
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

			// if we are at the last page, break out the for loop
			if resp.Links == nil || resp.Links.IsLastPage() {
				break
			}

			page, err := resp.Links.CurrentPage()
			if err != nil {
				errorChannel <- err
				continue
			}

			// set the page we want for the next request
			options.Page = page + 1
		}

		TagList = append(TagList, TagRepository)

		options.Page = 1
	}

	tagsChannel <- TagList

	close(tagsChannel)
}

func (registryManager RegistryManager) StartGarbageCollection(ctx context.Context) (string, error) {
	request := &godo.StartGarbageCollectionRequest{
		Type: "untagged manifests and unreferenced blobs",
	}
	gc, _, err := registryManager.client.Registry.StartGarbageCollection(ctx, registryManager.registry, request)
	if err != nil {
		return "", err
	}

	return gc.Status, nil
}
