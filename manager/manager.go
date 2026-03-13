package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
	"sync"
)

func Initialize(digitalOceanToken string, registry string, deleteCount int, waitPeriod time.Duration) RegistryManager {
	registryManager := RegistryManager{
		client:      godo.NewFromToken(digitalOceanToken),
		registry:    registry,
		deleteCount: deleteCount,
		waitPeriod:  waitPeriod,
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

	options := &godo.TokenListOptions{}

	var repositoryList []Repository

	for {
		repositories, resp, err := registryManager.client.Registry.ListRepositoriesV2(ctx, registryManager.registry, options)

		if err != nil {
			fmt.Println(err)
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

// HasRecentPushActivity checks if any image was pushed within the given duration.
func (registryManager RegistryManager) HasRecentPushActivity(ctx context.Context, since time.Duration) (bool, error) {
	options := &godo.TokenListOptions{}

	repositories, _, err := registryManager.client.Registry.ListRepositoriesV2(ctx, registryManager.registry, options)
	if err != nil {
		return false, err
	}

	cutoff := time.Now().Add(-since)

	for _, repo := range repositories {
		tags, _, err := registryManager.client.Registry.ListRepositoryTags(ctx, registryManager.registry, repo.Name, &godo.ListOptions{Page: 1, PerPage: 1})
		if err != nil {
			return false, err
		}
		if len(tags) > 0 && tags[0].UpdatedAt.After(cutoff) {
			return true, nil
		}
	}

	return false, nil
}

// WaitForQuietPeriod waits until no new images have been pushed for the configured
// wait period. It polls at half the wait period interval. Returns true if it's safe
// to start GC, false if the context was cancelled.
func (registryManager RegistryManager) WaitForQuietPeriod(ctx context.Context) bool {
	if registryManager.waitPeriod == 0 {
		return true
	}

	pollInterval := registryManager.waitPeriod / 2
	if pollInterval < 30*time.Second {
		pollInterval = 30 * time.Second
	}

	fmt.Printf("Waiting for quiet period (%s with no new pushes) before starting garbage collection...\n", registryManager.waitPeriod)

	for {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(pollInterval):
			active, err := registryManager.HasRecentPushActivity(ctx, registryManager.waitPeriod)
			if err != nil {
				fmt.Printf("Error checking push activity: %v, retrying...\n", err)
				continue
			}
			if !active {
				fmt.Println("No recent push activity detected. Safe to start garbage collection.")
				return true
			}
			fmt.Println("Recent push activity detected, waiting...")
		}
	}
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
