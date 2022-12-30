package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

const (
	ItemTypeTextFile  = 0
	ItemTypeDirectory = 1
	// todo: add more
)

func main() {
	if err := do(); err != nil {
		log.Fatal(err)
	}
}

func do() error {
	ex, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "getting executable")
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)

	hostFlagPtr := flag.String("host", "localhost", "host to serve from, defaults to localhost")
	portFlagPtr := flag.Int("port", 1234, "port to run the gopher server on, defaults to 1234")
	serveDirFlagPtr := flag.String(
		"directory",
		exPath+"/sample",
		"path of directory to serve, defaults to sample folder in CWD",
	)

	flag.Parse()

	address := *hostFlagPtr + ":" + strconv.Itoa(*portFlagPtr)
	serveDir := *serveDirFlagPtr

	log.Println("listening on: ", address)
	log.Println("serving dir: ", serveDir)

	l, err := net.Listen("tcp", address)
	if err != nil {
		return errors.Wrapf(err, "listening at %s", address)
	}

	defer l.Close()

	for {
		c, err := l.Accept()

		if err != nil {
			return errors.Wrap(err, "accepting conn")
		}

		// TODO
		// setup sigterm/int/etc handler to stop accepting new conns and wait XXX time
		// for active conns to close
		// https://pkg.go.dev/os/signal#example_Notify
		go handleConnection(c, serveDir, *portFlagPtr, *hostFlagPtr)
	}
}

func handleConnection(c net.Conn, serveDir string, port int, host string) {
	defer c.Close()

	scanner := bufio.NewScanner(c)

	scanner.Split(carriageReturnLineFeedSplitter)
	scanner.Scan()
	input := scanner.Bytes() // get bytes out after scanning

	if len(input) == 0 {
		entries, err := fs.ReadDir(os.DirFS(serveDir), ".")

		if err != nil {
			log.Println(err)
			// TODO: write back to client with err instead off exiting
			os.Exit(1)
		}

		gopherString := buildGopherStringForDirectoryEntries(entries, "", port, host)
		c.Write([]byte(string(gopherString))) // respond to the client

		return
	}

	// selector ends in "/", directory lookup
	if len(input) > 0 && input[len(input)-1] == '/' {
		// this is probably extremely unsafe and not correctly sand boxed
		// 0:input.len - 1 because we don't want the trailing slash in the lookup
		entries, err := fs.ReadDir(os.DirFS(serveDir), string(input[0:len(input)-1]))

		if err != nil {
			log.Println(err)
			// TODO: write back to client with err instead off exiting
			os.Exit(1)
		}

		gopherString := buildGopherStringForDirectoryEntries(entries, string(input), port, host)
		c.Write([]byte(string(gopherString))) // respond to the client

		return
	}

	// by this point, it must be a file - read directly
	file, err := fs.ReadFile(os.DirFS(serveDir), string(input))

	if err != nil {
		log.Println(err)
		// TODO: write back to client with err instead off exiting
		os.Exit(1)
	}

	c.Write(file)
}

func carriageReturnLineFeedSplitter(data []byte, atEof bool) (
	advance int, token []byte, err error,
) {
	if atEof && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		if j := bytes.IndexByte(data, '\n'); j == (i + 1) {
			// j-1 here so we drop the \r\n bytes.
			return j + 1, data[0 : j-1], nil
		}
	}

	// bombadillo gopher client doesn't give you \r\n, only gives you \n...
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0:i], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEof {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}

func buildGopherStringForDirectoryEntries(
	directoryEntries []fs.DirEntry, selectorBase string, port int, host string,
) string {
	gopherString := ""

	for _, de := range directoryEntries {
		// append item type
		if de.Type().IsRegular() {
			gopherString += strconv.Itoa(ItemTypeTextFile)
		} else {
			gopherString += strconv.Itoa(ItemTypeDirectory)
		}

		gopherString += de.Name() // append name
		gopherString += "\t"      // append tab

		selector := selectorBase

		if de.Type().IsRegular() {
			selector += de.Name()
		} else {
			selector += de.Name()
			selector += "/"
		}
		gopherString += selector           // append selector
		gopherString += "\t"               // append tab
		gopherString += host               // append hostname
		gopherString += "\t"               // append tab
		gopherString += strconv.Itoa(port) // append port
		gopherString += "\r\n"             // append new line!
	}

	gopherString += "." // append deliminating period

	return gopherString
}
