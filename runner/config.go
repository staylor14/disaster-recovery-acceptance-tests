package runner

import "time"

type CloudFoundryConfig struct {
	Name              string `json:"cf_deployment_name"`
	ApiUrl            string `json:"cf_api_url"`
	AdminUsername     string `json:"cf_admin_username"`
	AdminPassword     string `json:"cf_admin_password"`
	NFSServiceName    string `json:"nfs_service_name,omitempty"`
	NFSPlanName       string `json:"nfs_plan_name,omitempty"`
	NFSBrokerUser     string `json:"nfs_broker_user,omitempty"`
	NFSBrokerPassword string `json:"nfs_broker_password,omitempty"`
	NFSBrokerUrl      string `json:"nfs_broker_url,omitempty"`
}

type BoshConfig struct {
	BoshURL          string `json:"bosh_environment"`
	BoshClient       string `json:"bosh_client"`
	BoshClientSecret string `json:"bosh_client_secret"`
	BoshCaCert       string `json:"bosh_ca_cert"`
}

type Config struct {
	CloudFoundryConfig
	BoshConfig
	Timeout             time.Duration `json:"timeout"`
	DeleteAndRedeployCF bool          `json:"delete_and_redeploy_cf"`
}
