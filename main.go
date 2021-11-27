package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	hsconfig "github.com/philips-software/go-hsdp-api/config"
	"github.com/philips-software/go-hsdp-api/console"
	"github.com/philips-software/go-hsdp-api/console/docker"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("cp")
	viper.SetDefault("tags", "latest")
	viper.AutomaticEnv()

	ctx := context.Background()
	sourceLogin := viper.GetString("source_login")
	sourcePassword := viper.GetString("source_password")
	sourceRepo := viper.GetString("source_repo")
	sourceHost := viper.GetString("source_host")
	sourceNamespace := viper.GetString("source_namespace")
	if sourceLogin == "" || sourcePassword == "" || sourceHost == "" || sourceNamespace == "" {
		fmt.Printf("some required SOURCE details are missing\n")
		return
	}
	sourceRegion, err := hostRegion(sourceHost)
	if err != nil {
		fmt.Printf("SOURCE region error: %v\n", err)
	}

	tags := strings.Split(viper.GetString("tags"), ",")
	if len(tags) == 0 || tags[0] == "" {
		fmt.Printf("missing TAGS, expecting a comma separated list\n")
		return
	}

	destLogin := viper.GetString("dest_login")
	destPassword := viper.GetString("dest_password")
	destHost := viper.GetString("dest_host")
	destNamespace := viper.GetString("dest_namespace")
	if destNamespace == "" {
		destNamespace = sourceNamespace
	}
	if destLogin == "" || destPassword == "" || destHost == "" {
		fmt.Printf("some requires DEST details are missing\n")
	}

	if err := login(sourceHost, sourceLogin, sourcePassword); err != nil {
		fmt.Printf("failed to login to SOURCE: %v\n", err)
		return
	}
	sc, err := console.NewClient(nil, &console.Config{
		Region: *sourceRegion,
	})
	if err != nil {
		fmt.Printf("failed to provision console client: %v\n", err)
		return
	}
	err = sc.Login(sourceLogin, sourcePassword)
	if err != nil {
		fmt.Printf("failed to login to SOURCE console: %v\n", err)
		return
	}
	sourceDC, err := docker.NewClient(sc, &docker.Config{
		Region: *sourceRegion,
	})
	if err != nil {
		fmt.Printf("failed to provision docker client: %v\n", err)
		return
	}
	if err := login(destHost, destLogin, destPassword); err != nil {
		fmt.Printf("failed to login to DEST: %v\n", err)
		return
	}

	sourceNS, err := sourceDC.Namespaces.GetNamespaceByID(ctx, sourceNamespace)
	if err != nil {
		fmt.Printf("SOURCE namepace error: %v\n", err)
		return
	}

	var reposToSync []string
	if sourceRepo != "" {
		reposToSync = append(reposToSync, fmt.Sprintf("%s/%s", sourceNamespace, sourceRepo))
	} else { // All repos in namespace
		repos, err := sourceDC.Namespaces.GetRepositories(ctx, sourceNS.ID)
		if err != nil {
			fmt.Printf("error retrieving SOURCE repositories: %v\n", err)
			return
		}
		for _, r := range *repos {
			reposToSync = append(reposToSync, r.ID)
		}
	}

	for _, repoToSync := range reposToSync {
		tag, err := sourceDC.Repositories.GetLatestTag(ctx, repoToSync)
		if err != nil {
			fmt.Printf("skipping: no latest tag for '%s'\n", repoToSync)
			continue
		}
		tagName := ":latest"
		if tag.Name != "" {
			tagName = ":" + tag.Name
		}
		src := fmt.Sprintf("%s/%s%s", sourceHost, repoToSync, tagName)
		dest := fmt.Sprintf("%s/%s%s", destHost, repoToSync, tagName)
		fmt.Printf("copying: %s -> %s\n", src, dest)
		err = crane.Copy(src, dest)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}
	}
}

func hostRegion(host string) (*string, error) {
	sd, err := hsconfig.New()
	if err != nil {
		return nil, err
	}
	for _, region := range sd.Regions() {
		if strings.Contains(sd.Region(region).Service("docker-registry").Host, host) {
			return &region, nil
		}
	}
	return nil, fmt.Errorf("region for '%s' not found", host)
}

func login(serverAddress string, user string, password string) interface{} {
	cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return err
	}
	creds := cf.GetCredentialsStore(serverAddress)
	if serverAddress == name.DefaultRegistry {
		serverAddress = authn.DefaultAuthKey
	}
	if err := creds.Store(types.AuthConfig{
		ServerAddress: serverAddress,
		Username:      user,
		Password:      password,
	}); err != nil {
		return err
	}
	if err := cf.Save(); err != nil {
		return err
	}
	return nil
}
