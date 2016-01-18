package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"gopkg.in/fsnotify.v1"
	"os"
	"path/filepath"
)

func main() {
	app := cli.NewApp()
	app.Name = "watch"
	app.Usage = "a glorified file watcher"
	app.Action = func(c *cli.Context) {
		run(c.Args().First())
	}
	app.Before = func(c *cli.Context) error {
		if c.Args().First() == "" {
			return errors.New("must pass in one directory to watch argument")
		}
		return nil
	}

	app.RunAndExitOnError()
}

func run(dir string) {
	log.WithField("path", dir).Info("Scanning path")

	// create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// set up 'watching'
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.WithFields(log.Fields{
					"path":      event.Name,
					"operation": event.Op,
				}).Info("change detected")
			case err := <-watcher.Errors:
				log.WithField("error", err).Error("error on watch")
			}
		}
	}()

	// recursively add watcher to everything
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.WithFields(log.Fields{
				"path":  path,
				"error": err,
			}).Warn("unable to watch path")
			return nil
		}
		if err = watcher.Add(path); err != nil {
			log.WithFields(log.Fields{
				"path":  path,
				"error": err,
			}).Warn("unable to watch path")
		}
		return nil
	})
	if err != nil {
		log.WithFields(log.Fields{
			"path":  dir,
			"error": err,
		}).Fatal("unable to fully walk path, exiting")
	}
	done := make(chan struct{})
	<-done
}
