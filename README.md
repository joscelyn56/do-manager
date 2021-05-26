# DO Manager
This is a tool that enables you automate digitalocean container registry cleanup to create more memory space.
It deletes old images from the registry, deletes the manifests for the images, and starts the garbage collection process to clear and reset the container registry available memory hence saving you extra expenses.

## Scenarios
>This is perfect if you push updates constantly to the digitalocean container registry, and you need to manage the registry garbage collection process automatically. This can be setup to run automatically as part of your CI/CD process.

## Pre-requisites
**Golang installation**

Ensure you have golang installed

**Give Bash Scripts Permission**

Give executable permission to the bash scripts
```bash
chmod 755 build.sh
chmod 755 run.sh
```

## Automated
**Build executable script**

run the command below to build the go program and creates an executable file.
<br>
```bash
./build.sh
```

**Run executable script**

run the command below to run the executable file.

```bash
./run.sh args1 args2 args3
```
> The arguments to be specified are defined below:\
**args1** - *digitalocean api token*\
**args2** - *container registry name*\
**args3** - *the minimum number of images to be left in your registry*

# Manual
<!-- blank line -->
>There are two ways to run this program manually, either via the main golang program file or via the cli golang program file.

**Via the main go file**

Export the following variables to your environment:
```bash
export DIGITALOCEANTOKEN={Digitalocean api token}
export REGISTRY={Digitalocean container registry name}
```

Run go program
```bash
go run main.go
```

**Via the CLI go file**

Navigate to CMD directory
```bash
cd cmd
```

Run go program
```bash
go run clean_registry.go -token {Digitalocean api token} -registry {Digitalocean container registry name} -count {the minimum number of images to be left in your registry}

```