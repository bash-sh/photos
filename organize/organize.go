package organize

import (
	"bufio"
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
	"github.com/evanoberholster/imagemeta/exif"
	"github.com/evanoberholster/imagemeta/meta"
	"github.com/evanoberholster/imagemeta/xmp"
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
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Source photos from PATH: ")
	lib.InPath, _ = reader.ReadString('\n')
	lib.InPath = strings.TrimSuffix(lib.InPath, "\n")
	lib.InPath = strings.TrimSuffix(lib.InPath, string(os.PathSeparator))
	fmt.Print("Export photos to PATH: ")
	lib.OutPath, _ = reader.ReadString('\n')
	lib.OutPath = strings.TrimSuffix(lib.OutPath, "\n")
	lib.OutPath = strings.TrimSuffix(lib.OutPath, string(os.PathSeparator))
	fmt.Print("Topic of the processed photos (e.g., location, event): ")
	lib.Topic, _ = reader.ReadString('\n')
	lib.Topic = strings.TrimSuffix(lib.Topic, "\n")
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
	log.Println("Successfully validated library")
}

// getDateCreated of photo or video
func getDateCreated(f *os.File) (DateTime time.Time) {
	var extension string = strings.ToLower(filepath.Ext(f.Name()))
	var err error
	if extension == ".jpg" || extension == ".heic" {
		var x xmp.XMP
		var e *exif.Data
		exifDecodeFn := func(r io.Reader, m *meta.Metadata) error {
			e, err = e.ParseExifWithMetadata(f, m)
			return nil
		}
		xmpDecodeFn := func(r io.Reader, m *meta.Metadata) error {
			x, err = xmp.ParseXmp(r)
			return err
		}
		_, err := imagemeta.NewMetadata(f, xmpDecodeFn, exifDecodeFn)
		if err != nil {
			log.Fatal(err)
		}
		DateTime, _ = e.DateTime()
	}
	if extension == ".mov" {
		const epochAdjust = 2082844800
		var i int64
		buf := [8]byte{}
		for {
			_, err := f.Read(buf[:])
			if err != nil {
				log.Fatal(err)
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
			log.Fatal(err)
		}
		s := string(buf[4:8])
		switch s {
		case "mvhd":
			if _, err := f.Seek(4, 1); err != nil {
				log.Fatal(err)
			}
			_, err = f.Read(buf[:4])
			if err != nil {
				log.Fatal(err)
			}
			i = int64(binary.BigEndian.Uint32(buf[:4]))
			DateTime = time.Unix(i-epochAdjust, 0).Local()
		case "cmov":
			log.Fatal("moov atom is compressed")
		default:
			log.Fatal("expected to find 'mvhd' header, didn't")
		}
	}
	return
}

// Process library
func (lib *Library) Process() {
	log.Println("Processing library")
	filepath.WalkDir(lib.InPath, func(oldPath string, info os.DirEntry, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		if !info.IsDir() {
			log.Printf("Original File Path: %s\n", oldPath)
			f, err := os.Open(oldPath)
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				err = f.Close()
				if err != nil {
					log.Fatal(err)
				}
			}()
			var dateCreated time.Time = getDateCreated(f)
			var newPath string = lib.OutPath + string(os.PathSeparator) + dateCreated.Format("2006") + "-" + lib.Topic + string(os.PathSeparator) + dateCreated.Format("2006_01_02-Monday")
			var newFile string = lib.Topic + "_" + dateCreated.Format("20060102_150405") + strings.ToLower(filepath.Ext(f.Name()))
			log.Printf("New File Path: %s\n", newPath+string(os.PathSeparator)+newFile)
			os.MkdirAll(newPath, 0750)
			n, err := os.Create(newPath + string(os.PathSeparator) + newFile)
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				err = n.Close()
				if err != nil {
					log.Fatal(err)
				}
			}()
			io.Copy(n, f)
		}
		return nil
	})
}
