package config

// Entry describes a single configuration property.
type Entry struct {
	Section  string
	Key      string
	Default  string
	EnvVar   string
	IsSecret bool
}

// Registry is the authoritative list of all HyperFleet CLI config properties.
// Order determines display order within each section.
var Registry = []Entry{
	// hyperfleet
	{Section: "hyperfleet", Key: "api-url", Default: "http://localhost:8000", EnvVar: "HF_API_URL"},
	{Section: "hyperfleet", Key: "api-version", Default: "v1", EnvVar: "HF_API_VERSION"},
	{Section: "hyperfleet", Key: "token", EnvVar: "HF_TOKEN", IsSecret: true},
	{Section: "hyperfleet", Key: "context", EnvVar: "HF_KUBE_CONTEXT"},
	{Section: "hyperfleet", Key: "namespace", EnvVar: "HF_KUBE_NAMESPACE"},
	{Section: "hyperfleet", Key: "gcp-project", Default: "hcm-hyperfleet", EnvVar: "HF_GCP_PROJECT"},
	{Section: "hyperfleet", Key: "cluster-id", EnvVar: "HF_CLUSTER_ID"},
	{Section: "hyperfleet", Key: "cluster-name", EnvVar: "HF_CLUSTER_NAME"},
	{Section: "hyperfleet", Key: "nodepool-id", EnvVar: "HF_NODEPOOL_ID"},
	// maestro
	{Section: "maestro", Key: "maestro-consumer", Default: "cluster1", EnvVar: "HF_MAESTRO_CONSUMER"},
	{Section: "maestro", Key: "maestro-http-endpoint", Default: "http://localhost:8100", EnvVar: "HF_MAESTRO_HTTP_ENDPOINT"},
	{Section: "maestro", Key: "maestro-grpc-endpoint", Default: "localhost:8090", EnvVar: "HF_MAESTRO_GRPC_ENDPOINT"},
	{Section: "maestro", Key: "maestro-namespace", Default: "maestro", EnvVar: "HF_MAESTRO_NAMESPACE"},
	// portforward
	{Section: "portforward", Key: "pf-api-port", Default: "8000", EnvVar: "HF_PF_API_PORT"},
	{Section: "portforward", Key: "pf-pg-port", Default: "5432", EnvVar: "HF_PF_PG_PORT"},
	{Section: "portforward", Key: "pf-maestro-http-port", Default: "8100", EnvVar: "HF_PF_MAESTRO_HTTP_PORT"},
	{Section: "portforward", Key: "pf-maestro-http-remote-port", Default: "8000", EnvVar: "HF_PF_MAESTRO_HTTP_REMOTE_PORT"},
	{Section: "portforward", Key: "pf-maestro-grpc-port", Default: "8090", EnvVar: "HF_PF_MAESTRO_GRPC_PORT"},
	// database
	{Section: "database", Key: "db-host", Default: "localhost", EnvVar: "HF_DB_HOST"},
	{Section: "database", Key: "db-port", Default: "5432", EnvVar: "HF_DB_PORT"},
	{Section: "database", Key: "db-name", EnvVar: "HF_DB_NAME"},
	{Section: "database", Key: "db-user", EnvVar: "HF_DB_USER"},
	{Section: "database", Key: "db-password", EnvVar: "HF_DB_PASSWORD", IsSecret: true},
	// rabbitmq
	{Section: "rabbitmq", Key: "rabbitmq-host", Default: "localhost", EnvVar: "HF_RABBITMQ_HOST"},
	{Section: "rabbitmq", Key: "rabbitmq-mgmt-port", Default: "15672", EnvVar: "HF_RABBITMQ_MGMT_PORT"},
	{Section: "rabbitmq", Key: "rabbitmq-user", Default: "guest", EnvVar: "HF_RABBITMQ_USER"},
	{Section: "rabbitmq", Key: "rabbitmq-password", EnvVar: "HF_RABBITMQ_PASSWORD", IsSecret: true},
	{Section: "rabbitmq", Key: "rabbitmq-vhost", Default: "/", EnvVar: "HF_RABBITMQ_VHOST"},
	// registry
	{Section: "registry", Key: "registry", EnvVar: "HF_REGISTRY"},
}

// Sections returns the unique section names in registry order.
func Sections() []string {
	seen := map[string]bool{}
	out := []string{}
	for _, e := range Registry {
		if !seen[e.Section] {
			seen[e.Section] = true
			out = append(out, e.Section)
		}
	}
	return out
}

// EntriesForSection returns all registry entries belonging to section.
func EntriesForSection(section string) []Entry {
	var out []Entry
	for _, e := range Registry {
		if e.Section == section {
			out = append(out, e)
		}
	}
	return out
}

// LookupEntry returns the registry entry for the given key, or false.
func LookupEntry(key string) (Entry, bool) {
	for _, e := range Registry {
		if e.Key == key {
			return e, true
		}
	}
	return Entry{}, false
}
