package testcases

import (
	"path"

	"strings"

	. "github.com/cloudfoundry-incubator/disaster-recovery-acceptance-tests/runner"
	. "github.com/onsi/gomega"
)

type CfCredhubSSITestCase struct {
	uniqueTestID  string
	name          string
	appName       string
	secondAppName string
	brokerName    string
	svcName       string
	svcInstance   string
}

func NewCfCredhubSSITestCase() *CfCredhubSSITestCase {
	id := RandomStringNumber()
	return &CfCredhubSSITestCase{
		uniqueTestID:  id,
		name:          "cf-credhub",
		svcName:       "service" + id,
		svcInstance:   "instance" + id,
		brokerName:    "broker" + id,
		appName:       "app" + id,
		secondAppName: "second-app" + id,
	}
}

func (tc *CfCredhubSSITestCase) Name() string {
	return tc.name
}

func (tc *CfCredhubSSITestCase) BeforeBackup(config Config) {
	RunCommandSuccessfully("cf api --skip-ssl-validation", config.CloudFoundryConfig.ApiUrl)
	RunCommandSuccessfully("cf auth", config.CloudFoundryConfig.AdminUsername, config.CloudFoundryConfig.AdminPassword)
	cmdResponse := RunCommandSuccessfully("cf running-environment-variable-group").Out.Contents()
	if !strings.Contains(string(cmdResponse), "CREDHUB_API") {
		RunCommandSuccessfully("cf set-running-environment-variable-group '{ \"CREDHUB_API\": \"https://credhub.service.cf.internal:8844\"}'")
	}
	RunCommandSuccessfully("cf create-org acceptance-test-org-" + tc.uniqueTestID)
	RunCommandSuccessfully("cf create-space acceptance-test-space-" + tc.uniqueTestID + " -o acceptance-test-org-" + tc.uniqueTestID)
	RunCommandSuccessfully("cf target -s acceptance-test-space-" + tc.uniqueTestID + " -o acceptance-test-org-" + tc.uniqueTestID)
	var testBrokerPath = path.Join(CurrentTestDir(), "/../fixtures/credhub_service_broker/")

	RunCommandSuccessfully("cf push " + tc.brokerName + " -p " + testBrokerPath + " -f " + testBrokerPath + "/manifest.yml" + " -b go_buildpack" + " -n " + tc.brokerName)
	RunCommandSuccessfully("cf set-env " + tc.brokerName + " SERVICE_NAME " + tc.svcName)
	RunCommandSuccessfully("cf restart " + tc.brokerName)

	serviceUrl := GetAppUrl(tc.brokerName)
	RunCommandSuccessfully("cf create-service-broker " + tc.brokerName + " " + config.CloudFoundryConfig.AdminUsername + " " + config.CloudFoundryConfig.AdminPassword + " https://" + serviceUrl)
	RunCommandSuccessfully("cf enable-service-access " + tc.svcName + " -o " + "acceptance-test-org-" + tc.uniqueTestID)
	RunCommandSuccessfully("cf create-service " + tc.svcName + " credhub-read-plan " + tc.svcInstance)

	var testAppPath = path.Join(CurrentTestDir(), "/../fixtures/credhub_enabled_app/credhub-enabled-app.jar")
	RunCommandSuccessfully("cf push " + tc.appName + " -p " + testAppPath + " --no-start" + " -n " + tc.appName)
	RunCommandSuccessfully("cf set-env " + tc.appName + " SERVICE_NAME " + tc.svcName)
	RunCommandSuccessfully("cf bind-service " + tc.appName + " " + tc.svcInstance)
	RunCommandSuccessfully("cf start " + tc.appName)

	appUrl := GetAppUrl(tc.appName)
	appResponse := RunCommandSuccessfully("curl", "-k", appUrl+"/test").Out.Contents()
	Expect(appResponse).To(ContainSubstring("pinkyPie"))
	Expect(appResponse).To(ContainSubstring("rainbowDash"))

	RunCommandSuccessfully("cf push " + tc.secondAppName + " -p " + testAppPath + " -n " + tc.secondAppName)
}

func (tc *CfCredhubSSITestCase) AfterBackup(config Config) {
	//do another bind that restore will clobber
	RunCommandSuccessfully("cf set-env " + tc.secondAppName + " SERVICE_NAME " + tc.svcName)
	RunCommandSuccessfully("cf bind-service " + tc.secondAppName + " " + tc.svcInstance)
	RunCommandSuccessfully("cf restart " + tc.secondAppName)

	secondAppUrl := GetAppUrl(tc.secondAppName)
	appResponse := RunCommandSuccessfully("curl", "-k", secondAppUrl+"/test").Out.Contents()
	Expect(appResponse).To(ContainSubstring("pinkyPie"))
	Expect(appResponse).To(ContainSubstring("rainbowDash"))
}

func (tc *CfCredhubSSITestCase) AfterRestore(config Config) {
	appUrl := GetAppUrl(tc.appName)
	appResponse := RunCommandSuccessfully("curl", "-k", appUrl+"/test").Out.Contents()
	Expect(appResponse).To(ContainSubstring("pinkyPie"))
	Expect(appResponse).To(ContainSubstring("rainbowDash"))

	secondAppUrl := GetAppUrl(tc.secondAppName)
	secondAppResponse := RunCommandSuccessfully("curl", "-k", secondAppUrl+"/test").Out.Contents()
	Expect(secondAppResponse).NotTo(ContainSubstring("pinkyPie"))
	Expect(secondAppResponse).NotTo(ContainSubstring("rainbowDash"))
}

func (tc *CfCredhubSSITestCase) Cleanup(config Config) {
	RunCommandSuccessfully("cf target -o acceptance-test-org-" + tc.uniqueTestID)
	RunCommandSuccessfully("cf delete -f " + tc.appName)
	RunCommandSuccessfully("cf delete -f " + tc.secondAppName)
	RunCommandSuccessfully("cf purge-service-instance -f " + tc.svcInstance)
	RunCommandSuccessfully("cf delete-service-broker -f " + tc.brokerName)
	RunCommandSuccessfully("cf delete-space -f acceptance-test-space-" + tc.uniqueTestID)
	RunCommandSuccessfully("cf delete-org -f acceptance-test-org-" + tc.uniqueTestID)
}
