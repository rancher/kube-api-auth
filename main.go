package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	kubeapiauth "github.com/rancher/kube-api-auth/pkg"
	"github.com/rancher/kube-api-auth/pkg/service"
	log "github.com/sirupsen/logrus"
)

var VERSION = "v0.0.0-dev"

// legacyCommand is the subcommand earlier versions required. Rancher's
// DaemonSet passes no arguments, so only the image and hand-written
// invocations ever supplied it. Accepted with a warning so an image or runbook
// still carrying it keeps working.
const legacyCommand = "serve"

type config struct {
	Listen     string
	Namespace  string
	Kubeconfig string
	Debug      bool
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if err := checkArgs(args); err != nil {
		return err
	}

	cfg := configFromEnv(os.Getenv)
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug enabled!")
	}

	log.Infof("kube-api-auth version %s is starting", VERSION)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return service.Serve(ctx, cfg.Listen, cfg.Namespace, cfg.Kubeconfig)
}

// checkArgs rejects anything the binary no longer understands, so a stale flag
// fails at startup instead of being silently replaced by a default.
func checkArgs(args []string) error {
	if len(args) > 0 && args[0] == legacyCommand {
		log.Warnf("The %q command is deprecated and does nothing, drop it from the invocation", legacyCommand)
		args = args[1:]
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized arguments %v, kube-api-auth is configured through CATTLE_DEBUG (or RANCHER_DEBUG), KUBECONFIG, CATTLE_NAMESPACE and CATTLE_LISTEN", args)
	}

	return nil
}

// configFromEnv reads the configuration through the given lookup, which is
// os.Getenv outside of tests.
func configFromEnv(getenv func(string) string) config {
	cfg := config{
		Listen:     kubeapiauth.DefaultListenHostPort,
		Namespace:  kubeapiauth.DefaultNamespace,
		Kubeconfig: getenv("KUBECONFIG"),
		Debug:      isTrue(getenv("CATTLE_DEBUG")) || isTrue(getenv("RANCHER_DEBUG")),
	}

	if listen := getenv("CATTLE_LISTEN"); listen != "" {
		cfg.Listen = listen
	}
	if namespace := getenv("CATTLE_NAMESPACE"); namespace != "" {
		cfg.Namespace = namespace
	}

	return cfg
}

func isTrue(value string) bool {
	enabled, err := strconv.ParseBool(value)
	return err == nil && enabled
}
