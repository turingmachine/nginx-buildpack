package integration_test

import (
	"flag"
	"fmt"
	"nginx/int2/cfapi"
	"nginx/int2/cfapi/setup"
	"os"

	"github.com/cloudfoundry/libbuildpack/cutlass"
)

var bpDir string
var cluster cfapi.Cluster

func init() {
	var err error
	var buildpackName, buildpackFile, buildpackVersion, clusterType string
	flag.StringVar(&buildpackName, "bpName", "nginx_buildpack", "name of buildpack to use")
	flag.StringVar(&buildpackFile, "bpFile", "", "location of buildpack file. (must include version flag)")
	flag.StringVar(&buildpackVersion, "bpVersion", "", "version to use (builds if empty)")
	flag.StringVar(&clusterType, "clusterType", "foundation", "cluster type to run against [foundation,pack,cflocal]")
	flag.BoolVar(&cutlass.Cached, "cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "64M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "64M", "default disk for pushed apps")
	flag.Parse()

	bpDir, cluster, err = setup.Suite(buildpackName, buildpackFile, buildpackVersion, clusterType)
	if err != nil {
		fmt.Printf("Error in SuiteSetup: %s\n", err)
		os.Exit(1)
	}
}
