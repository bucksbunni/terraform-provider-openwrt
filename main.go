package main

import (
	"context"
	"flag"
	"log"

	"github.com/bucksbunni/terraform-provider-openwrt/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// "scaffold" default. Will be set by .gorelease configuration.
	version = "dev"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable debug mode for the provider")
	flag.Parse()

	ctx := context.Background()
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/bucksbunni/openwrt",
		Debug:   debug,
	}
	prv := provider.NewProvider(version)
	if err := providerserver.Serve(ctx, prv, opts); err != nil {
		log.Fatal(err)
	}
}
