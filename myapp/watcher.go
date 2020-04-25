package myapp

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"search-and-sort-movies/myapp/constants"
	"time"
)

var watch *fsnotify.Watcher
var err error

func MyWaTcher(location string) {
	watch, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watch.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watch.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					re := regexp.MustCompile(constants.RegexFile)
					if !_checkIfDir(event) {
						if re.MatchString(filepath.Ext(event.Name)) {
							go _fsNotifyCreateFile(event, re)
						}
					}
				}
			case err := <-watch.Errors:
				log.Println("error:", err)
				//close(done)
			}
		}
	}()

	log.Printf("add watcher %s\n", location)
	if err := watch.Add(location); err != nil {
		log.Fatal(err)
	}

	<-done
}

func _ticker(event fsnotify.Event, c *chan bool) {
	ticker := time.NewTicker(1 * time.Second)
	//ticker := time.NewTicker(5 * time.Second)
	var size int64 = -1
	go func() {
		for range ticker.C {
			f, err := os.Stat(event.Name)
			if err != nil {
				log.Println(err)
			}
			//log.Printf("Name: %s\n\tInfo size: %d - Size: %d\n\n", event.Name, f.Size(), size)
			if f.Size() != size {
				size = f.Size()
				continue
			}
			ticker.Stop()
			*c <- true
		}
	}()
}

func _stat(event fsnotify.Event) (os.FileInfo, fsnotify.Event) {
	f, err := os.Stat(event.Name)
	if err != nil {
		log.Println(err)
	}
	return f, event
}

func _checkIfDir(event fsnotify.Event) bool {
	f, e := _stat(event)
	log.Printf("f: %v, e: %v", f, e)
	if f.IsDir() && filepath.Dir(f.Name()) != GetEnv("dlna") {
		err := watch.Add(e.Name)
		log.Printf("add watcher %s\n", e.Name)
		if err != nil {
			log.Println(err)
		}
		return true
	}
	return false
}

func _fsNotifyCreateFile(event fsnotify.Event, re *regexp.Regexp) {
	_, e := _stat(event)

	finish := make(chan bool)
	go _ticker(event, &finish)
	<-finish

	if re.MatchString(filepath.Ext(e.Name)) {
		log.Println("Détection de :", filepath.Base(e.Name))
		//folder := filepath.Dir(e.Name)
		Process(event.Name)
		//if folder != GetEnv("dlna") {
		//	files, _ := ioutil.ReadDir(folder)
		//	var countFile = 0
		//	for _, file := range files {
		//		if file.Name()[0] != '.'{
		//			countFile++
		//		}
		//	}
		//	fmt.Println(countFile)
		//	if countFile == 0 {
		//		log.Printf("remove watcher %s\n", e.Name)
		//		err := watch.Remove(folder)
		//		if err != nil {
		//			log.Println(err)
		//		}
		//	}
		//}
	}
}
