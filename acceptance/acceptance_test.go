package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/disaster-recovery-acceptance-tests/runner"
	"github.com/cloudfoundry-incubator/disaster-recovery-acceptance-tests/testcases"
	. "github.com/onsi/ginkgo"
)

const defaultTimeout = 15 * time.Minute

var _ = Describe("backing up Cloud Foundry", func() {
	var config runner.Config
	var testCases []runner.TestCase

	focusedSuiteName := os.Getenv("FOCUSED_SUITE_NAME")
	skipSuiteName := os.Getenv("SKIP_SUITE_NAME")
	testCases = runner.FilterTestCasesWithRegexes(testcases.OpenSourceTestCases(), skipSuiteName, focusedSuiteName)

	if os.Getenv("CONFIG") != "" {
		config = getConfigFromFile(os.Getenv("CONFIG"))
	} else {
		config = getConfigFromEnv()
	}

	runner.RunDisasterRecoveryAcceptanceTests(config, testCases)
})

func getConfigFromFile(path string) runner.Config {
	configFromFile, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprint(fmt.Sprintf("Could not load config from file: %s\n", path)))
	}

	conf := runner.Config{}
	err = json.Unmarshal(configFromFile, &conf)
	if err != nil {
		panic(fmt.Sprint("Could not unmarshal Config"))
	}

	if conf.Timeout == 0 {
		conf.Timeout = defaultTimeout
	}

	return conf
}

func getConfigFromEnv() runner.Config {
	boshConfig := runner.BoshConfig{
		BoshURL:          mustHaveEnv("BOSH_ENVIRONMENT"),
		BoshClient:       mustHaveEnv("BOSH_CLIENT"),
		BoshClientSecret: mustHaveEnv("BOSH_CLIENT_SECRET"),
		BoshCaCert:       mustHaveEnv("BOSH_CA_CERT"),
	}
	deploymentConfig := runner.CloudFoundryConfig{
		Name:          mustHaveEnv("CF_DEPLOYMENT_NAME"),
		ApiUrl:        mustHaveEnv("CF_API_URL"),
		AdminUsername: mustHaveEnv("CF_ADMIN_USERNAME"),
		AdminPassword: mustHaveEnv("CF_ADMIN_PASSWORD"),
	}

	deploymentConfig.NFSServiceName = os.Getenv("NFS_SERVICE_NAME")
	deploymentConfig.NFSPlanName = os.Getenv("NFS_PLAN_NAME")
	deploymentConfig.NFSBrokerUser = os.Getenv("NFS_BROKER_USER")
	deploymentConfig.NFSBrokerPassword = os.Getenv("NFS_BROKER_PASSWORD")
	deploymentConfig.NFSBrokerUrl = os.Getenv("NFS_BROKER_URL")

	var timeout time.Duration
	timeoutStr := os.Getenv("DEFAULT_TIMEOUT_MINS")
	if timeoutStr != "" {
		timeoutInt, err := strconv.Atoi(timeoutStr)
		if err != nil {
			panic(fmt.Sprint("DEFAULT_TIMEOUT_MINS, if set, must be an integer\n"))
		}
		timeout = time.Duration(timeoutInt) * time.Minute
	} else {
		timeout = defaultTimeout
	}

	deleteAndRedeployCF := os.Getenv("DELETE_AND_REDEPLOY_CF") == "true"

	return runner.Config{
		CloudFoundryConfig:  deploymentConfig,
		BoshConfig:          boshConfig,
		Timeout:             timeout,
		DeleteAndRedeployCF: deleteAndRedeployCF,
	}
}

func mustHaveEnv(keyname string) string {
	val := os.Getenv(keyname)
	if val == "" {
		panic(fmt.Sprintf("Env var %s not set\n", keyname))
	}
	return val
}
