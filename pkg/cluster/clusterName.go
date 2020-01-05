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
package cluster

import (
	"fmt"

	"github.com/rancher/k3d/pkg/types"
)

// CheckName ensures that a cluster name is also a valid host name according to RFC 1123.
// We further restrict the length of the cluster name to maximum 'clusterNameMaxSize'
// so that we can construct the host names based on the cluster name, and still stay
// within the 64 characters limit.
func CheckName(name string) error {
	if err := ValidateHostname(name); err != nil {
		return fmt.Errorf("Invalid cluster name. %+v", ValidateHostname(name))
	}
	if len(name) > types.DefaultClusterNameMaxLength {
		return fmt.Errorf("Cluster name must be <= %d characters, but has %d", types.DefaultClusterNameMaxLength, len(name))
	}
	return nil
}

// ValidateHostname ensures that a cluster name is also a valid host name according to RFC 1123.
func ValidateHostname(name string) error {

	if len(name) == 0 {
		return fmt.Errorf("No name provided")
	}

	if name[0] == '-' || name[len(name)-1] == '-' {
		return fmt.Errorf("Hostname '%s' must not start or end with '-' (dash)", name)
	}

	for _, c := range name {
		switch {
		case '0' <= c && c <= '9':
		case 'a' <= c && c <= 'z':
		case 'A' <= c && c <= 'Z':
		case c == '-':
			break
		default:
			return fmt.Errorf("Hostname '%s' contains characters other than 'Aa-Zz', '0-9' or '-'", name)

		}
	}

	return nil
}
