// from https://github.com/labstack/echo/blob/master/cookbook/file-upload/multiple/server.go

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"net/http"

	rice "github.com/GeertJohan/go.rice"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func upload(c echo.Context) error {
	// Read form fields
	name := c.FormValue("name")
	email := c.FormValue("email")

	//------------
	// Read files
	//------------

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["files"]

	for _, file := range files {
		// Source
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Destination
		dst, err := os.Create(updir + dirsymble + file.Filename)
		if err != nil {
			return err
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			return err
		}

	}

	return c.HTML(http.StatusOK, fmt.Sprintf("<p>Uploaded successfully %d files with fields name=%s and email=%s.</p><p><a href=\"/\">RETURN</a></p>", len(files), name, email))
}

// upload to this directory
var updir string = ""
var dirsymble string = "/"

var port int = 9080

const HOME_UPLOAD_DIR = "${HOME}/upload"

// TODO: https://github.com/labstack/echo/blob/master/cookbook/auto-tls/server.go

func main() {

	var err error
	var homedir string = os.Getenv("HOME")
	var argline string = strings.Join(os.Args[1:], " ")

	if runtime.GOOS == "windows" {
		dirsymble = "\\"
	}

	flag.StringVar(&updir, "w", HOME_UPLOAD_DIR, "directory to write uploaded file")
	flag.IntVar(&port, "p", 9080, "listening port")

	flag.Parse()

	if updir == HOME_UPLOAD_DIR {
		updir = homedir + "/upload"
	}

	err = os.MkdirAll(updir, 0700)
	if err != nil {
		log.Fatalf("change to the upload directory: %s\n", err)
	}

	err = os.Chdir(updir)
	if err != nil {
		log.Fatalf("change to the upload directory: %s\n", err)
	}
	upfile, err := os.Create(updir + "/.uploader.rw.check")
	if err != nil {
		log.Printf("WARNING: write to the upload directory: %s\n", err)
	}
	upfile.Close()
	os.Remove(updir + "/.uploader.rw.check")

	log.Printf("[%s] Listening at %s, upload to %s ...\n", strings.Trim(os.Args[0]+" "+argline, " "), strconv.Itoa(port), updir)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// rice
	assetHandler := http.FileServer(rice.MustFindBox("embedded-static").HTTPBox())
	// serves the index.html from rice
	e.GET("/", echo.WrapHandler(assetHandler))

	e.Static("/upload", updir)
	e.POST("/upload", upload)

	e.Logger.Fatal(e.Start(":" + strconv.Itoa(port)))
}
