// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/ChainSafe/log15"
	"github.com/urfave/cli"
)

var app = cli.NewApp()

var (
	dumpConfigCommand = cli.Command{
		Action:      FixFlagOrder(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       AllFlags(),
		Category:    "CONFIGURATION DEBUGGING",
		Description: `The dumpconfig command shows configuration values.`,
	}
	initCommand = cli.Command{
		Action:      FixFlagOrder(initNode),
		Name:        "init",
		Usage:       "Initialize node genesis state",
		ArgsUsage:   "",
		Flags:       append(CLIFlags, NodeFlags...),
		Category:    "INITIALIZATION",
		Description: `The init command initializes the node with a genesis state. Usage: gossamer init --genesis genesis.json`,
	}
	accountCommand = cli.Command{
		Action:   FixFlagOrder(handleAccounts),
		Name:     "account",
		Usage:    "manage gossamer keystore",
		Flags:    append(CLIFlags, append(NodeFlags, AccountFlags...)...),
		Category: "KEYSTORE",
		Description: "The account command is used to manage the gossamer keystore.\n" +
			"\tTo generate a new sr25519 account: gossamer account --generate\n" +
			"\tTo generate a new ed25519 account: gossamer account --generate --ed25519\n" +
			"\tTo generate a new secp256k1 account: gossamer account --generate --secp256k1\n" +
			"\tTo import a keystore file: gossamer account --import=path/to/file\n" +
			"\tTo list keys: gossamer account --list",
	}
)

// init initializes gossamer
func init() {
	app.Action = gossamer
	app.Copyright = "Copyright 2019 ChainSafe Systems Authors"
	app.Name = "gossamer"
	app.Usage = "Official gossamer command-line interface"
	app.Author = "ChainSafe Systems 2019"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		dumpConfigCommand,
		initCommand,
		accountCommand,
	}
	app.Flags = append(app.Flags, CLIFlags...)
	app.Flags = append(app.Flags, NodeFlags...)
	app.Flags = append(app.Flags, AccountFlags...)
	app.Flags = append(app.Flags, NetworkFlags...)
	app.Flags = append(app.Flags, RPCFlags...)
}

func main() {
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startLogger(ctx *cli.Context) error {
	logger := log.Root()
	handler := logger.GetHandler()
	var lvl log.Lvl

	if lvlToInt, err := strconv.Atoi(ctx.String(VerbosityFlag.Name)); err == nil {
		lvl = log.Lvl(lvlToInt)
	} else if lvl, err = log.LvlFromString(ctx.String(VerbosityFlag.Name)); err != nil {
		return err
	}
	log.Root().SetHandler(log.LvlFilterHandler(lvl, handler))

	return nil
}

// initNode loads the genesis file and loads the initial state into the DB
func initNode(ctx *cli.Context) error {
	err := startLogger(ctx)
	if err != nil {
		log.Error("[gossamer] Failed to start logger", "error", err)
		return err
	}

	err = initializeNode(ctx)
	if err != nil {
		log.Error("[gossamer] Failed to initialize node", "error", err)
		return err
	}

	return nil
}

// gossamer is the main entrypoint into the gossamer system
func gossamer(ctx *cli.Context) error {
	err := startLogger(ctx)
	if err != nil {
		log.Error("[gossamer] Failed to start logger", "error", err)
		return err
	}

	if arguments := ctx.Args(); len(arguments) > 0 {
		return fmt.Errorf("failed to read command argument: %q", arguments[0])
	}

	node, _, err := makeNode(ctx)
	if err != nil {
		log.Error("[gossamer] Failed to initialize node", "error", err)
		return err
	}

	log.Info("[gossamer] Starting node services...", "name", node.Name)
	node.Start()

	return nil
}