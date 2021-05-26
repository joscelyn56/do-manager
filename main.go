package main

import (
	"context"
	"do-manager/manager"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
)

func main() {
	ctx := context.TODO()
	digitalOceanToken, tokenFound := os.LookupEnv("DIGITALOCEANTOKEN")
	registry, registryFound := os.LookupEnv("REGISTRY")

	if !tokenFound {
		log.Fatal("DIGITALOCEAN TOKEN NOT SET")
	}

	if !registryFound {
		log.Fatal("REGISTRY NAME NOT SET")
	}

	totalSpaceUsed := new(float64)
	deletedTags := new(int)

	subscriptionMemoryChannel := make(chan float64)
	repositoryChannel := make(chan []manager.Repository)
	tagsChannel := make(chan [][]manager.RepositoryTag)
	errorChannel := make(chan error)

	registryManager := manager.Initialize(digitalOceanToken, registry, 2)

	waitGroup := new(sync.WaitGroup)

	waitGroup.Add(2)

	go registryManager.GetAllocatedSubscriptionMemory(ctx, subscriptionMemoryChannel, errorChannel, waitGroup)
	go registryManager.GetRepositories(ctx, repositoryChannel, errorChannel, waitGroup)

	subscriptionMemoryAllocated, repositories := <-subscriptionMemoryChannel, <-repositoryChannel

	go registryManager.GetRepositoryTags(ctx, repositories, totalSpaceUsed, tagsChannel, errorChannel)

	tags := <-tagsChannel

	percentageSpaceUsed := math.Ceil((*totalSpaceUsed / subscriptionMemoryAllocated) * 100)

	fmt.Printf("You have used over %.0f percent of allocated memory for the month\n", percentageSpaceUsed)

	if percentageSpaceUsed > float64(80) {
		waitGroup.Add(1)
		go registryManager.DeleteExtraTags(ctx, repositories, tags, deletedTags, waitGroup)
	}

	if *deletedTags > 1 {
		status, err := registryManager.StartGarbageCollection(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Your current garbage collection status is %s\n", status)
	}

	for err := range errorChannel {
		if err != nil {
			log.Fatal(err)
		}
	}

	waitGroup.Wait()
	close(errorChannel)
}
