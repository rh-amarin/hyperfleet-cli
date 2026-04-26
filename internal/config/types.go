package config

import "os"

// ── section structs ───────────────────────────────────────────────────────────

type HyperfleetConfig struct {
	APIURL     string `yaml:"api-url,omitempty"`
	APIVersion string `yaml:"api-version,omitempty"`
	Token      string `yaml:"token,omitempty"`
	GCPProject string `yaml:"gcp-project,omitempty"`
}

type KubernetesConfig struct {
	Context   string `yaml:"context,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

type MaestroConfig struct {
	Consumer     string `yaml:"consumer,omitempty"`
	HTTPEndpoint string `yaml:"http-endpoint,omitempty"`
	GRPCEndpoint string `yaml:"grpc-endpoint,omitempty"`
	Namespace    string `yaml:"namespace,omitempty"`
}

type PortForwardConfig struct {
	APIPort           int `yaml:"api-port,omitempty"`
	PGPort            int `yaml:"pg-port,omitempty"`
	MaestroHTTPPort   int `yaml:"maestro-http-port,omitempty"`
	MaestroHTTPRemote int `yaml:"maestro-http-remote-port,omitempty"`
	MaestroGRPCPort   int `yaml:"maestro-grpc-port,omitempty"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Name     string `yaml:"name,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host,omitempty"`
	MgmtPort int    `yaml:"mgmt-port,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	VHost    string `yaml:"vhost,omitempty"`
}

type RegistryConfig struct {
	Name  string `yaml:"name,omitempty"`
	Token string `yaml:"token,omitempty"`
}

// Config is the full static configuration (config.yaml).
type Config struct {
	Hyperfleet  HyperfleetConfig  `yaml:"hyperfleet,omitempty"`
	Kubernetes  KubernetesConfig  `yaml:"kubernetes,omitempty"`
	Maestro     MaestroConfig     `yaml:"maestro,omitempty"`
	PortForward PortForwardConfig `yaml:"port-forward,omitempty"`
	Database    DatabaseConfig    `yaml:"database,omitempty"`
	RabbitMQ    RabbitMQConfig    `yaml:"rabbitmq,omitempty"`
	Registry    RegistryConfig    `yaml:"registry,omitempty"`
}

// State is the active runtime state (state.yaml) — flat, top-level keys.
type State struct {
	ActiveEnvironment string `yaml:"active-environment"`
	ClusterID         string `yaml:"cluster-id"`
	ClusterName       string `yaml:"cluster-name"`
	NodePoolID        string `yaml:"nodepool-id"`
}

// Defaults returns a Config pre-populated with all built-in default values.
func Defaults() Config { return defaults() }

func defaults() Config {
	return Config{
		Hyperfleet: HyperfleetConfig{
			APIURL:     "http://localhost:8000",
			APIVersion: "v1",
			GCPProject: "hcm-hyperfleet",
		},
		Maestro: MaestroConfig{
			Consumer:     "cluster1",
			HTTPEndpoint: "http://localhost:8100",
			GRPCEndpoint: "localhost:8090",
			Namespace:    "maestro",
		},
		PortForward: PortForwardConfig{
			APIPort:           8000,
			PGPort:            5432,
			MaestroHTTPPort:   8100,
			MaestroHTTPRemote: 8000,
			MaestroGRPCPort:   8090,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "hyperfleet",
			Name:     "hyperfleet",
			Password: "foobar-bizz-buzz",
		},
		RabbitMQ: RabbitMQConfig{
			Host:     "rabbitmq",
			MgmtPort: 15672,
			User:     "guest",
			Password: "guest",
			VHost:    "/",
		},
		Registry: RegistryConfig{
			Name: os.Getenv("USER"),
		},
	}
}

// secretPaths lists the dotted paths of secret fields.
var secretPaths = map[string]bool{
	"hyperfleet.token":  true,
	"database.password": true,
	"rabbitmq.password": true,
	"registry.token":    true,
}

// IsSecret reports whether the dotted path is a secret field.
func IsSecret(path string) bool { return secretPaths[path] }
