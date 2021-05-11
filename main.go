package main

import (
	"do-manager/manager"
	"fmt"
	"math"
	"os"
	"sync"
	"time"
)

func main() {
	digitalOceanToken, tokenFound := os.LookupEnv("DIGITALOCEANTOKEN")
	registry, registryFound := os.LookupEnv("REGISTRY")

	if !tokenFound {
		fmt.Println("DIGITALOCEAN TOKEN NOT SET")
		os.Exit(1)
	}

	if !registryFound {
		fmt.Println("REGISTRY NAME NOT SET")
		os.Exit(1)
	}

	totalSpaceUsed := new(float64)
	deletedTags := new(int)

	subscriptionMemoryChannel := make(chan float64)
	repositoryChannel := make(chan []manager.Repository)
	tagsChannel := make(chan [][]manager.RepositoryTag)

	fmt.Println(time.Now())

	manager := manager.Initialize(digitalOceanToken, registry, 2)

	waitGroup := new(sync.WaitGroup)

	waitGroup.Add(2)

	go manager.GetAllocatedSubscriptionMemory(subscriptionMemoryChannel, waitGroup)
	go manager.GetRepositories(repositoryChannel, waitGroup)

	subscriptionMemoryAllocated, repositories := <-subscriptionMemoryChannel, <-repositoryChannel

	go manager.GetRepositoryTags(repositories, totalSpaceUsed, tagsChannel)

	tags := <-tagsChannel

	percentageSpaceUsed := math.Ceil((*totalSpaceUsed / subscriptionMemoryAllocated) * 100)

	fmt.Printf("You have used over %.0f percent of allocated memory for the month\n", percentageSpaceUsed)

	if percentageSpaceUsed > float64(80) {
		waitGroup.Add(1)
		go manager.DeleteExtraTags(repositories, tags, deletedTags, waitGroup)
	}

	if *deletedTags > 1 {
		status := manager.StartGarbageCollection()
		fmt.Printf("Your current garbage collection status is %s\n", status)
	}

	waitGroup.Wait()

	fmt.Println(time.Now())
}
