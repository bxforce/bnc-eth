/*
Copyright 2020 IRT SystemX

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	node "github.com/bxforce/bnc-eth/internal/app/node"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"time"
)

var app = cli.NewApp()

func main() {

	node, err := node.NewNode()

	app := &cli.App{
		Name:    "node",
		Usage:   "node for eth api",
		Version: "1.0.0",
		Commands: []*cli.Command{
			{
				Name:  "bootstrap",
				Usage: "bootstrap a network by enrolling validators",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "node", Value: false, Usage: "node", EnvVars: []string{"NODE"}},
					&cli.StringFlag{Name: "enroll", Value: "", Usage: "enroll", EnvVars: []string{"ENROLL"}},
					&cli.StringFlag{Name: "privateKey", Value: "", Usage: "privateKey", EnvVars: []string{"PRIVATE"}},
					&cli.StringFlag{Name: "config", Value: "", Usage: "config", EnvVars: []string{"CONFIG"}},
					&cli.IntFlag{Name: "validators", Value: 3, Usage: "min validators", EnvVars: []string{"VALIDATORS"}},
				},
				Action: func(ctx *cli.Context) error {
					node.SetPrivateKey(ctx.String("privateKey"))
					node.SetLimit(ctx.Int("validators"))
					node.SetConfig(ctx.String("config"))
					node.Initialize()
					enroll := ctx.String("enroll")
					if len(enroll) > 0 {
						go func() {
							time.Sleep(2 * time.Second)
							node.Enroll("localhost:"+node.GetServerPort(), enroll, ctx.Bool("node"))
							node.Run()
						}()
					}
					node.Serve()
					return nil
				},
			},
			{
				Name:  "join",
				Usage: "join a network and possibly enroll as validator",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "node", Value: false, Usage: "node", EnvVars: []string{"NODE"}},
					&cli.StringFlag{Name: "enroll", Value: "", Usage: "enroll", EnvVars: []string{"ENROLL"}},
					&cli.StringFlag{Name: "privateKey", Value: "", Usage: "privateKey", EnvVars: []string{"PRIVATE"}},
				},
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() == 0 {
						return errors.New("Error: bootstrap url is required")
					}
					bootstrap := ctx.Args().First()
					if len(ctx.String("enroll")) == 0 {
						return errors.New("Error: enroll name is required")
					}
					node.Initialize()
					node.Enroll(bootstrap, ctx.String("enroll"), ctx.Bool("node"))
					node.Run()
					return nil
				},
			},
			{
				Name:  "connect",
				Usage: "connect to a network through one node",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "privateKey", Value: "", Usage: "privateKey", EnvVars: []string{"PRIVATE"}},
				},
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() == 0 {
						return errors.New("Error: bootstrap url is required")
					}
					peer := ctx.Args().First()
					node.Initialize()
					node.Connect(peer)
					node.Run()
					return nil
				},
			},
			{
				Name:  "run",
				Usage: "run a simple node",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "privateKey", Value: "", Usage: "privateKey", EnvVars: []string{"PRIVATE"}},
					&cli.StringFlag{Name: "peers", Value: "", Usage: "peers", EnvVars: []string{"PEERS"}},
					&cli.BoolFlag{Name: "discover", Value: false, Usage: "discover"},
				},
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() == 0 {
						return errors.New("Error: genesis path is required")
					}
					genesis := ctx.Args().First()
					node.SetPrivateKey(ctx.String("privateKey"))
					node.Initialize()
					node.SetPeers(ctx.String("peers"))
					node.SetDiscover(ctx.Bool("discover"))
					node.Run(genesis)
					return nil
				},
			},
		},
		Action: func(ctx *cli.Context) error {
			cli.ShowAppHelp(ctx)
			return nil
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
