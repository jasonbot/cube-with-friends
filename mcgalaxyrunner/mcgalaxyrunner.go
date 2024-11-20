package mcgalaxyrunner

import (
	"archive/zip"
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed MCGalaxy.zip
var mcGalaxyZip []byte

//go:embed game/**
var overlayFiles embed.FS

func unpackServerExecutableFiles(wd string) error {
	log.Println("Unpacking MCGalaxy server files")

	zipReader := bytes.NewReader(mcGalaxyZip)
	reader, err := zip.NewReader(zipReader, int64(len(mcGalaxyZip)))
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(wd, f.FileInfo().Name())

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				log.Fatal(err)
				return err
			}

			_, err := os.Stat(fpath)
			if os.IsNotExist(err) {
				f, err := os.OpenFile(
					fpath,
					os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
					f.Mode(),
				)
				if err != nil {
					return err
				}
				defer f.Close()

				_, err = io.Copy(f, rc)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func unpackOverlayFiles(wd string) error {
	log.Println("Unpacking MCGalaxy overlays")

	err := fs.WalkDir(overlayFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "game/README.txt" {
			return nil
		}

		opath := filepath.Join(wd, strings.TrimPrefix(path, "game/"))

		_, err = os.Stat(opath)
		fileExists := true
		if os.IsNotExist(err) {
			fileExists = false
		}

		if d.Type().IsDir() && !fileExists {
			log.Println("Creating directory for overlay:", path)
			if err := os.MkdirAll(opath, 0755); err != nil {
				return err
			}
		} else if d.Type().IsRegular() && !fileExists {
			log.Println("Writing overlay:", opath)

			data, err := fs.ReadFile(overlayFiles, path)
			if err != nil {
				return err
			}

			if err := os.WriteFile(opath, data, 0755); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func runServer(wd string, cancel context.CancelFunc, c context.Context, wg *sync.WaitGroup) (func(string), error) {
	monoExec, err := exec.LookPath("mono")

	if err != nil {
		log.Println("Could not find mono runtime for MCGalaxy server:", err)
		return nil, err
	}

	log.Println("Starting up MCGalaxyCLI server...")

	cmd := exec.Command(monoExec, "MCGalaxyCLI.exe")
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()

	if err != nil {
		return nil, err
	}

	sendCommand := func(command string) {
		stdin.Write([]byte(fmt.Sprintf("%s\n", command)))
	}

	err = cmd.Start()

	if err != nil {
		log.Println("Can't start MCGalaxy:", err)
		return nil, err
	}

	processdied := make(chan interface{})
	go func() {
		cmd.Wait()
		log.Printf("MCGalaxy process finished")
		processdied <- nil
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if cmd.Err != nil {
			log.Println("Command did not start up:", cmd.Err)
			return
		}

		defer cmd.Process.Kill()
		defer stdin.Close()

		select {
		case <-c.Done():
			break
		case <-processdied:
			log.Fatal("MCGalaxy died randomly?")
			cancel()
			return
		}
		sendCommand("I got shut down.")
		log.Println("Killing MCGalaxy server...")
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}()

	return sendCommand, nil
}

func RunGalaxyServer(cancel context.CancelFunc, c context.Context, wg *sync.WaitGroup) (func(string), error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	wd = filepath.Join(wd, "game")

	err = os.MkdirAll(wd, 0755)
	if err != nil {
		return nil, err
	}

	err = unpackServerExecutableFiles(wd)
	if err != nil {
		return nil, err
	}

	err = unpackOverlayFiles(wd)
	if err != nil {
		return nil, err
	}

	cmd, err := runServer(wd, cancel, c, wg)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
