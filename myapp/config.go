package myapp

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Config :
type Config struct {
	Dlna   string `json:"dlna"`
	Series string `json:"series"`
	Movies string `json:"movies"`
	Mail   string `json:"mail"`
	// Music  string `json:"music"`
}

const (
	// ConfigFile :
	FOLDER_CONFIG = "./searchMoviesConfig"
	LOGFILE       = FOLDER_CONFIG + string(os.PathSeparator) + "log_SearchAndSort"
	MOVIESFILE    = FOLDER_CONFIG + string(os.PathSeparator) + ".movies.json"
	SERIESFILE    = FOLDER_CONFIG + string(os.PathSeparator) + ".series.json"
	ConfigFile    = FOLDER_CONFIG + string(os.PathSeparator) + ".config.json"
)

// GetEnv :
func GetEnv(key string) string {
	checkIfConfigFileIsExist()

	jsonType := readJSONFile(ConfigFile)

	return jsonType[key].(string)
}

// SetEnv :
func SetEnv(key, value string) {
	checkIfConfigFileIsExist()
	// open file using READ & WRITE permission
	jsonType := readJSONFile(ConfigFile)

	jsonType[key] = value

	j, err := json.MarshalIndent(jsonType, "", " ")
	if err != nil {
		log.Println(err)
	}

	writeJSONFile(ConfigFile, j)
}

// CheckIfConfigFileIsExist : Create file is not exist
func checkIfConfigFileIsExist() {
	if _, err := os.Stat(FOLDER_CONFIG); os.IsNotExist(err) {
		os.Mkdir(FOLDER_CONFIG, os.ModeSticky|0755)
	}

	// detect if file exists
	var _, err = os.Stat(ConfigFile)

	// create file if not exists
	if os.IsNotExist(err) {
		newJSON := &Config{
			Dlna:   "", //pwd("dlna", true),
			Series: "", //pwd("dlna/Series", true),
			Movies: "", //pwd("dlna/Movies", true),
			// Music:  pwd("dlna/Music", true),
		}
		j, err := json.MarshalIndent(newJSON, "", " ")
		if err != nil {
			log.Println(err)
		}

		writeJSONFile(ConfigFile, j)
	}
}

// ReadJSONFile :
func readJSONFile(file string) map[string]interface{} {
	f, err := ioutil.ReadFile(file)

	if err != nil {
		log.Println(err)
	}
	var jsonType map[string]interface{}

	json.Unmarshal(f, &jsonType)

	return jsonType
}

func writeJSONFile(file string, jsonByte []byte) {
	err := ioutil.WriteFile(file, jsonByte, 0644)
	// file, err := os.Create(ConfigFile)
	if err != nil {
		log.Println(err)
	}
}

func pwd(name string, endPathSeparator bool) string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	if endPathSeparator {
		return filepath.Clean(pwd+string(os.PathSeparator)+name) + string(os.PathSeparator)
	}
	return filepath.Clean(pwd + string(os.PathSeparator) + name)
}
