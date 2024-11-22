package cron

import (
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/jedib0t/go-pretty/table"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type cronCommand struct {
	app  contracts.Application
	cron Service
	log  logger.Logger
}

func NewCronCommands(app contracts.Application, cron Service, log logger.Logger) []*cli.Command {
	cmd := &cronCommand{
		app:  app,
		cron: cron,
		log:  log,
	}
	return []*cli.Command{
		{
			Category: "cron",
			Name:     "cron:list",
			Usage:    "List all cron tasks",
			Action:   cmd.cronList,
		},
		{
			Category: "cron",
			Name:     "cron:call",
			Usage:    "Call cron task by name",
			Action:   cmd.cronRun,
		},
	}
}

func (c *cronCommand) cronList(*cli.Context) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Next run"})

	for _, task := range c.cron.Tasks() {
		next := task.Schedule().Next(time.Now())
		t.AppendRow(table.Row{task.Name(), next.Format("2006-01-02 15:04:05")})
	}

	t.Render()
	return nil
}

func (c *cronCommand) cronRun(ctx *cli.Context) error {
	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM) //nolint:govet,staticcheck

	name := ctx.Args().First()
	c.app.InitService()

	done := make(chan bool, 1)
	go func() {
		err := c.cron.Call(name)
		if err != nil {
			c.log.Errorw("cron command failed", zap.Error(err))
		}
		done <- true
	}()

	select {
	case s := <-exit:
		c.log.Infow(fmt.Sprintf("cron:call %s is canceled", name), "signal", s.String())
	case <-done:
		c.log.Infof("cron:call %s is finished", name)
	}

	return nil
}
