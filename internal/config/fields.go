package config

import (
	"fmt"
	"strconv"
)

// GetField returns the string representation of the field at dotted path.
func GetField(cfg *Config, path string) (string, error) { return getField(cfg, path) }

// getField returns the string representation of the field at dotted path.
func getField(cfg *Config, path string) (string, error) {
	switch path {
	case "hyperfleet.api-url":
		return cfg.Hyperfleet.APIURL, nil
	case "hyperfleet.api-version":
		return cfg.Hyperfleet.APIVersion, nil
	case "hyperfleet.token":
		return cfg.Hyperfleet.Token, nil
	case "hyperfleet.gcp-project":
		return cfg.Hyperfleet.GCPProject, nil
	case "kubernetes.context":
		return cfg.Kubernetes.Context, nil
	case "kubernetes.namespace":
		return cfg.Kubernetes.Namespace, nil
	case "maestro.consumer":
		return cfg.Maestro.Consumer, nil
	case "maestro.http-endpoint":
		return cfg.Maestro.HTTPEndpoint, nil
	case "maestro.grpc-endpoint":
		return cfg.Maestro.GRPCEndpoint, nil
	case "maestro.namespace":
		return cfg.Maestro.Namespace, nil
	case "port-forward.api-port":
		return strconv.Itoa(cfg.PortForward.APIPort), nil
	case "port-forward.pg-port":
		return strconv.Itoa(cfg.PortForward.PGPort), nil
	case "port-forward.maestro-http-port":
		return strconv.Itoa(cfg.PortForward.MaestroHTTPPort), nil
	case "port-forward.maestro-http-remote-port":
		return strconv.Itoa(cfg.PortForward.MaestroHTTPRemote), nil
	case "port-forward.maestro-grpc-port":
		return strconv.Itoa(cfg.PortForward.MaestroGRPCPort), nil
	case "database.host":
		return cfg.Database.Host, nil
	case "database.port":
		return strconv.Itoa(cfg.Database.Port), nil
	case "database.name":
		return cfg.Database.Name, nil
	case "database.user":
		return cfg.Database.User, nil
	case "database.password":
		return cfg.Database.Password, nil
	case "rabbitmq.host":
		return cfg.RabbitMQ.Host, nil
	case "rabbitmq.mgmt-port":
		return strconv.Itoa(cfg.RabbitMQ.MgmtPort), nil
	case "rabbitmq.user":
		return cfg.RabbitMQ.User, nil
	case "rabbitmq.password":
		return cfg.RabbitMQ.Password, nil
	case "rabbitmq.vhost":
		return cfg.RabbitMQ.VHost, nil
	case "registry.name":
		return cfg.Registry.Name, nil
	}
	return "", fmt.Errorf("unknown config path %q", path)
}

// setField sets the field at dotted path to value (string → typed conversion).
func setField(cfg *Config, path, value string) error {
	toInt := func(s string) (int, error) { return strconv.Atoi(s) }
	switch path {
	case "hyperfleet.api-url":
		cfg.Hyperfleet.APIURL = value
	case "hyperfleet.api-version":
		cfg.Hyperfleet.APIVersion = value
	case "hyperfleet.token":
		cfg.Hyperfleet.Token = value
	case "hyperfleet.gcp-project":
		cfg.Hyperfleet.GCPProject = value
	case "kubernetes.context":
		cfg.Kubernetes.Context = value
	case "kubernetes.namespace":
		cfg.Kubernetes.Namespace = value
	case "maestro.consumer":
		cfg.Maestro.Consumer = value
	case "maestro.http-endpoint":
		cfg.Maestro.HTTPEndpoint = value
	case "maestro.grpc-endpoint":
		cfg.Maestro.GRPCEndpoint = value
	case "maestro.namespace":
		cfg.Maestro.Namespace = value
	case "port-forward.api-port":
		n, err := toInt(value)
		if err != nil {
			return err
		}
		cfg.PortForward.APIPort = n
	case "port-forward.pg-port":
		n, err := toInt(value)
		if err != nil {
			return err
		}
		cfg.PortForward.PGPort = n
	case "port-forward.maestro-http-port":
		n, err := toInt(value)
		if err != nil {
			return err
		}
		cfg.PortForward.MaestroHTTPPort = n
	case "port-forward.maestro-http-remote-port":
		n, err := toInt(value)
		if err != nil {
			return err
		}
		cfg.PortForward.MaestroHTTPRemote = n
	case "port-forward.maestro-grpc-port":
		n, err := toInt(value)
		if err != nil {
			return err
		}
		cfg.PortForward.MaestroGRPCPort = n
	case "database.host":
		cfg.Database.Host = value
	case "database.port":
		n, err := toInt(value)
		if err != nil {
			return err
		}
		cfg.Database.Port = n
	case "database.name":
		cfg.Database.Name = value
	case "database.user":
		cfg.Database.User = value
	case "database.password":
		cfg.Database.Password = value
	case "rabbitmq.host":
		cfg.RabbitMQ.Host = value
	case "rabbitmq.mgmt-port":
		n, err := toInt(value)
		if err != nil {
			return err
		}
		cfg.RabbitMQ.MgmtPort = n
	case "rabbitmq.user":
		cfg.RabbitMQ.User = value
	case "rabbitmq.password":
		cfg.RabbitMQ.Password = value
	case "rabbitmq.vhost":
		cfg.RabbitMQ.VHost = value
	case "registry.name":
		cfg.Registry.Name = value
	default:
		return fmt.Errorf("unknown config path %q", path)
	}
	return nil
}

// AllPaths returns every dotted path in declaration order, grouped by section.
var AllPaths = []struct {
	Section string
	Path    string
}{
	{"hyperfleet", "hyperfleet.api-url"},
	{"hyperfleet", "hyperfleet.api-version"},
	{"hyperfleet", "hyperfleet.token"},
	{"hyperfleet", "hyperfleet.gcp-project"},
	{"kubernetes", "kubernetes.context"},
	{"kubernetes", "kubernetes.namespace"},
	{"maestro", "maestro.consumer"},
	{"maestro", "maestro.http-endpoint"},
	{"maestro", "maestro.grpc-endpoint"},
	{"maestro", "maestro.namespace"},
	{"port-forward", "port-forward.api-port"},
	{"port-forward", "port-forward.pg-port"},
	{"port-forward", "port-forward.maestro-http-port"},
	{"port-forward", "port-forward.maestro-http-remote-port"},
	{"port-forward", "port-forward.maestro-grpc-port"},
	{"database", "database.host"},
	{"database", "database.port"},
	{"database", "database.name"},
	{"database", "database.user"},
	{"database", "database.password"},
	{"rabbitmq", "rabbitmq.host"},
	{"rabbitmq", "rabbitmq.mgmt-port"},
	{"rabbitmq", "rabbitmq.user"},
	{"rabbitmq", "rabbitmq.password"},
	{"rabbitmq", "rabbitmq.vhost"},
	{"registry", "registry.name"},
}
