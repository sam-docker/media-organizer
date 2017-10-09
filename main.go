package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"search-and-sort-movies/controllers"
	"strings"

	"github.com/gonutz/w32"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.LstdFlags | log.Lshortfile)

}

var (
	BuildVersion string
	BuildHash    string
	BuildDate    string
	BuildClean   string
	Name         = "search-and-sort-movies"
)

func main() {

	vers := flag.Bool("v", false, "Indique la version de l'application")
	scan := flag.Bool("scan", false, "Lancer le scan au démarrage de l'application")
	windows := flag.Bool("windows", false, "Lancer l'application sans l'invite de commandes")
	flag.Parse()

	if *vers {
		// flag.PrintDefaults()
		fmt.Printf("Name: %s\n", Name)
		fmt.Printf("Version: %s\n", BuildVersion)
		fmt.Printf("Git Commit Hash: %s\n", BuildHash)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Built from clean source tree: %s\n", BuildClean)
		fmt.Printf("OS: %s\n", runtime.GOOS)
		fmt.Printf("Architecture: %s\n", runtime.GOARCH)
		os.Exit(1)
	}

	if *scan {
		startScan(true)
	}

	if *windows {
		console := w32.GetConsoleWindow()
		if console != 0 {
			_, consoleProcID := w32.GetWindowThreadProcessId(console)
			if w32.GetCurrentProcessId() == consoleProcID {
				w32.ShowWindowAsync(console, w32.SW_HIDE)
			}
		}
	}

	// Write log to file : log_SearchAndSort
	f, err := os.OpenFile("log_SearchAndSort", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Check if it's the first connection
	if firstConnect() {
		firstConfig()
	} else {
		for {
			if controllers.GetEnv("dlna") == "" || controllers.GetEnv("movies") == "" || controllers.GetEnv("series") == "" {
				firstConfig()
			} else {
				break
			}
		}
	}

	checkFolderExists(controllers.GetEnv("dlna"))
	checkFolderExists(controllers.GetEnv("movies"))
	checkFolderExists(controllers.GetEnv("series"))

	fmt.Println("Start :-D")

	// startScan(false)

	fmt.Println("Ecoute sur le dossier : " + controllers.GetEnv("dlna"))
	controllers.Watcher(controllers.GetEnv("dlna"))

}

func startScan(auto bool) {
	if count, file := fileInFolder(); count > 0 {
		if auto {
			fmt.Println("Scan automatique")
			go boucleFiles(file)
		} else {
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("Je vois qu'il y a des fichiers vidéos actuellement dans ton dossier source.")
			fmt.Println("Veux tu faire le tri? (O/n)")
			text, _ := reader.ReadString('\n')
			fmt.Println(text)
			if strings.TrimSpace(text) == "n" || strings.TrimSpace(text) == "N" {
				return
			}

			go boucleFiles(file)
		}
	}
}

func firstConnect() bool {
	_, err := os.Stat(".config.json")

	if os.IsNotExist(err) {
		log.Println(err)
		return true
	}
	return false
}

func readJSONFile() {
	f, err := ioutil.ReadFile(".config.json")

	if err != nil {
		log.Println(err)
	}

	fmt.Printf("%s\n", string(f))
}

func firstConfig() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Hello, bienvenue sur l'application de tri des vidéos.")
		fmt.Println("Ceci est ta première connexion donc il faut configurer des petites choses.")
		fmt.Println("Commençons par l'emplacement où les fichiers sont téléchargés : ")
		pwd, _ := os.Getwd()
		fmt.Println("A savoir, que tu te trouves dans le répertoire : " + pwd)
		text, _ := reader.ReadString('\n')
		fmt.Println(text)
		controllers.SetEnv("dlna", path.Clean(strings.TrimSpace(text)))
		fmt.Println("Ensuite, il faut renter le dossier des films : ")
		text, _ = reader.ReadString('\n')
		fmt.Println(text)
		controllers.SetEnv("movies", path.Clean(strings.TrimSpace(text)))
		fmt.Println("Pour finir, il faut rentrer le dossier des séries : ")
		text, _ = reader.ReadString('\n')
		fmt.Println(text)
		controllers.SetEnv("series", path.Clean(strings.TrimSpace(text)))
		fmt.Println("Pour la musique, il faut attendre les prochaines versions. :-(  ")

		fmt.Println("Super. Voilà tout est configuré. On va vérifier le fichier : ")
		fmt.Println('\n')
		readJSONFile()
		fmt.Println("Est-ce que cela est correct? (O/n)")
		text, _ = reader.ReadString('\n')
		if strings.TrimSpace(text) == "n" || strings.TrimSpace(text) == "N" {
			continue
		} else {
			break
		}

	}

	fmt.Println("Cool!!! C'est parti. Enjoy")
}

func checkFolderExists(folder string) {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, os.ModePerm)
	}
}

func fileInFolder() (int, []os.FileInfo) {
	files, err := ioutil.ReadDir(controllers.GetEnv("dlna"))
	if err != nil {
		log.Fatal(err)
	}

	var count int
	for _, f := range files {
		if !f.IsDir() {
			re := regexp.MustCompile(`(.mkv|.mp4|.avi|.flv)`)
			if re.MatchString(filepath.Ext(f.Name())) {
				count++
			}
		}
	}
	return count, files
}

func boucleFiles(files []os.FileInfo) {
	log.Println("Démarrage du tri !")
	for _, f := range files {
		if !f.IsDir() {
			log.Println("Movies : " + f.Name())
			controllers.Process(f.Name())
		}
	}
	log.Println("Tri terminé !")
}
