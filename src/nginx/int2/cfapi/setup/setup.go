package setup

import (
	"fmt"
	"nginx/int2/cfapi"
	"nginx/int2/cfapi/cflocal"
	"nginx/int2/cfapi/foundation"
	"nginx/int2/cfapi/pack"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/onsi/gomega"
)

func Suite(buildpackName, buildpackFile, buildpackVersion, clusterType string) (string, cfapi.Cluster, error) {
	bpDir, err := cutlass.FindRoot()
	if err != nil {
		return "", nil, fmt.Errorf("Could not find buildpack root dir: %s", err)
	}

	if err := cutlass.CopyCfHome(); err != nil {
		return "", nil, fmt.Errorf("Could not copy cf home dir: %s", err)
	}
	cutlass.SeedRandom()
	gomega.SetDefaultEventuallyTimeout(10 * time.Second)

	var cluster cfapi.Cluster
	switch clusterType {
	case "foundation":
		cluster = foundation.NewCluster() // 83s (1m23s) & 81s
	case "pack":
		cluster = pack.NewCluster() // 18s & 16s
	case "cflocal":
		cluster = cflocal.NewCluster() // ??s & ??s
	default:
		return "", nil, fmt.Errorf("Unknown clusterType: %s. Available: foundation,pack,cflocal", clusterType)
	}

	if buildpackFile != "" {
		fmt.Println("Uploading Buildpack:", buildpackFile)
		if err := cluster.UploadBuildpack(buildpackName, buildpackVersion, buildpackFile); err != nil {
			return "", nil, fmt.Errorf("Could not upload default buildpack: %s", err)
		}
	} else if buildpackVersion != "" {
		fmt.Printf("Using buildpack %s version %s\n", buildpackName, buildpackVersion)
		if err := cluster.UploadBuildpack(buildpackName, buildpackVersion, ""); err != nil {
			return "", nil, fmt.Errorf("Could not upload default buildpack: %s", err)
		}
	} else {
		fmt.Println("Building Buildpack:", buildpackName)
		buildpack, err := cutlass.PackageUniquelyVersionedBuildpack()
		if err != nil {
			return "", nil, fmt.Errorf("Could not build buildpack: %s", err)
		}
		fmt.Println("Uploading Buildpack:", buildpack.File)
		if err := cluster.UploadBuildpack(buildpackName, buildpack.Version, buildpack.File); err != nil {
			return "", nil, fmt.Errorf("Could not upload default buildpack: %s", err)
		}
	}
	return bpDir, cluster, nil
}
