package main

import (
	"context"
	"os"
	"time"

	taskExample "worker-ex/tasks"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/log"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	app *cli.App
	// taskServer *machinery.Server
)

func init() {
	app = cli.NewApp()
	app.Name = "worker-example"
	app.Usage = "enqueue task pattern"

	// config := config.Config{
	// 	Broker: "localhost:6432",
	// 	DefaultQueue: "localhost:6433",
	// }
	// taskServer = machinery.NewServer()

}

func main() {
	app.Commands = []cli.Command{
		{
			Name:  "worker",
			Usage: "launch machinery worker",
			Action: func(c *cli.Context) error {
				if err := worker(); err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
		{
			Name:  "server",
			Usage: "send example task",
			Action: func(c *cli.Context) error {
				if err := send(); err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}

func startServer() (*machinery.Server, error) {
	cnf := &config.Config{
		DefaultQueue:    "machinery_tasks",
		ResultsExpireIn: 3600,
		Broker:          "redis://localhost:6379/0",
		ResultBackend:   "redis://localhost:6379/0",
		Redis: &config.RedisConfig{
			MaxIdle:                3,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
		},
	}

	server, err := machinery.NewServer(cnf)
	if err != nil {
		logrus.Fatal(err)
		return nil, err
	}

	tasks := map[string]interface{}{
		"add":       taskExample.Add,
		"long_task": taskExample.LongRunningTask,
		"panic":     taskExample.PanicTask,
	}

	return server, server.RegisterTasks(tasks)
}

func worker() error {
	consumerTag := "machinery_worker"

	server, err := startServer()
	if err != nil {
		logrus.Fatal(err)
		return err
	}

	worker := server.NewWorker(consumerTag, 2)

	errorhandler := func(err error) {
		log.ERROR.Println("I am an error handler:", err)
	}

	pretaskhandler := func(signature *tasks.Signature) {
		log.INFO.Println("I am a start of task handler for:", signature.Name)
	}

	posttaskhandler := func(signature *tasks.Signature) {
		log.INFO.Println("I am an end of task handler for:", signature.Name)
	}

	worker.SetErrorHandler(errorhandler)
	worker.SetPreTaskHandler(pretaskhandler)
	worker.SetPostTaskHandler(posttaskhandler)

	return worker.Launch()
}

func send() error {
	server, err := startServer()
	if err != nil {
		logrus.Fatal(err)
	}

	asyncRes, err := server.SendTaskWithContext(context.Background(), &tasks.Signature{
		Name: "add",
		Args: []tasks.Arg{
			{
				Type:  "int64",
				Value: 1,
			},
			{
				Type:  "int64",
				Value: 2,
			},
			{
				Type:  "int64",
				Value: 3,
			},
		},
	})
	if err != nil {
		logrus.Errorf("could not send task : %v", err.Error())
		return err
	}

	res, err := asyncRes.Get(time.Duration(time.Millisecond * 5))
	if err != nil {
		logrus.Errorf("could not get result : %s", err.Error())
		return err
	}
	log.INFO.Printf("result of sum tasks : %v", tasks.HumanReadableResults(res))

	return nil
}
