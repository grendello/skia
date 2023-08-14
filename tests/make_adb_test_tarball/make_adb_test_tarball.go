// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This program is used to create a tarball containing a precompiled Android test. This tarball
// can be passed to the adb_test rule to run the test on an Android device via adb.
//
// Design notes: Currently, building Go with Bazel on Mac can be slow. The problem is compounded
// by the fact that we roll the Skia Infra repo into Skia via a go.mod update, which busts Bazel's
// repository cache and causes Bazel to re-download a larg number of Go modules. To mitigate
// slowness on Mac, this program does not use any external dependencies. This by itself does not
// necessarily make the build faster on Macs, but it unblocks the following potential optimization.
// We could build this binary using a separate, minimalistic go.mod file that does not include the
// Skia Infra repository. If the rules_go[1] rules don't allow multiple go.mod files, we could work
// work around that limitation by shelling out to the Bazel-downloaded "go" binary from a genrule
// (something like "go build -o foo foo.go").
//
// [1] https://github.com/bazelbuild/rules_go

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	execpathsFlag := flag.String("execpaths", "", "Space-separated list of the execpaths of files to be included in the tarball.")
	rootpathsFlag := flag.String("rootpaths", "", "Space-separated list of the rootpaths of files to be included in the tarball.")
	outputFileFlag := flag.String("output-file", "", "Filename of the tarball to create.")
	flag.Parse()

	if *execpathsFlag == "" {
		die("Flag --execpaths is required.\n")
	}
	if *rootpathsFlag == "" {
		die("Flag --rootpaths is required.\n")
	}
	if *outputFileFlag == "" {
		die("Flag --output-file is required.\n")
	}

	execpaths := flagToStrings(*execpathsFlag)
	rootpaths := flagToStrings(*rootpathsFlag)

	if len(execpaths) != len(rootpaths) {
		die("Flags --execpaths and --rootpaths were passed lists of different lenghts: %d and %d.\n", len(execpaths), len(rootpaths))
	}

	outputFile, err := os.Create(*outputFileFlag)
	if err != nil {
		die("Could not create file %q: %s", *outputFileFlag, err)
	}
	defer outputFile.Close()

	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for i := range execpaths {
		// execpaths point to physical files generated by Bazel (e.g.
		// bazel-out/k8-linux_x64-dbg/bin/tests/some_test), whereas rootpaths are the paths that a
		// binary running via "bazel run" or "bazel test" expects (e.g tests/some_test). Thus, we must
		// map the former to the latter.
		//
		// Reference:
		// https://bazel.build/reference/be/make-variables#predefined_label_variables
		if err := addFileToTarball(tarWriter, execpaths[i], rootpaths[i]); err != nil {
			die("Adding file %q to tarball: %s", execpaths[i], err)
		}
	}
}

func flagToStrings(flag string) []string {
	var values []string
	for _, value := range strings.Split(flag, " ") {
		values = append(values, strings.TrimSpace(value))
	}
	return values
}

func addFileToTarball(w *tar.Writer, readFromPath, saveAsPath string) error {
	contents, err := os.ReadFile(readFromPath)
	if err != nil {
		return err
	}

	stat, err := os.Stat(readFromPath)
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name: saveAsPath,
		Size: stat.Size(),
		Mode: int64(stat.Mode()),
	}
	if err := w.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(w, bytes.NewBuffer(contents))
	return err
}

func die(msg string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, a...)
	os.Exit(1)
}
