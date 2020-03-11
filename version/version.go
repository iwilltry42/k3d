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
package version

import (
	"github.com/heroku/docker-registry-client/registry"
	log "github.com/sirupsen/logrus"
)

// Version is the string that contains version
var Version string

// K3sVersion should contain the latest version tag of k3s (hardcoded at build time)
// we're setting a default version v1.0.0, because it's stable and because the 'latest' tag is not actively maintained
var K3sVersion = "v1.0.0" // TODO: can we try to dynamically fetch the latest version at runtime and only fallback to this if it fails?

// GetVersion returns the version for cli, it gets it from "git describe --tags" or returns "dev" when doing simple go build
func GetVersion() string {
	if len(Version) == 0 {
		return "v3-dev"
	}
	return Version
}

// GetK3sVersion returns the version string for K3s
func GetK3sVersion(latest bool) string {
	if latest {
		version, err := fetchLatestK3sVersion()
		if err != nil || version == "" {
			log.Warnln("Failed to fetch latest K3s version from DockerHub, falling back to hardcoded version.")
			return K3sVersion
		}
		return version
	}
	return K3sVersion
}

// fetchLatestK3sVersion tries to fetch the latest version of k3s from DockerHub
func fetchLatestK3sVersion() (string, error) {

	url := "https://registry-1.docker.io/"
	username := "" // anonymous
	password := "" // anonymous
	repository := "rancher/k3s"

	hub, err := registry.New(url, username, password)
	if err != nil {
		return "", err
	}

	tags, err := hub.Tags(repository)
	if err != nil || len(tags) == 0 {
		return "", err
	}

	log.Debugln("Fetched the following tags for rancher/k3s from DockerHub:")
	log.Debugln(tags)

	return "sampleTag", nil

}
