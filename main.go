package main

import (
	"context"
	"flag"
	"log"

	"github.com/bucksbunni/terraform-provider-openwrt/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version = "0.1.0"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable debug mode for the provider")
	flag.Parse()

	ctx := context.Background()
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/bucksbunni/terraform-provider-openwrt",
		Debug:   debug,
	}
	prv := provider.NewProvider(version)
	if err := providerserver.Serve(ctx, prv, opts); err != nil {
		log.Fatal(err)
	}
}
