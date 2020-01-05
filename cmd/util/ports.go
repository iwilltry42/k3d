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
	"net"
	"strconv"
	"strings"

	k3d "github.com/rancher/k3d/pkg/types"
	log "github.com/sirupsen/logrus"
)

// ParseAPIPort parses/validates a string to create an exposeAPI struct from it
func ParseAPIPort(portString string) (k3d.ExposeAPI, error) {

	var exposeAPI k3d.ExposeAPI

	split := strings.Split(portString, ":")
	if len(split) > 2 {
		log.Errorln("Failed to parse API Port specification")
		return exposeAPI, fmt.Errorf("api-port format error")
	}

	if len(split) == 1 {
		exposeAPI = k3d.ExposeAPI{Port: split[0]}
	} else {
		// Make sure 'host' can be resolved to an IP address
		addrs, err := net.LookupHost(split[0])
		if err != nil {
			return exposeAPI, err
		}
		exposeAPI = k3d.ExposeAPI{Host: split[0], HostIP: addrs[0], Port: split[1]}
	}

	// Verify 'port' is an integer and within port ranges
	p, err := strconv.Atoi(exposeAPI.Port)
	if err != nil {
		return exposeAPI, err
	}

	if p < 0 || p > 65535 {
		log.Errorln("Failed to parse API Port specification")
		return exposeAPI, fmt.Errorf("port value '%d' out of range", p)
	}

	return exposeAPI, nil

}

// ValidatePortMap validates a port mapping
func ValidatePortMap(portmap string) (string, error) {
	return portmap, nil // TODO: add validation of port mapping
}
