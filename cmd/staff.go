package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tespkg/bytes-be/config"
	"github.com/tespkg/bytes-be/svc"
	"github.com/tespkg/bytes-be/svc/staff/rest"
	"os"
	"os/signal"
	"syscall"
	"tespkg.in/kit/log"
)

var configPath string
var signals chan os.Signal

var staffCmd = &cobra.Command{
	Use:   "staff",
	Short: "start a staff server for bytes backend",
	Run: func(cmd *cobra.Command, args []string) {
		servers := []svc.Service{
			&rest.Server{},
		}

		runServices(servers)
	},
}

func runServices(servers []svc.Service) {
	//load default config
	cfg, err := config.LoadWithDefault(configPath)
	if err != nil {
		log.Fatalf("load config, path[%s], err: %v", configPath, err)
	}

	log.Infof("load config OK")

	//run all servers
	for _, server := range servers {
		runOneService(server, cfg)
	}

	//register stop sign
	var stopFns []func(signal os.Signal)
	for _, server := range servers {
		stopFns = append(stopFns, server.Stop)
	}

	var stop = func(s os.Signal) {
		for idx := len(stopFns) - 1; idx >= 0; idx-- {
			fn := stopFns[idx]
			fn(s)
		}
	}

	embedNotifySignal(stop)
}

func runOneService(server svc.Service, cfg config.Config) {
	readyChan := make(chan bool, 1)

	//load server
	if err := server.Load(cfg); err != nil {
		log.Fatalf("load %s server, err: %v", server.Name(), err)
	}

	log.Infof("load %s server OK", server.Name())

	go server.Run(readyChan)
	<-readyChan

	log.Infof("%s server started", server.Name())
}

func embedNotifySignal(fn func(os.Signal)) {
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sign := <-signals
	fn(sign)
}

func init() {
	staffCmd.PersistentFlags().StringVarP(
		&configPath,
		"config",
		"c",
		"server.yaml",
		"the path of yaml file, the config are loaded by environment variable.",
	)

	signals = make(chan os.Signal)

	rootCmd.AddCommand(staffCmd)
}
