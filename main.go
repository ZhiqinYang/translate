package main

import (
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	cli "github.com/urfave/cli"

	"google.golang.org/grpc"

	"translate/proto"

	log "github.com/Sirupsen/logrus"
)

func main() {
	go func() {
		log.Info(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	app := &cli.App{
		Name: "translate",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "listen",
				Value: ":10000",
				Usage: "listening address:port",
			},
			&cli.StringFlag{
				Name:  "path",
				Value: "/data/DIC.DAT",
				Usage: "dic snapshot file",
			},

			&cli.StringFlag{
				Name:  "apikey",
				Value: "AIzaSyBVh_ckjskfakldaskfkkdsakfkl",
				Usage: "translate services apikey",
			},
			&cli.DurationFlag{
				Name:  "persistence-interval",
				Value: 30 * time.Minute,
				Usage: "persistence interval",
			},
			&cli.StringFlag{
				Name:  "db",
				Value: "dic",
				Usage: "db name",
			},
			&cli.IntFlag{
				Name:  "limit",
				Value: 100000,
				Usage: "max translate cache data",
			},
		},
		Action: func(c *cli.Context) error {
			log.Println("listen:", c.String("listen"))
			log.Println("path:", c.String("path"))
			log.Println("apikey:", c.String("apikey"))
			log.Println("persistence-interval:", c.String("persistence-interval"))
			// 监听
			lis, err := net.Listen("tcp", c.String("listen"))
			if err != nil {
				log.Panic(err)
				os.Exit(-1)
			}
			log.Info("listening on ", lis.Addr())
			ins := new(server)
			ins.init(c)
			s := grpc.NewServer()
			// 注册服务
			proto.RegisterTranslateServiceServer(s, ins)
			//	初始化Serivces
			return s.Serve(lis)
		},
	}
	app.Run(os.Args)
}
