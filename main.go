package main

import (
	"os"

	"github.com/rancher/kube-api-auth/pkg"
	"github.com/rancher/kube-api-auth/pkg/service"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var VERSION = "v0.0.0-dev"

var appConfig = struct {
	Listen     string
	Namespace  string
	Kubeconfig string
}{}

func main() {
	textFormatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(textFormatter)

	app := cli.NewApp()
	app.Name = "kube-api-auth"
	app.Usage = "Rancher Kubernetes API auth service"
	app.Author = "Rancher Labs, Inc."
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Set true to get debug logs",
		},
		cli.StringFlag{
			Name:        "kubeconfig",
			EnvVar:      "KUBECONFIG",
			Usage:       "Kube config for accessing k8s cluster",
			Destination: &appConfig.Kubeconfig,
		},
		cli.StringFlag{
			Name:        "namespace",
			Value:       kubeapiauth.DefaultNamespace,
			Usage:       "Namespace of secrets",
			Destination: &appConfig.Namespace,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "create",
			Action: createToken,
		},
		{
			Name:   "serve",
			Action: startService,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "listen, l",
					Value:       kubeapiauth.DefaultListenHostPort,
					Usage:       "host:port to listen on (TCP)",
					Destination: &appConfig.Listen,
				},
			},
		},
	}
	app.Before = appBefore

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func appBefore(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
	log.Debug("Debug enabled!")
	return nil
}

func createToken(_ *cli.Context) error {
	log.Info("Not yet implemented")
	return nil
}

func startService(_ *cli.Context) error {
	return service.Serve(appConfig.Listen, appConfig.Namespace, appConfig.Kubeconfig)
}
