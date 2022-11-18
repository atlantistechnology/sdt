/* The purpose of this small program is to canonicalize existing JSON
 * files.  It provides no options and always requires exactly one filename
 * as an argument.  For users with needs beyond a plugin for `sdt`, the
 * tool `jq` is an excellent, well-tested, and powerful superset of this.
 */
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/atlantistechnology/sdt/pkg/utils"
)

func main() {
	var err error
	var body, output []byte
	var data interface{}

	if len(os.Args) != 2 {
		utils.Fail("`%s` requires exactly one filename argument", os.Args[0])
	}
	filename := os.Args[1]
	body, err = os.ReadFile(filename)
	if err != nil {
		utils.Fail("Unable for read file %s", filename)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		utils.Fail("Unable to parse JSON format from %s", filename)
	}
	output, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		utils.Fail("Unable to serialize data from %s as JSON", filename)
	}
	fmt.Println(string(output))
}
