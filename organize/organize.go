package organize

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/evanoberholster/imagemeta"
	"github.com/rs/zerolog/log"
)

// Library object
type Library struct {
	InPath         string
	OutPath        string
	Topic          string
	CountProcessed int
}

// Init variables
func (lib *Library) Init() {
	log.Info().Msg("Initializing variables")
	fmt.Println("Source photos from PATH:")
	fmt.Scan(&lib.InPath)
	lib.InPath = strings.TrimSuffix(lib.InPath, string(os.PathSeparator))
	fmt.Println("Export photos to PATH:")
	fmt.Scan(&lib.OutPath)
	lib.OutPath = strings.TrimSuffix(lib.OutPath, string(os.PathSeparator))
	fmt.Println("Topic of the processed photos (e.g., location, event):")
	fmt.Scan(&lib.Topic)
	log.Info().Msg("Variables initialized")
}

// Validate library
func (lib *Library) Validate() {
	log.Info().Msg("Validating library")
	_, err := os.Stat(lib.InPath)
	if os.IsNotExist(err) {
		log.Fatal().Msgf("InPath does not exist: %s", lib.InPath)
	}
	_, err = os.Stat(lib.OutPath)
	if os.IsNotExist(err) {
		log.Fatal().Msgf("OutPath does not exist: %s", lib.OutPath)
	}
	f := func(r rune) bool {
		return r < 'A' || r > 'z'
	}
	if strings.IndexFunc(lib.Topic, f) != -1 {
		log.Fatal().Msgf("Topic should only contain ASCII characters: %s", lib.Topic)
	}
	log.Info().Msg("Library validated")
}

// getDateCreated of photo or video
func getDateCreated(path string) (DateTime time.Time) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal().Err(err).Msgf("OutPath does not exist: %s", path)
	}
	defer f.Close()
	var extension string = strings.ToLower(filepath.Ext(f.Name()))
	if extension == ".jpg" || extension == ".heic" {
		e, err := imagemeta.Decode(f)
		if err != nil {
			log.Fatal().Err(err).Msg("Cannot extract metadata from image")
		} else {
			log.Debug().Str("exif", e.String()).Msg("Metadata extracted from image")
			DateTime = e.CreateDate()
		}
	} else if extension == ".png" {
		e, err := imagemeta.DecodePng(f)
		if err != nil {
			log.Fatal().Err(err).Msg("Cannot extract metadata from image")
		} else {
			log.Debug().Str("exif", e.String()).Msg("Metadata extracted from image")
			DateTime = e.CreateDate()
		}
	} else if extension == ".mov" {
		const epochAdjust = 2082844800
		var i int64
		buf := [8]byte{}
		for {
			_, err := f.Read(buf[:])
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot read file")
			}
			if bytes.Equal(buf[4:8], []byte("moov")) {
				break
			} else {
				atomSize := binary.BigEndian.Uint32(buf[:])
				f.Seek(int64(atomSize)-8, 1)
			}
		}
		_, err = f.Read(buf[:])
		if err != nil {
			log.Fatal().Err(err).Msg("Cannot read file")
		}
		s := string(buf[4:8])
		switch s {
		case "mvhd":
			if _, err := f.Seek(4, 1); err != nil {
				log.Fatal().Err(err).Msg("Cannot read file")
			}
			_, err = f.Read(buf[:4])
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot read file")
			}
			i = int64(binary.BigEndian.Uint32(buf[:4]))
			DateTime = time.Unix(i-epochAdjust, 0).Local()
		case "cmov":
			log.Debug().Msg("moov atom is compressed")
		default:
			log.Debug().Msg("expected to find 'mvhd' header, didn't")
		}
	} else {
		log.Info().Msgf("File extension not supported: %s", extension)
	}
	return
}

// Validate if a file already exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return !info.IsDir()
}

// Process library
func (lib *Library) Process() {
	log.Info().Msg("Processing library")
	lib.CountProcessed = 0
	filepath.WalkDir(lib.InPath, func(oldPath string, info os.DirEntry, err error) error {
		if err != nil {
			log.Fatal().Err(err).Msg("Cannot read file")
		}
		if !info.IsDir() {
			log.Debug().Msgf("Original File Path: %s\n", oldPath)
			var dateCreated time.Time = getDateCreated(oldPath)
			f, err := os.OpenFile(oldPath, os.O_RDONLY, 0644)
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot open file")
			}
			defer f.Close()
			var newPath string = lib.OutPath + string(os.PathSeparator) + dateCreated.Format("2006") + "-" + lib.Topic + string(os.PathSeparator) + dateCreated.Format("2006_01_02-Monday")
			var newFile string = lib.Topic + "_" + dateCreated.Format("20060102_150405.000") + strings.ToLower(filepath.Ext(f.Name()))
			for fileExists(newPath + string(os.PathSeparator) + newFile) {
				newFile = lib.Topic + "_" + dateCreated.Format("20060102_150405.000") + "_" + strconv.Itoa(rand.Int()) + strings.ToLower(filepath.Ext(f.Name()))
			}
			log.Debug().Msgf("New File Path: %s\n", newPath+string(os.PathSeparator)+newFile)
			os.MkdirAll(newPath, 0750)
			n, err := os.Create(newPath + string(os.PathSeparator) + newFile)
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot create file")
			}
			defer n.Close()
			_, err = io.Copy(n, f)
			if err != nil {
				log.Fatal().Err(err).Msg("Cannot copy file")
			} else {
				lib.CountProcessed++
			}
		}
		return nil
	})
	log.Info().Msgf("Processed %s files in library", strconv.Itoa(lib.CountProcessed))
}
