package transport

import (
	"fmt"
	"github.com/jedib0t/go-pretty/table"
	"github.com/urfave/cli/v2"
	"os"
)

type grpcListCommand struct {
	server GrpcServer
}

func NewRpcListCommand(server GrpcServer) *cli.Command {
	cmd := &grpcListCommand{server}
	return &cli.Command{
		Category: "grpc",
		Name:     "grpc:list",
		Usage:    "List all rpc methods",
		Action:   cmd.handle,
	}
}

func (c *grpcListCommand) handle(*cli.Context) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Full name", "Client stream", "Server stream", "Metadata"})

	for service, info := range c.server.GetServiceInfo() {
		for _, method := range info.Methods {
			name := fmt.Sprintf("/%s/%s", service, method.Name)
			t.AppendRow(table.Row{name, method.IsClientStream, method.IsServerStream, info.Metadata})
		}
	}

	t.Render()
	return nil
}
