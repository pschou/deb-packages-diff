package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
)

var version = "test"
var indexFileName = "Packages.gz"

type pkg_info struct {
	hash    string
	name    string
	size    int
	version string
}

var after *time.Time
var debug *bool

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Debian Package Diff, Version: %s\n\nUsage: %s [options...]\n\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	var newFile = flag.String("new", "NEW_Packages.gz", "The newer Packages.gz file or repodir/ dir for comparison")
	var oldFile = flag.String("old", "OLD_Packages.gz", "The older Packages.gz file or repodir/ dir for comparison")
	var inRepoPath = flag.String("repo", "", "Repo path to use in file list")
	var outputFile = flag.String("output", "-", "Output file for comparison result")
	var showNew = flag.Bool("showAdded", false, "Display packages only in the new list")
	var showOld = flag.Bool("showRemoved", false, "Display packages only in the old list")
	var showCommon = flag.Bool("showCommon", false, "Display packages in both the new and old lists")
	debug = flag.Bool("debug", false, "Turn on debug output for inspection")

	flag.Parse()

	var new_pkg_index = []pkg_info{}
	var old_pkg_index = []pkg_info{}

	if *debug {
		log.Println("loading new")
	}
	if _, isdir := isDirectory(*newFile); *newFile != "" {
		if isdir {
			*newFile = path.Join(*newFile, indexFileName)
		}
		new_pkg_index = loadIndex(*newFile)
	}
	if *debug {
		log.Println("loading old")
	}
	if _, isdir := isDirectory(*oldFile); *oldFile != "" {
		if isdir {
			*oldFile = path.Join(*oldFile, indexFileName)
		}
		old_pkg_index = loadIndex(*oldFile)
	}

	if *debug {
		log.Println("assigning output")
	}
	out := os.Stdout

	if *outputFile != "-" {
		f, err := os.Create(*outputFile)
		check(err)

		defer f.Close()
		out = f
	}

	// initialized with zeros
	newMatched := make([]int8, len(new_pkg_index))
	oldMatched := make([]int8, len(old_pkg_index))

	log.Println("doing matchups")
matchups:
	for iNew, pNew := range new_pkg_index {
		for iOld, pOld := range old_pkg_index {
			//if reflect.DeepEqual(pNew, pOld) {
			if pNew.hash == pOld.hash &&
				pNew.name == pOld.name &&
				pNew.size == pOld.size &&
				pNew.version == pOld.version {
				newMatched[iNew] = 1
				oldMatched[iOld] = 1
				continue matchups
			}
		}
	}

	fmt.Fprintln(out, "# Debian-diff matchup, version:", version)
	fmt.Fprintln(out, "# new:", *newFile, "old:", *oldFile)
	fmt.Fprintln(out, "# repodir:", *inRepoPath)

	var totalSize uint64
	if *showNew {
		for iNew, v := range new_pkg_index {
			if newMatched[iNew] == 0 {
				totalSize += uint64(v.size)
			}
		}
	}
	if *showCommon {
		for iNew, v := range new_pkg_index {
			if newMatched[iNew] == 1 {
				totalSize += uint64(v.size)
			}
		}
	}
	if *showOld {
		for iOld, v := range old_pkg_index {
			if oldMatched[iOld] == 0 {
				totalSize += uint64(v.size)
			}
		}
	}

	fmt.Fprintln(out, "# filelist size:", humanize.Bytes(totalSize))

	startPath := "" // getBottomDir(strings.TrimPrefix(*inRepoPath, "/"), 2)

	if *showNew {
		for iNew, v := range new_pkg_index {
			if newMatched[iNew] == 0 {
				// This package was not seen in OLD
				fmt.Fprintf(out, "%s %d %s/%s\n", v.hash, v.size, startPath, v.name)
			}
		}
	}

	if *showCommon {
		for iNew, v := range new_pkg_index {
			if newMatched[iNew] == 1 {
				// This package was seen in BOTH
				fmt.Fprintf(out, "%s %d %s/%s\n", v.hash, v.size, startPath, v.name)
			}
		}
	}

	if *showOld {
		for iOld, v := range old_pkg_index {
			if oldMatched[iOld] == 0 {
				// This package was not seen in NEW
				fmt.Fprintf(out, "%s %d %s/%s\n", v.hash, v.size, startPath, v.name)
			}
		}
	}
}

func loadIndex(indexPath string) (pkg_index []pkg_info) {
	pkg_index = []pkg_info{}
	fd, err := os.Open(indexPath)
	check(err)

	defer fd.Close()

	gzr, err := gzip.NewReader(fd)
	check(err)

	indexData := bufio.NewScanner(gzr)

	for {
		// This is what the sections of a Packages.gz may look like (format wise)

		// Package: aiccu
		// Priority: optional
		// Section: universe/net
		// Installed-Size: 191
		// Maintainer: Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>
		// Original-Maintainer: Reinier Haasjes <reinier@haasjes.com>
		// Architecture: amd64
		// Version: 20070115-14.1ubuntu3.1
		// Depends: debconf (>= 0.5) | debconf-2.0, upstart-job, libc6 (>= 2.14), libgnutls26 (>= 2.12.6.1-0), debconf, lsb-base, ucf, iputils-ping, iputils-tracepath, iproute
		// Recommends: ntpdate | ntp | time-daemon, bind9-host | dnsutils
		// Filename: pool/universe/a/aiccu/aiccu_20070115-14.1ubuntu3.1_amd64.deb
		// Size: 51220
		// MD5sum: 079e10cb6983b13f0a998079df62135b
		// SHA1: 7f2c6dc25a41c3fc4dbf406ebe81016609dca166
		// SHA256: 2c31e52c6be536f98d7d24793d9c9c92d0a9720030d290c86dbd858b53fac803
		// Description: SixXS Automatic IPv6 Connectivity Client Utility
		// Homepage: http://www.sixxs.net/tools/aiccu/
		// Description-md5: 064dfb516e6eb18f4217214256491d71
		// Bugs: https://bugs.launchpad.net/ubuntu/+filebug
		// Origin: Ubuntu
		// Npp-Applications: 92650c4d-4b8e-4d2a-b7eb-24ecf4f6b63a,
		//                     ec8030f7-c20a-464f-9b0e-13a3a9e97384

		// Read in blob for a package
		var pkgInfo pkg_info
		for indexData.Scan() {
			line := strings.TrimSpace(indexData.Text())
			if strings.TrimSpace(line) == "" {
				break
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			field := parts[0]
			value := strings.TrimSpace(parts[1])

			switch field {
			case "MD5sum":
				pkgInfo.hash = "{md5}" + value
			case "SHA1":
				pkgInfo.hash = "{sha1}" + value
			case "SHA256":
				pkgInfo.hash = "{sha256}" + value
			case "SHA512":
				pkgInfo.hash = "{sha512}" + value
			case "Filename":
				pkgInfo.name = strings.TrimPrefix(value, "/")
			case "Size":
				pkgInfo.size, err = strconv.Atoi(value)
				check(err)
			}

		}
		if *debug {
			log.Println("found package", pkgInfo.name)
		}

		if len(pkgInfo.name) > 0 {
			pkg_index = append(pkg_index, pkgInfo)
		} else {
			break
		}
	}

	return
}

// isDirectory determines if a file represented
// by `path` is a directory or not
func isDirectory(path string) (exist bool, isdir bool) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, false
	}
	return true, fileInfo.IsDir()
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func getBottomDir(dirName string, n int) string {
	tree := []string{dirName}
	for d := dirName; d != "/" && d != "."; d = filepath.Dir(d) {
		tree = append(tree, d)
	}
	if n >= len(tree) {
		return ""
	}
	return tree[len(tree)-n]
}
