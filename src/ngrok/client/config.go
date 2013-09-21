package client

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"net"
	"net/url"
	"ngrok/log"
	"os"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type Configuration struct {
	HttpProxy          string                         `yaml:"http_proxy"`
	ServerAddr         string                         `yaml:"server_addr"`
	InspectAddr        string                         `yaml:"inspect_addr"`
	TrustHostRootCerts bool                           `yaml:"trust_host_root_certs"`
	AuthToken          string                         `yaml:"auth_token"`
	Tunnels            map[string]TunnelConfiguration `yaml:"tunnels"`
	LogTo              string
}

type TunnelConfiguration struct {
	Subdomain string            `yaml:"subdomain"`
	Hostname  string            `yaml:"hostname"`
	Protocols map[string]string `yaml:"proto"`
	HttpAuth  string            `yaml:"auth"`
}

func LoadConfiguration(opts *Options) (config *Configuration, err error) {
	configPath := opts.config
	if configPath == "" {
		configPath = defaultPath()
	}

	log.Info("Reading configuration file %s", configPath)
	configBuf, err := ioutil.ReadFile(configPath)
	if err != nil {
		// failure to read a configuration file is only a fatal error if
		// the user specified one explicitly
		if opts.config != "" {
			err = fmt.Errorf("Failed to read configuration file %s: %v", configPath, err)
			return
		}
	}

	// deserialize/parse the config
	config = new(Configuration)
	if err = goyaml.Unmarshal(configBuf, &config); err != nil {
		err = fmt.Errorf("Error parsing configuration file %s: %v", configPath, err)
		return
	}

	// try to parse the old .ngrok format for backwards compatibility
	matched := false
	content := strings.TrimSpace(string(configBuf))
	if matched, err = regexp.MatchString("^[0-9a-zA-Z_\\-!]+$", content); err != nil {
		return
	} else if matched {
		config = &Configuration{AuthToken: content}
	}

	// set configuration defaults
	if config.ServerAddr == "" {
		config.ServerAddr = defaultServerAddr
	}

	if config.InspectAddr == "" {
		config.InspectAddr = "127.0.0.1:4040"
	}

	if config.HttpProxy == "" {
		config.HttpProxy = os.Getenv("http_proxy")
	}

	// override configuration with command-line options
	config.LogTo = opts.logto
	if opts.authtoken != "" {
		config.AuthToken = opts.authtoken
	}

	switch opts.command {
	// start a single tunnel, the default, simple ngrok behavior
	case "default":
		config.Tunnels = make(map[string]TunnelConfiguration)
		config.Tunnels["default"] = TunnelConfiguration{
			Subdomain: opts.subdomain,
			Hostname:  opts.hostname,
			HttpAuth:  opts.httpauth,
			Protocols: make(map[string]string),
		}

		for _, proto := range strings.Split(opts.protocol, "+") {
			config.Tunnels["default"].Protocols[proto] = opts.args[0]
		}

	// start tunnels
	case "start":
		if len(opts.args) == 0 {
			err = fmt.Errorf("You must specify at least one tunnel to start")
			return
		}

		requestedTunnels := make(map[string]bool)
		for _, arg := range opts.args {
			requestedTunnels[arg] = true

			if _, ok := config.Tunnels[arg]; !ok {
				err = fmt.Errorf("Requested to start tunnel %s which is not defined in the config file.", arg)
				return
			}
		}

		for name, _ := range config.Tunnels {
			if !requestedTunnels[name] {
				delete(config.Tunnels, name)
			}
		}

	default:
		err = fmt.Errorf("Unknown command: %s", opts.command)
		return
	}

	// validate and normalize configuration
	if config.InspectAddr, err = normalizeAddress(config.InspectAddr, "inspect_addr"); err != nil {
		return
	}

	if config.ServerAddr, err = normalizeAddress(config.ServerAddr, "server_addr"); err != nil {
		return
	}

	if config.HttpProxy != "" {
		var proxyUrl *url.URL
		if proxyUrl, err = url.Parse(config.HttpProxy); err != nil {
			return
		} else {
			if proxyUrl.Scheme != "http" && proxyUrl.Scheme != "https" {
				err = fmt.Errorf("Proxy url scheme must be 'http' or 'https', got %v", proxyUrl.Scheme)
				return
			}
		}
	}

	for name, t := range config.Tunnels {
		for k, addr := range t.Protocols {
			tunnelName := fmt.Sprintf("tunnel %s[%s]", name, k)
			if t.Protocols[k], err = normalizeAddress(addr, tunnelName); err != nil {
				return
			}

			if err = validateProtocol(k, tunnelName); err != nil {
				return
			}
		}
	}

	return
}

func defaultPath() string {
	user, err := user.Current()

	// user.Current() does not work on linux when cross compilling because
	// it requires CGO; use os.Getenv("HOME") hack until we compile natively
	homeDir := os.Getenv("HOME")
	if err != nil {
		log.Warn("Failed to get user's home directory: %s. Using $HOME: %s", err.Error(), homeDir)
	} else {
		homeDir = user.HomeDir
	}

	return path.Join(homeDir, ".ngrok")
}

func normalizeAddress(addr string, propName string) (string, error) {
	// normalize port to address
	if _, err := strconv.Atoi(addr); err == nil {
		addr = ":" + addr
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("Invalid %s '%s': %s", propName, addr, err.Error())
	}

	if tcpAddr.IP == nil {
		tcpAddr.IP = net.ParseIP("127.0.0.1")
	}

	return tcpAddr.String(), nil
}

func validateProtocol(proto, propName string) (err error) {
	switch proto {
	case "http", "https", "http+https", "tcp":
	default:
		err = fmt.Errorf("Invalid protocol for %s: %s", propName, proto)
	}

	return
}
