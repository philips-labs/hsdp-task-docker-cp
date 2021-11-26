package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("cp")
	viper.SetDefault("tags", "latest")
	viper.AutomaticEnv()

	sourceLogin := viper.GetString("source_login")
	sourcePassword := viper.GetString("source_password")
	sourceRepo := viper.GetString("source_repo")
	sourceHost := viper.GetString("source_host")
	sourceNamespace := viper.GetString("source_namespace")
	if sourceLogin == "" || sourcePassword == "" || sourceHost == "" || sourceNamespace == "" {
		fmt.Printf("some SOURCE details are missing\n")
		return
	}

	tags := strings.Split(viper.GetString("tags"), ",")
	if len(tags) == 0 || tags[0] == "" {
		fmt.Printf("missing TAGS, expecting a comma separated list\n")
		return
	}

	destLogin := viper.GetString("dest_login")
	destPassword := viper.GetString("dest_password")
	destRepo := viper.GetString("dest_repo")
	destHost := viper.GetString("dest_host")
	destNamespace := viper.GetString("dest_namespace")
	if destLogin == "" || destPassword == "" || destHost == "" || destNamespace == "" {
		fmt.Printf("some DEST details are missing\n")
	}

	if err := login(sourceHost, sourceLogin, sourcePassword); err != nil {
		fmt.Printf("failed to login to SOURCE: %v\n", err)
		return
	}
	if err := login(destHost, destLogin, destPassword); err != nil {
		fmt.Printf("failed to login to DEST: %v\n", err)
		return
	}

	for _, tag := range tags {
		src := fmt.Sprintf("%s/%s/%s:%s", sourceHost, sourceNamespace, sourceRepo, tag)
		dest := fmt.Sprintf("%s/%s/%s:%s", destHost, destNamespace, destRepo, tag)
		err := crane.Copy(src, dest)
		if err != nil {
			fmt.Printf("'%s' --> '%s' error: %v\n", src, dest, err)
			return
		}
	}
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
