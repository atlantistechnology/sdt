package ruby

import (
    "log"
	"os/exec"
)

func Diff(filename string) []byte {
	var out []byte
	var err error
    rubyCmd := "ruby" // TODO: Determine executable in more dynamic way
	cmd := exec.Command(rubyCmd, "--dump=parsetree", filename)
	out, err = cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return out


}
