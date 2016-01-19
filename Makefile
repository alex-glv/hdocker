.PHONY: default build

GO=go

TOP_LEVEL="github.com/alex-glv/hdocker"

get:
	$(GO) get $(TOP_LEVEL)
build: 
	$(GO) build $(TOP_LEVEL)

install: build
	$(GO) install $(TOP_LEVEL)

clean:
	$(GO) clean $(TOP_LEVEL)


default: build
