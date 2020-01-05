/*
Copyright © 2020 The k3d Author(s)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package util

import (
	"fmt"
	"os"
	"strings"
)

// ValidateVolumeMount checks, if the source of volume mounts exists and if the destination is an absolute path
// - SRC: source directory/file -> tests: must exist
// - DEST: source directory/file -> tests: must be absolute path
func ValidateVolumeMount(volumeMount string) (string, error) {
	src := ""
	dest := ""

	// validate 'SRC[:DEST]' substring
	split := strings.Split(volumeMount, ":")
	if len(split) < 1 || len(split) > 2 {
		return "", fmt.Errorf("Invalid volume mount '%s': only one ':' allowed", volumeMount)
	}

	// we only have SRC specified -> DEST = SRC
	if len(split) == 1 {
		src = split[0]
		dest = src
	} else {
		src = split[0]
		dest = split[1]
	}

	// verify that the source exists
	if src != "" {
		if _, err := os.Stat(src); err != nil {
			return "", fmt.Errorf("Failed to stat file/dir that you're trying to mount: '%s' in '%s'", src, volumeMount)
		}
	}

	// verify that the destination is an absolute path
	if !strings.HasPrefix(dest, "/") {
		return "", fmt.Errorf("Volume mount destination doesn't appear to be an absolute path: '%s' in '%s'", dest, volumeMount)
	}

	return fmt.Sprintf("%s:%s", src, dest), nil
}
