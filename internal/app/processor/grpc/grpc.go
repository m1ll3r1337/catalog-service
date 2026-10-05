package pgrpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/m1ll3r1337/catalog-service/internal/app/config/section"
	"github.com/m1ll3r1337/catalog-service/internal/app/processor"
	catalogv1 "github.com/m1ll3r1337/catalog-service/internal/pkg/grpc/gen/catalog/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcProc struct {
	server *grpc.Server
	addr   string
}

func NewGRPC(
	catalogV1 catalogv1.CatalogServiceServer,
	cfg section.ProcessorGrpc,
) processor.Processor {
	srv := grpc.NewServer()
	catalogv1.RegisterCatalogServiceServer(srv, catalogV1)
	reflection.Register(srv)

	return &grpcProc{
		server: srv,
		addr:   fmt.Sprintf(":%d", cfg.ListenPort),
	}
}

func (p *grpcProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	lc := net.ListenConfig{}
	l, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start listening TCP addr")
		return
	}

	log.Info().Str("listen_addr", p.addr).Msg("Listening of TCP addr for gRPC server has been started")

	go func() {
		_ = p.server.Serve(l)
	}()

	processor.WatchForShutdown(ctx, wg, processor.CloserFunc(func() error {
		p.server.GracefulStop()
		return nil
	}))
}
