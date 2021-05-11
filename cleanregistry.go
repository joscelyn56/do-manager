package main

import (
	"do-manager/registrymanager"
	"flag"
	"fmt"
	"math"
	"os"
	"sync"
	"time"
)

func main() {
	apiToken := flag.String("token", "", "Digitalocean API token")
	registryName := flag.String("registry", "", "Digitalocean container registry name")
	count := flag.Int("count", 3, "Minimum number of tags allowed in the repository")
	flag.Parse()

	if *apiToken == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *registryName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	totalSpaceUsed := new(float64)
	deletedTags := new(int)

	subscriptionMemoryChannel := make(chan float64)
	repositoryChannel := make(chan []registrymanager.Repository)
	tagsChannel := make(chan [][]registrymanager.RepositoryTag)

	fmt.Println(time.Now())

	manager := registrymanager.Initialize(*apiToken, *registryName, *count)

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
