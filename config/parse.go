package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type StaticBootstrap struct {
	Admin struct {
		Address struct {
			SocketAddress struct {
				Address   string `yaml:"address"`
				PortValue int    `yaml:"port_value"`
			} `yaml:"socket_address"`
		} `yaml:"address"`
	} `yaml:"admin"`
	StaticResources struct {
		Listeners []struct {
			Name    string `yaml:"name"`
			Address struct {
				SocketAddress struct {
					Address   string `yaml:"address"`
					PortValue int    `yaml:"port_value"`
				} `yaml:"socket_address"`
			} `yaml:"address"`
			FilterChains []struct {
				Filters []struct {
					Name        string `yaml:"name"`
					TypedConfig struct {
						Type        string `yaml:"@type"`
						StatPrefix  string `yaml:"stat_prefix"`
						CodecType   string `yaml:"codec_type"`
						RouteConfig struct {
							Name         string `yaml:"name"`
							VirtualHosts []struct {
								Name    string   `yaml:"name"`
								Domains []string `yaml:"domains"`
								Routes  []struct {
									Match struct {
										Prefix string `yaml:"prefix"`
									} `yaml:"match"`
									Route struct {
										Cluster string `yaml:"cluster"`
									} `yaml:"route"`
								} `yaml:"routes"`
							} `yaml:"virtual_hosts"`
						} `yaml:"route_config"`
						HTTPFilters []struct {
							Name        string `yaml:"name"`
							TypedConfig struct {
								Type string `yaml:"@type"`
							} `yaml:"typed_config"`
						} `yaml:"http_filters"`
					} `yaml:"typed_config"`
				} `yaml:"filters"`
			} `yaml:"filter_chains"`
		} `yaml:"listeners"`
		Clusters []struct {
			Name           string `yaml:"name"`
			ConnectTimeout string `yaml:"connect_timeout"`
			Type           string `yaml:"type"`
			LbPolicy       string `yaml:"lb_policy"`
			LoadAssignment struct {
				ClusterName string `yaml:"cluster_name"`
				Endpoints   []struct {
					LbEndpoints []struct {
						Endpoint struct {
							Address struct {
								SocketAddress struct {
									Address   string `yaml:"address"`
									PortValue int    `yaml:"port_value"`
								} `yaml:"socket_address"`
							} `yaml:"address"`
						} `yaml:"endpoint"`
					} `yaml:"lb_endpoints"`
				} `yaml:"endpoints"`
			} `yaml:"load_assignment"`
		} `yaml:"clusters"`
	} `yaml:"static_resources"`
}

type BackendServer struct {
	Address string
	Port    int
}

func GetYAMLdata() (StaticBootstrap, []BackendServer) {
	var staticBootstrap StaticBootstrap
	var backendServers []BackendServer

	staticData, err := os.ReadFile("config/static.yaml")
	if err != nil {
		fmt.Println("Error reading 'static.yaml'")
	} else {
		err = yaml.Unmarshal(staticData, &staticBootstrap)
		if err != nil {
			fmt.Println("Error unmarshaling 'listeners.yaml'")
		}

		// Extract backend server information from the configuration
		for _, cluster := range staticBootstrap.StaticResources.Clusters {
			for _, endpoint := range cluster.LoadAssignment.Endpoints {
				for _, lbEndpoint := range endpoint.LbEndpoints {
					server := BackendServer{
						Address: lbEndpoint.Endpoint.Address.SocketAddress.Address,
						Port:    int(lbEndpoint.Endpoint.Address.SocketAddress.PortValue),
					}
					backendServers = append(backendServers, server)
				}
			}
		}
	}

	return staticBootstrap, backendServers
}
