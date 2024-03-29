package lib

import (
	"bytes"
	"fmt"
	"github.com/sam-docker/media-organizer/constants"
	"github.com/sam-docker/media-organizer/logger"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	movies = constants.MOVIES
	series = constants.SERIES
)

type typeSerieOrMovie uint

const (
	MOVIE typeSerieOrMovie = iota
	NOTHING
)

type myFile struct {
	file           string
	ext            string
	resolution     string
	fileWithoutDir string
	complete       string
	completeSlug   string
	name           string
	serieName      string
	serieNumber    string
	season         int
	year           int
	episode        int
	episodeRaw     string
	language       string
	duration       float64
	ForceType      string
}

func (m *myFile) Process() {
	_, m.complete = filepath.Split(m.file)
	m.fileWithoutDir = m.complete
	err := m.start(NOTHING)
	if err != nil {
		logger.L(logger.Red, "%s => %s", m.fileWithoutDir, err)
		return
	}
	//logger.L(logger.Yellow, "complete: %s", m.complete)
}

func (m *myFile) start(serieOrMovieOrBoth typeSerieOrMovie) error {
	err := m.slugFile()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if m.serieName == "" || serieOrMovieOrBoth == MOVIE {
		m.isMovie()
	} else {
		m.isSerie()
	}
	return nil
}

func (m *myFile) isMovie() {
	extension := filepath.Ext(m.file)
	// logger.L(logger.Yellow, "name: %s", m.name)

	var path1 string
	m.complete = m.name + extension
	path1 = movies + string(os.PathSeparator) + m.complete

	start := time.Now()
	if moveOrRenameFile(m.file, path1) {
		duration := time.Now().Sub(start)
		logger.L(logger.Green, "Movie: %s has been moved to: %s - %s", m.fileWithoutDir, path1, duration)
	}
}

func (m *myFile) isSerie() {
	m.checkFolderSerie()
}

func (m *myFile) checkFolderSerie() (string, string) {
	// serieName, exist := folderExist(series, serieName)
	ss := func() string {
		if m.season == 0 {
			return "00"
		}
		return oneToNine(m.season)
	}()

	newFolder := string(os.PathSeparator) + m.serieName + string(os.PathSeparator) + "season-" + ss
	folderOk := series + string(os.PathSeparator) + m.serieName
	if _, err := os.Stat(folderOk); os.IsNotExist(err) {
		logger.L(logger.Yellow, "Create folder: "+m.serieName)
		createFolder(folderOk)
	}
	if _, err := os.Stat(series + newFolder); os.IsNotExist(err) {
		logger.L(logger.Yellow, "Create folder : "+newFolder)
		createFolder(series + newFolder)
	}

	finalFilePath := series + newFolder + string(os.PathSeparator) + m.complete
	start := time.Now()
	if moveOrRenameFile(m.file, finalFilePath) {
		duration := time.Now().Sub(start)
		logger.L(logger.Green, "Episode: %s has been moved to: %s - %s", m.fileWithoutDir, finalFilePath, duration)
	}

	return m.complete, finalFilePath
}

func (m *myFile) formatageSerie() {
	format := constants.REGEX_SERIE
	re := regexp.MustCompile(`\{(\w+)}`)
	result := re.ReplaceAllStringFunc(format, func(serie string) string {
		switch serie {
		case "{name}":
			return m.name
		case "{season}":
			return fmt.Sprintf("%s", oneToNine(m.season))
		case "{episode}":
			return fmt.Sprintf("%s", oneToNine(m.episode))
		case "{resolution}":
			if m.resolution == "" {
				return ""
			}
			return m.resolution
		case "{year}":
			if m.year == 0 {
				return ""
			}
			return fmt.Sprintf("%d", m.year)
		case "{language}":
			if m.language == "" {
				return ""
			}
			return m.language
		default:
			return serie
		}
	})

	result = strings.ReplaceAll(result, " - ", " ")
	result = strings.ReplaceAll(result, "- ", " ")
	result = strings.ReplaceAll(result, "()", "")
	result = strings.TrimSpace(result)
	result = strings.TrimSuffix(result, "-")

	m.complete = result + m.ext
	m.name = result
	m.serieNumber = fmt.Sprintf("s%se%s", oneToNine(m.season), oneToNine(m.episode))
}

func oneToNine(number int) string {
	if number < 10 {
		return "0" + strconv.Itoa(number)
	}
	return strconv.Itoa(number)
}

func (m *myFile) formatageMovie() {
	// constants.REGEX_MOVIE
	format := constants.REGEX_MOVIE
	re := regexp.MustCompile(`\{(\w+)}`)
	result := re.ReplaceAllStringFunc(format, func(movie string) string {
		switch movie {
		case "{name}":
			return m.name
		case "{resolution}":
			if m.resolution == "" {
				return ""
			}
			return m.resolution
		case "{year}":
			if m.year == 0 {
				return ""
			}
			return fmt.Sprintf("%d", m.year)
		case "{language}":
			if m.language == "" {
				return ""
			}
			return m.language
		default:
			return movie
		}
	})

	result = strings.ReplaceAll(result, " - ", " ")
	result = strings.ReplaceAll(result, "- ", " ")
	result = strings.ReplaceAll(result, "()", "")
	result = strings.TrimSpace(result)

	m.complete = result + m.ext
	m.name = result
}

func (m *myFile) removeFirstBrackets() {
	// remove first brackets
	re := regexp.MustCompile(`(?mi)^\[(.*?)]`)
	m.name = re.ReplaceAllString(m.name, "")
	m.name = strings.TrimSpace(m.name)
}

func createFolder(folder string) {
	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		logger.L(logger.Red, "%s", err)
	}
}

var mu sync.Mutex

func moveOrRenameFile(filePathOld, filePathNew string) bool {
	mu.Lock()
	defer mu.Unlock()

	filePathOld = filepath.Clean(filePathOld)
	filePathNew = filepath.Clean(filePathNew)

	defer constants.ObsSlice.Remove(filePathOld)

	err := os.Chown(filePathOld, constants.UID, constants.GID)
	if err != nil {
		logger.L(logger.Red, "Failed Chown file => %s", filePathOld)
	}
	// Convertir la chaîne octale en int64
	chmodInt, err := strconv.ParseInt(constants.CHMOD, 8, 64)
	if err != nil {
		logger.L(logger.Red, "Failed to convert octal string to int64")
	}
	err = os.Chmod(filePathOld, os.FileMode(chmodInt))
	if err != nil {
		logger.L(logger.Red, "Failed Chmod file => %s", filePathOld)
	}
	err = os.Rename(filePathOld, strings.ToLower(filePathNew))
	if err != nil {
		logger.L(logger.Red, "Error os.Rename. Test mv => %s", "mv \""+filePathOld+"\" \""+filePathNew+"\"")
		cmd := exec.Command("/bin/sh", "-c", "mv \""+filePathOld+"\" \""+filePathNew+"\"")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			logger.L(logger.Red, "Move Or Rename file: %s, Error: %s", err, stderr.String())
			return false
		}
	}
	logger.L(logger.Yellow, "File Rename => %s", filePathOld)

	folder := filepath.Dir(filePathOld)

	folder = getAbsolutePathWithRelative(folder)
	absoluteATrier := getAbsolutePathWithRelative(constants.BE_SORTED)

	if folder != absoluteATrier {
		file, _ := os.ReadDir(folder)
		if len(file) == 0 {
			err = watch.Remove(folder)
			if err != nil {
				logger.L(logger.Red, "Error. Can't delete watcher to folder: %s", folder)
			}
			logger.L(logger.Yellow, "Delete watcher to folder: %s", folder)
			err := os.Remove(folder)
			if err != nil {
				logger.L(logger.Red, "Error to delete folder: %s", folder)
			}
		}
	}

	return true
}

func getAbsolutePathWithRelative(folder string) string {
	abs, err := filepath.Abs(folder)
	if err == nil {
		return abs
	}
	return ""
}

func CleanFolder(str string) {
	err := filepath.Walk(str, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if path != str {
				err := os.Remove(path)
				if err != nil {
					logger.L(logger.Red, "Remove folder isn't possible because a file(s) is inside : %s", path)
				}
			}
		} else {
			re := regexp.MustCompile(constants.RegexFileExtension)
			if !re.MatchString(filepath.Ext(path)) {
				err := os.Remove(path)
				if err != nil {
					logger.L(logger.Red, "Error to remove file: %s", path)
				}
			}
		}
		return nil
	})
	if err != nil {
		return
	}
}
