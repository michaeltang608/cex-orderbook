package main

import (
	"net"

	"google.golang.org/grpc"
	pb "lightning-engine/api/match/v1"
	"lightning-engine/conf"
	"lightning-engine/internal/server"
	"lightning-engine/internal/status"
	mainlog "lightning-engine/log"
)

func init() {
	mainlog.InitLog()
}

func main() {
	symbol := conf.Gconfig.GetString("pair.symbol")
	pairs := []string{symbol}
	app, cleanup, err := wireApp(pairs)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	go app.sysSignalHandle.Begin()

	port := conf.Gconfig.GetString("server.port")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		mainlog.Info("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMatchServiceServer(grpcServer, app.server)
	mainlog.Info("[RPC] %s\n", port)
	grpcServer.Serve(lis)
	select {}
}

type app struct {
	status          *status.Status
	server          *server.Server
	sysSignalHandle *status.SysSignalHandle
}

func newApp(st *status.Status, se *server.Server, ss *status.SysSignalHandle) *app {
	return &app{
		status:          st,
		server:          se,
		sysSignalHandle: ss,
	}
}
