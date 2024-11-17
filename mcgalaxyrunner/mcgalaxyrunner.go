package mcgalaxyrunner

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed MCGalaxy.zip
var mcGalaxyZip []byte

func unpackFiles(wd string) error {
	log.Println("Unpacking MCGalaxy files")

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

func runServer(wd string, cancel context.CancelFunc, c context.Context, wg *sync.WaitGroup) (func(string), error) {
	monoExec, err := exec.LookPath("mono")

	if err != nil {
		log.Println("Could not find mono runtime for MCGalaxy server:", err)
		return nil, err
	}

	log.Println("Starting up MCGalaxyCLI server...")

	cmd := exec.Command(monoExec, "MCGalaxyCLI.exe", "/?")
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

	err = unpackFiles(wd)
	if err != nil {
		return nil, err
	}

	cmd, err := runServer(wd, cancel, c, wg)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
