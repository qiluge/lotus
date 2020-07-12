package main

import (
	"database/sql"
	"github.com/filecoin-project/lotus/cmd/lotus-chainwatch/processor"
	"github.com/filecoin-project/lotus/cmd/lotus-chainwatch/syncer"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/filecoin-project/lotus/build"
	lcli "github.com/filecoin-project/lotus/cli"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
)

var log = logging.Logger("chainwatch")

func main() {
	_ = logging.SetLogLevel("*", "INFO")
	if err := logging.SetLogLevel("rpc", "error"); err != nil {
		panic(err)
	}

	log.Info("Starting chainwatch")

	local := []*cli.Command{
		runCmd,
		//dotCmd,
	}

	app := &cli.App{
		Name:    "lotus-chainwatch",
		Usage:   "Devnet token distribution utility",
		Version: build.UserVersion(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				EnvVars: []string{"LOTUS_PATH"},
				Value:   "~/.lotus", // TODO: Consider XDG_DATA_HOME
			},
			&cli.StringFlag{
				Name:    "db",
				EnvVars: []string{"LOTUS_DB"},
				Value:   "postgres://postgres:password@192.168.168.10:5432/postgres?sslmode=disable",
			},
		},

		Commands: local,
	}

	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		os.Exit(1)
	}
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "Start lotus chainwatch",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "front",
			Value: "127.0.0.1:8418",
		},
		&cli.IntFlag{
			Name:  "max-batch",
			Value: 1000,
		},
	},
	Action: func(cctx *cli.Context) error {
		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		v, err := api.Version(ctx)
		if err != nil {
			return err
		}

		log.Infof("Remote version: %s", v.Version)

		maxBatch := cctx.Int("max-batch")

		db, err := sql.Open("postgres", cctx.String("db"))
		defer db.Close()
		/*
			st, err := openStorage(cctx.String("db"))
			if err != nil {
				return err
			}
			defer st.close() //nolint:errcheck

				runSyncer(ctx, api, st, maxBatch)

				h, err := newHandler(api, st)
				if err != nil {
					return xerrors.Errorf("handler setup: %w", err)
				}

				http.Handle("/", h)

				fmt.Printf("Open http://%s\n", cctx.String("front"))
		*/

		sync := syncer.NewSyncer(db, api)
		sync.Start(ctx)

		proc := processor.NewProcessor(db, api, maxBatch)
		proc.Start(ctx)

		go func() {
			<-ctx.Done()
			os.Exit(0)
		}()

		return http.ListenAndServe(cctx.String("front"), nil)
	},
}
