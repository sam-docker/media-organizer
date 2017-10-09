package controllers

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func Watcher(location string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("Détection du fichier : ", event.Name)
					Process(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(location)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
