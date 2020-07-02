# Snyk Onboarding

This project aims to ease the burden of importing the many [goof](https://github.com/snyk/goof) demo applications
used to evaluate the different language support of Snyk into the different SCMs supported. Additionally, it can be
used to keep your local goof environments, and their respective upstream repositories, up to date with the primary
repos maintained by Snyk. This project stops short of actually onboarding each application into Snyk itself which
must still be performed in the Snyk application.

#### Supported Repositories
* GitHub
* GitLab
* Bitbucket
* Azure DevOps

#### Goof Applications

To see a current list of the upstream goof applications to be cloned, review the `repolist.txt` file in the project
root. Additional repositories can be added to this list as long as they are publicly accessible.

## Requirements

Docker _or_ Golang: the application can run standalone or in the included container scope. See instructions further
below for information on how to run in either case.

## Getting Started

This application can be run via Docker and the included Dockerfile or by compiling the application and running it
directly. It is recommended to use Docker for portability, but the application is able to determine its scope
automatically.

Both steps start with cloning or downloading this repository.

#### Docker

Build the docker container. Be mindful to include the period at the end, it tells Docker what the current context of
this build is.

`docker build -t snyk-onboard .`

Run the Docker container. Specify the host directory where you want the `goof` repos to be cloned to:

`docker run -it --rm --user $(id -u):$(id -g) -v /tmp/repos:/repos snyk-onboard`

In this example above, all the repositories will be downloaded to `/tmp/repos`. You could also use any other path
or even variables or subshells to be more dynamic such as `~/Desktop/Snyk` or `$(PWD)/repos`. You might have to manually
create the folder first or else you might see permission problems from Docker.

### Golang

If you have Golang installed, you can run the application natively. Ensure you are running a newer versions of Golang
that is using Go Modules by default (`1.13` and up.)

```shell script
go get
go build -o snyk-onboard
./snyk-onboard
```

Running the application standalone will create a directory wherever you run it called `repos` that will contain all the
downloaded goof repositories.

## Configuration

When the application launches for the first time, you will be prompted to enter configuration values to be used
for the creation process of the remote repos. Before you run the application, it is recommended you prepare the
following information before running the app:

* GitHub username and a [personal access token](https://github.com/settings/tokens) configured for the `repo` scope
* GitLab username and a [personal access token](https://gitlab.com/profile/personal_access_tokens) configured for the `api` scope
* Bitbucket username and an [app password](https://bitbucket.org/account/settings/app-passwords/) configured for the `repositories` scope
* Azure DevOps organization name and a [personal access token](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate) configured for `Read, write, & manage` on the `Code` and `Project and Team` scopes

These tokens will be saved for future use in the `repos` folder in a file called `.config.yaml`. This file can be
removed to run through setup again or edited directly to alter values.