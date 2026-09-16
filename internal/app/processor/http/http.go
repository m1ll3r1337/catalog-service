package rprocessor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/m1ll3r1337/catalog-service/internal/app/config/section"
	rhandler "github.com/m1ll3r1337/catalog-service/internal/app/handler/http"
	"github.com/m1ll3r1337/catalog-service/internal/app/processor"
	"github.com/m1ll3r1337/catalog-service/internal/app/util"
	"github.com/m1ll3r1337/catalog-service/internal/pkg/http/httph"
	"github.com/m1ll3r1337/catalog-service/internal/pkg/http/mzerolog"
	"github.com/rs/zerolog/log"
)

type httpProc struct {
	server http.Server
	addr   string
}

func NewHTTP(hHealth rhandler.Health, hCategory rhandler.Category, hProduct rhandler.Product, cfg section.ProcessorWebServer) processor.Processor {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(handlerNotFound)

	r.Use(
		httph.NewErrorMiddleware(),
		mzerolog.NewMiddleware(
			mzerolog.WithSkipper(util.IsFilteredHttpRoute),
		),
	)

	vGenericRegHealthCheck(r, hHealth)

	rV1 := r.PathPrefix("/v1").Subrouter()

	v1RegCategoryHandler(rV1, hCategory)
	v1RegProductHandler(rV1, hProduct)

	_ = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}

		if pathTemplate == "" {
			return nil
		}

		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}

		if len(methods) == 0 {
			return nil
		}

		log.Info().Strs("methods", methods).Str("path", pathTemplate).Msg("Route registered")

		return nil
	})

	p := httpProc{addr: fmt.Sprintf(":%d", cfg.ListenPort)}
	p.server.Handler = r

	return &p
}

func (p *httpProc) serve(l net.Listener) {
	_ = p.server.Serve(l)
}

func (p *httpProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	lc := net.ListenConfig{}
	l, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start listening TCP addr")
	}

	log.Info().Str("listen_addr", p.addr).Msg("Listening of TCP addr for HTTP server has been started")

	go p.serve(l)

	processor.WatchForShutdown(ctx, wg, processor.CloserFunc(l.Close))
	processor.WatchForShutdown(ctx, wg, processor.NewCloserContextFunc(p.server.Shutdown, context.Background(), 5*time.Second))
}
