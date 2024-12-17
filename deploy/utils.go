package deploy

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func patchDotEnv(path string, envs map[string]string) error {
	dotEnv, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dotEnv.Close()

	env, err := godotenv.Parse(dotEnv)
	if err != nil {
		return err
	}

	env = mergeMaps(env, envs)
	return godotenv.Write(env, path)
}

func mergeMaps(maps ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

func nameToContainerName(prefix string, name string) string {
	container_name := strings.ToLower(name)
	container_name = strings.ReplaceAll(container_name, " ", "-")
	return prefix + "-" + container_name
}
