package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kriive/hello-kube"
	"github.com/kriive/hello-kube/http"
)

const (
	DefaultAddr     = ":8080"
	EmptyConfigPath = ""
)

var (
	version = "N/A"
	commit  = "N/A"
)

type Config struct {
	HTTP struct {
		Addr   string `toml:"addr"`
		Domain string `toml:"domain"`
	} `toml:"http"`
}

func defaultConfig() Config {
	c := Config{}
	c.HTTP.Addr = DefaultAddr

	return c
}

type Main struct {
	HTTPServer *http.Server

	ConfigPath string
	Config     Config
}

func NewMain() *Main {
	return &Main{
		HTTPServer: http.NewServer(),
	}
}

func (m *Main) Run(ctx context.Context) error {
	m.HTTPServer.Addr = m.Config.HTTP.Addr
	m.HTTPServer.Domain = m.Config.HTTP.Domain

	if err := m.HTTPServer.Open(); err != nil {
		return err
	}

	// If TLS enabled, redirect non-TLS connections to TLS.
	if m.HTTPServer.UseTLS() {
		go func() {
			log.Fatal(http.ListenAndServeTLSRedirect(m.Config.HTTP.Domain))
		}()
	}

	log.Printf("running: url=%q", m.HTTPServer.URL())
	return nil
}

func (*Main) Close() error {
	return nil
}

func main() {
	hello.CommitHash = commit
	hello.Version = version

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	m := NewMain()

	if err := m.ParseConfig(ctx, os.Args[1:]); err == flag.ErrHelp {
		os.Exit(1)
	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "config not found: %s", m.ConfigPath)
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := m.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	<-ctx.Done()
	if err := m.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (m *Main) ParseConfig(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("hellod", flag.ContinueOnError)

	fs.StringVar(&m.ConfigPath, "config", EmptyConfigPath, "config path, leave unset for the default")
	if err := fs.Parse(args); err != nil {
		return err
	}

	m.Config = defaultConfig()

	// If the config path is empty, just assume we want the
	// default configuration.
	if m.ConfigPath == "" {
		return nil
	}

	configPath, err := expand(m.ConfigPath)
	if err != nil {
		return err
	}

	return parseConfigFile(configPath, &m.Config)
}

func parseConfigFile(path string, config *Config) error {
	c, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return toml.Unmarshal(c, config)
}

func expand(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		return path, nil
	}

	u, err := os.UserHomeDir()
	if err != nil {
		return path, err
	} else if u == "" {
		return path, fmt.Errorf("home directory unset")
	}

	if path == "~" {
		return u, nil
	}

	// Clean does not provide any security mechanism!
	return filepath.Clean(
		filepath.Join(u, strings.TrimPrefix(path, "~"+string(os.PathSeparator))),
	), nil
}
