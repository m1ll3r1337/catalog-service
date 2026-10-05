package pgateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/m1ll3r1337/catalog-service/internal/app/config/section"
	"github.com/m1ll3r1337/catalog-service/internal/app/processor"
	catalogv1 "github.com/m1ll3r1337/catalog-service/internal/pkg/grpc/gen/catalog/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	gatewayReadHeaderTimeout = 5 * time.Second
	gatewayReadTimeout       = 30 * time.Second
	gatewayWriteTimeout      = 30 * time.Second
	gatewayIdleTimeout       = 120 * time.Second
	gatewayShutdownTimeout   = 5 * time.Second
)

type gatewayProc struct {
	server   http.Server
	addr     string
	grpcAddr string
}

func NewGateway(
	cfgGateway section.ProcessorGateway,
	cfgGrpc section.ProcessorGrpc,
) processor.Processor {
	httpAddr := fmt.Sprintf(":%d", cfgGateway.ListenPort)
	grpcAddr := net.JoinHostPort("localhost", strconv.Itoa(int(cfgGrpc.ListenPort)))

	return &gatewayProc{
		server: http.Server{
			ReadHeaderTimeout: gatewayReadHeaderTimeout,
			ReadTimeout:       gatewayReadTimeout,
			WriteTimeout:      gatewayWriteTimeout,
			IdleTimeout:       gatewayIdleTimeout,
			Addr:              httpAddr,
		},
		addr:     httpAddr,
		grpcAddr: grpcAddr,
	}
}

func (p *gatewayProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	mux := runtime.NewServeMux()

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())

	err := catalogv1.RegisterCatalogServiceHandlerFromEndpoint(ctx, mux, p.grpcAddr, []grpc.DialOption{opts})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register catalog handler from endpoint")
		return
	}

	log.Info().Str("listen_addr", p.addr).Msg("Listening of TCP addr for gRPC-Gateway server has been started")

	p.server.Handler = mux

	lc := net.ListenConfig{}

	l, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start listening TCP addr")
		return
	}

	go func() { _ = p.server.Serve(l) }()

	processor.WatchForShutdown(ctx, wg, processor.CloserFunc(l.Close))
	processor.WatchForShutdown(ctx, wg, processor.NewCloserContextFunc(p.server.Shutdown, context.Background(), gatewayShutdownTimeout))
}
