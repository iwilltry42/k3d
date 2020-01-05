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
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	k3d "github.com/rancher/k3d/pkg/types"

	"regexp"
)

// Regexp pattern to match node filters
var filterRegexp = regexp.MustCompile(`^(?P<group>master|worker|all)(?P<subsetSpec>\[(?P<subset>(?P<subsetList>(\d+,?)+)|(?P<subsetRange>\d*:\d*)|(?P<subsetWildcard>\*))\])?$`)

// SplitFiltersFromFlag separates a flag's value from the node filter, if there is one
func SplitFiltersFromFlag(flag string) (string, []string, error) {

	/* Case 1) no filter specified */

	if !strings.Contains(flag, "@") {
		return flag, nil, nil
	}

	/* Case 2) filter indicated using '@' in flag */

	split := strings.Split(flag, "@")

	// max number of pieces after split = 2 (only one @ allowed in flag)
	if len(split) > 2 {
		return "", nil, fmt.Errorf("Invalid flag '%s': only one '@' for node filter allowed", flag)
	}

	// trailing or leading '@'
	if len(split) < 2 {
		return "", nil, fmt.Errorf("Invalid flag '%s' includes '@' but is missing either an object or a filter", flag)
	}

	return split[0], strings.Split(split[1], ";"), nil

}

// FilterNodes takes a string filter to return a filtered list of nodes
func FilterNodes(nodes []*k3d.Node, filters []string) ([]*k3d.Node, error) {

	if len(filters) == 0 || len(filters[0]) == 0 {
		log.Warnln("No filter specified")
		return nodes, nil
	}

	// map roles to subsets
	masterNodes := []*k3d.Node{}
	workerNodes := []*k3d.Node{}
	for _, node := range nodes {
		if node.Role == k3d.MasterRole {
			masterNodes = append(masterNodes, node)
		} else if node.Role == k3d.WorkerRole {
			workerNodes = append(workerNodes, node)
		}
	}

	filteredNodes := []*k3d.Node{}
	set := make(map[*k3d.Node]struct{})

	// range over all instances of group[subset] specs
	for _, filter := range filters {

		// match regex with capturing groups
		match := filterRegexp.FindStringSubmatch(filter)

		if len(match) == 0 {
			return nil, fmt.Errorf("Failed to parse node filters: invalid format or empty subset in '%s'", filter)
		}

		// map capturing group names to submatches
		submatches := mapSubexpNames(filterRegexp.SubexpNames(), match)

		// if one of the filters is 'all', we only return this and drop all others
		if submatches["group"] == "all" {
			// TODO: only log if really more than one is specified
			log.Warnf("Node filter 'all' set, but more were specified in '%+v'", filters)
			return nodes, nil
		}

		// Choose the group of nodes to operate on
		groupNodes := []*k3d.Node{}
		if submatches["group"] == string(k3d.MasterRole) {
			groupNodes = masterNodes
		} else if submatches["group"] == string(k3d.WorkerRole) {
			groupNodes = workerNodes
		}

		/* Option 1) subset defined by list */
		if submatches["subsetList"] != "" {
			for _, index := range strings.Split(submatches["subsetList"], ",") {
				if index != "" {
					num, err := strconv.Atoi(index)
					if err != nil {
						return nil, fmt.Errorf("Failed to convert subset number to integer in '%s'", filter)
					}
					if num < 0 || num >= len(groupNodes) {
						return nil, fmt.Errorf("Index out of range: index '%d' < 0 or > number of available nodes in filter '%s'", num, filter)
					}
					if _, exists := set[groupNodes[num]]; !exists {
						filteredNodes = append(filteredNodes, groupNodes[num])
						set[groupNodes[num]] = struct{}{}
					}
				}
			}

			/* Option 2) subset defined by range */
		} else if submatches["subsetRange"] != "" {

			/*
			 * subset specified by a range 'START:END', where each side is optional
			 */

			split := strings.Split(submatches["subsetRange"], ":")
			if len(split) != 2 {
				return nil, fmt.Errorf("Failed to parse subset range in '%s'", filter)
			}

			start := 0
			end := len(groupNodes) - 1

			var err error

			if split[0] != "" {
				start, err = strconv.Atoi(split[0])
				if err != nil {
					return nil, fmt.Errorf("Failed to convert subset range start to integer in '%s'", filter)
				}
				if start < 0 || start >= len(groupNodes) {
					return nil, fmt.Errorf("Invalid subset range: start < 0 or > number of available nodes in '%s'", filter)
				}

			}

			if split[1] != "" {
				end, err = strconv.Atoi(split[1])
				if err != nil {
					return nil, fmt.Errorf("Failed to convert subset range start to integer in '%s'", filter)
				}
				if end < start || end >= len(groupNodes) {
					return nil, fmt.Errorf("Invalid subset range: end < start or > number of available nodes in '%s'", filter)
				}
			}

			for i := start; i <= end; i++ {
				if _, exists := set[groupNodes[i]]; !exists {
					filteredNodes = append(filteredNodes, groupNodes[i])
					set[groupNodes[i]] = struct{}{}
				}
			}

			/* Option 3) subset defined by wildcard */
		} else if submatches["subsetWildcard"] == "*" {
			/*
			 * '*' = all nodes
			 */
			for _, node := range groupNodes {
				if _, exists := set[node]; !exists {
					filteredNodes = append(filteredNodes, node)
					set[node] = struct{}{}
				}
			}

			/* Option X) invalid/unknown subset */
		} else {
			return nil, fmt.Errorf("Failed to parse node specifiers: unknown subset in '%s'", filter)
		}

	}

	return filteredNodes, nil
}
