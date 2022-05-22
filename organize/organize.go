package organize

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evanoberholster/imagemeta"
)

// Library object
type Library struct {
	InPath  string
	OutPath string
	Topic   string
}

// Init variables
func (lib *Library) Init() {
	log.Println("Initializing variables")
	fmt.Println("Source photos from PATH:")
	fmt.Scan(&lib.InPath)
	lib.InPath = strings.TrimSuffix(lib.InPath, string(os.PathSeparator))
	fmt.Println("Export photos to PATH:")
	fmt.Scan(&lib.OutPath)
	lib.OutPath = strings.TrimSuffix(lib.OutPath, string(os.PathSeparator))
	fmt.Println("Topic of the processed photos (e.g., location, event):")
	fmt.Scan(&lib.Topic)
	log.Println("Variables initialized")
}

// Validate library
func (lib *Library) Validate() {
	log.Println("Validating library")
	_, err := os.Stat(lib.InPath)
	if os.IsNotExist(err) {
		log.Fatalf("InPath does not exist: %s", lib.InPath)
	}
	_, err = os.Stat(lib.OutPath)
	if os.IsNotExist(err) {
		log.Fatalf("OutPath does not exist: %s", lib.OutPath)
	}
	f := func(r rune) bool {
		return r < 'A' || r > 'z'
	}
	if strings.IndexFunc(lib.Topic, f) != -1 {
		log.Fatalf("Topic should only contain ASCII characters: %s", lib.Topic)
	}
	log.Println("Library validated")
}

// getDateCreated of photo or video
func getDateCreated(path string) (DateTime time.Time) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	var extension string = strings.ToLower(filepath.Ext(f.Name()))
	if extension == ".jpg" || extension == ".heic" {
		m, _ := imagemeta.Parse(f)
		e, _ := m.Exif()
		if e != nil {
			DateTime, _ = e.DateTime(time.Local)
		}
	}
	if extension == ".mov" {
		const epochAdjust = 2082844800
		var i int64
		buf := [8]byte{}
		for {
			_, err := f.Read(buf[:])
			if err != nil {
				log.Println(err)
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
			log.Println(err)
		}
		s := string(buf[4:8])
		switch s {
		case "mvhd":
			if _, err := f.Seek(4, 1); err != nil {
				log.Println(err)
			}
			_, err = f.Read(buf[:4])
			if err != nil {
				log.Println(err)
			}
			i = int64(binary.BigEndian.Uint32(buf[:4]))
			DateTime = time.Unix(i-epochAdjust, 0).Local()
		case "cmov":
			log.Println("moov atom is compressed")
		default:
			log.Println("expected to find 'mvhd' header, didn't")
		}
	}
	return
}

// Process library
func (lib *Library) Process() {
	log.Println("Processing library")
	filepath.WalkDir(lib.InPath, func(oldPath string, info os.DirEntry, err error) error {
		if err != nil {
			log.Println(err.Error())
		}
		if !info.IsDir() {
			log.Printf("Original File Path: %s\n", oldPath)
			var dateCreated time.Time = getDateCreated(oldPath)
			f, err := os.OpenFile(oldPath, os.O_RDONLY, 0644)
			if err != nil {
				log.Println(err)
			}
			defer f.Close()
			var newPath string = lib.OutPath + string(os.PathSeparator) + dateCreated.Format("2006") + "-" + lib.Topic + string(os.PathSeparator) + dateCreated.Format("2006_01_02-Monday")
			var newFile string = lib.Topic + "_" + dateCreated.Format("20060102_150405") + strings.ToLower(filepath.Ext(f.Name()))
			log.Printf("New File Path: %s\n", newPath+string(os.PathSeparator)+newFile)
			os.MkdirAll(newPath, 0750)
			n, err := os.Create(newPath + string(os.PathSeparator) + newFile)
			if err != nil {
				log.Println(err)
			}
			defer n.Close()
			_, err = io.Copy(n, f)
			if err != nil {
				log.Println(err)
			}
		}
		return nil
	})
	log.Println("Library processed")
}
