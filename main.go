// Copyright 2015 mokey Authors. All rights reserved.
// Use of this source code is governed by a BSD style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/mokey/server"
	"github.com/ubccr/mokey/tools"
	"github.com/ubccr/mokey/util"
	"github.com/urfave/cli"
)

func init() {
	viper.SetConfigName("mokey")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/mokey/")
}

func main() {
	app := cli.NewApp()
	app.Name = "mokey"
	app.Authors = []cli.Author{cli.Author{Name: "Andrew E. Bruno", Email: "aebruno2@buffalo.edu"}}
	app.Usage = "mokey"
	app.Version = "0.5.6"
	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "conf,c", Usage: "Path to conf file"},
		&cli.BoolFlag{Name: "debug,d", Usage: "Print debug messages"},
	}
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.WarnLevel)
		}

		conf := c.GlobalString("conf")
		if len(conf) > 0 {
			viper.SetConfigFile(conf)
		}

		err := viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("Failed reading config file - %s", err)
		}

		if !viper.IsSet("enc_key") || !viper.IsSet("auth_key") {
			if !viper.IsSet("enc_key") {
				secret, err := util.GenerateSecret(16)
				if err != nil {
					return fmt.Errorf("Failed to generate enc_key - %s", err)
				}
				viper.Set("enc_key", secret)
			}
			if !viper.IsSet("auth_key") {
				secret, err := util.GenerateSecret(32)
				if err != nil {
					return fmt.Errorf("Failed to generate auth_key - %s", err)
				}
				viper.Set("auth_key", secret)
			}

			err = viper.WriteConfig()
			if err != nil {
				return fmt.Errorf("Failed to save secret keys to config file  - %s", err)
			}
		}

		return nil
	}
	app.Commands = []cli.Command{
		{
			Name:  "server",
			Usage: "Run http server",
			Action: func(c *cli.Context) error {
				err := server.Run()
				if err != nil {
					log.Fatal(err)
					return cli.NewExitError(err, 1)
				}

				return nil
			},
		},
		{
			Name:  "resetpw",
			Usage: "Send reset password email",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "uid, u", Usage: "User id"},
				&cli.StringFlag{Name: "email, e", Usage: "User email (if not provided, email from profile is used)", Value: ""},
			},
			Action: func(c *cli.Context) error {
				uid := c.String("uid")
				if len(uid) == 0 {
					return cli.NewExitError(errors.New("Please provide a uid"), 1)
				}
				email := c.String("email")

				err := tools.SendResetPasswordEmail(uid, email)
				if err != nil {
					return cli.NewExitError(err, 1)
				}

				return nil
			},
		},
		{
			Name:  "verify-email",
			Usage: "Re-send verify email",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "uid, u", Usage: "User id"},
			},
			Action: func(c *cli.Context) error {
				uid := c.String("uid")
				if len(uid) == 0 {
					return cli.NewExitError(errors.New("Please provide a uid"), 1)
				}

				err := tools.SendVerifyEmail(uid)
				if err != nil {
					return cli.NewExitError(err, 1)
				}

				return nil
			},
		},
		{
			Name:  "status",
			Usage: "Display token status for user",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "uid, u", Usage: "User id"},
			},
			Action: func(c *cli.Context) error {
				uid := c.String("uid")
				if len(uid) == 0 {
					return cli.NewExitError(errors.New("Please provide a uid"), 1)
				}

				err := tools.Status(uid)
				if err != nil {
					return cli.NewExitError(err, 1)
				}

				return nil
			},
		}}

	app.RunAndExitOnError()
}
