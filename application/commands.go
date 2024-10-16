package application

import (
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/jedib0t/go-pretty/table"
	"github.com/namsral/flag"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
)

type appCommand struct {
	app contracts.Application
	cfg contracts.Config
	log logger.Logger
}

func NewAppCommands(app contracts.Application, cfg contracts.Config, log logger.Logger) []*cli.Command {
	cmd := &appCommand{
		app: app,
		cfg: cfg,
		log: log.With(logger.WithComponent, "application"),
	}
	return []*cli.Command{
		{
			Category: "app",
			Name:     "app:env",
			Usage:    "Print all configurable envs",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "out",
					Usage: "Out variant (compose, helm)",
				},
			},
			Action: cmd.appEnv,
		},
		{
			Category: "app",
			Name:     "app:start",
			Usage:    "Start application and all services",
			Action:   cmd.appStart,
		},
	}
}

func (c *appCommand) appEnv(ctx *cli.Context) error {
	type configFlag struct {
		Value    string
		Default  string
		Usage    string
		Provider []string
	}
	configs := make(map[string]configFlag)
	c.cfg.Visit(
		func(f contracts.ConfigFlag) {
			if f.Name() != flag.DefaultConfigFlagname {
				if _, ok := configs[f.Name()]; !ok {
					configs[f.Name()] = configFlag{
						Value:   f.Value(),
						Default: f.Default(),
						Usage:   f.Usage(),
					}
				}
				cfg := configs[f.Name()]
				cfg.Provider = append(cfg.Provider, f.Provider())
				configs[f.Name()] = cfg
			}
		},
	)
	keys := make([]string, 0, len(configs))
	for name := range configs {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	switch ctx.String("out") {
	case "configmap":
		fmt.Println("--------------------------------")
		fmt.Println("data:")
		for _, key := range keys {
			cfg := configs[key]
			fmt.Printf("  %s: '%s'\n", key, cfg.Value)
		}
		fmt.Println("--------------------------------")
	case "compose":
		fmt.Println("--------------------------------")
		fmt.Println("environment:")
		for _, key := range keys {
			cfg := configs[key]
			fmt.Printf("  - %s=%s\n", key, cfg.Value)
		}
		fmt.Println("--------------------------------")
	case "helm":
		fmt.Println("--------------------------------")
		fmt.Println("env:")
		for _, key := range keys {
			cfg := configs[key]
			fmt.Printf("  - name: %s\n", key)
			fmt.Printf("    value: \"%s\"\n", cfg.Value)
		}
		fmt.Println("--------------------------------")
	default:
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Name", "Default", "Current", "Usage", "Provider"})
		for _, key := range keys {
			cfg := configs[key]
			t.AppendRow(table.Row{key, cfg.Default, cfg.Value, cfg.Usage, strings.Join(cfg.Provider, ", ")})
		}
		t.Render()
	}
	return nil
}

//goland:noinspection SpellCheckingInspection
func (c *appCommand) appStart(*cli.Context) error {
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT) //nolint:govet,staticcheck

	c.app.InitService()
	c.app.StartService()

	c.log.Infow("application is finished", "signal", (<-done).String())
	return nil
}
