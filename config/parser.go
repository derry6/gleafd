package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/derry6/gleafd/version"
	yaml "gopkg.in/yaml.v2"
)

type parser struct {
	Cfg         *Config       `yaml:"gleafd"`
	showVersion bool          `yaml:"-"`
	flagSet     *flag.FlagSet `yaml:"-"`
	fileName    string        `yaml:"-"`
}

func (p *parser) parseFromEnv() (err error) {
	setted := make(map[string]bool)
	p.flagSet.Visit(func(f *flag.Flag) {
		setted[f.Name] = true
	})
	p.flagSet.VisitAll(func(f *flag.Flag) {
		if _, ok := setted[f.Name]; ok {
			return
		}
		name := envPrefix + strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
		if val, found := os.LookupEnv(name); found {
			if ferr := f.Value.Set(val); ferr != nil {
				err = ferr
			}
		}
	})
	return err
}

func (p *parser) parseFromFile(fileName string) (err error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(data, p); err != nil {
		return err
	}
	return p.validate()
}

func (p *parser) validate() error {
	return nil
}

func (p *parser) parse(args []string) error {
	err := p.flagSet.Parse(args)
	if err != nil {
		return err
	}
	if len(p.flagSet.Args()) != 0 {
		return fmt.Errorf("'%s' is not a valid flag", p.flagSet.Arg(0))
	}
	// Show versions
	if p.showVersion {
		fmt.Printf("Version: %s\n", version.Version)
		fmt.Printf("Git SHA: %s\n", version.GitSHA)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
	if p.fileName != "" {
		return p.parseFromFile(p.fileName)
	}
	return p.parseFromEnv()
}

func Load(args []string) (*Config, error) {
	p := &parser{
		Cfg:     newConfig(),
		flagSet: flag.NewFlagSet("gleafd", flag.ExitOnError),
	}
	flagSet := p.flagSet

	flagSet.StringVar(&p.fileName, "config", "", "Location of server config file")
	flagSet.BoolVar(&p.showVersion, "version", false, "show version")

	// Server
	flagSet.StringVar(&p.Cfg.Name, "name", p.Cfg.Name, "Assign a name to the server")
	flagSet.StringVar(&p.Cfg.Addr, "addr", p.Cfg.Addr, "Listen address")
	flagSet.StringVar(&p.Cfg.Log, "log", p.Cfg.Log, "Log level [debug|info|warn|error|fatal]")

	// Segment
	seg := &p.Cfg.Segment
	flagSet.BoolVar(&seg.Enable, "segment-enable", seg.Enable, "Enable segment")
	flagSet.StringVar(&seg.DBHost, "segment-db-host", seg.DBHost, "")
	flagSet.StringVar(&seg.DBName, "segment-db-name", seg.DBName, "")
	flagSet.StringVar(&seg.DBUser, "segment-db-user", seg.DBUser, "")
	flagSet.StringVar(&seg.DBPass, "segment-db-pass", seg.DBPass, "")

	// Snowflake
	sf := &p.Cfg.Snowflake
	flagSet.BoolVar(&sf.Enable, "snowflake-enable", sf.Enable, "")
	flagSet.StringVar(&sf.RedisAddresss, "snowflake-redis-addr", sf.RedisAddresss, "")

	if err := p.parse(args); err != nil {
		return nil, err
	}
	return p.Cfg, nil
}
