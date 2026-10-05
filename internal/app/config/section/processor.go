package section

type (
	Processor struct {
		WebServer ProcessorWebServer `split_words:"true"`
		Grpc      ProcessorGrpc      `split_words:"true"`
		Gateway   ProcessorGateway   `split_words:"true"`
	}

	ProcessorGateway struct {
		ListenPort uint32 `split_words:"true" default:"8081"`
	}

	ProcessorWebServer struct {
		ListenPort uint32 `split_words:"true" default:"8080"`
	}

	ProcessorGrpc struct {
		ListenPort uint32 `split_words:"true" default:"50051"`
	}
)
