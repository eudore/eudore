package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/eudore/eudore"
)

func main() {
	out := ReadOut()
	ChecktRace(out)

	fmt.Print("::group::Coverage\r\n")
	for _, pkg := range strings.Split(eudore.GetAnyDefault(os.Getenv("PACKAGES"), "github.com/eudore/eudore,github.com/eudore/eudore/middleware,github.com/eudore/eudore/policy"), ",") {
		ChecktCoverage(out, pkg)
	}
	fmt.Print("::endgroup::\r\n")
}

func ReadOut() []byte {
	out, err := os.ReadFile(eudore.GetAnyDefault(os.Getenv("ACTION_OUTPUT"), "output"))
	if err != nil {
		panic(err)
	}
	return out
}

func ChecktCoverage(data []byte, pkg string) {
	reg := regexp.MustCompile(fmt.Sprintf("\t%s\tcoverage: ([0-9\\.]+)%% of statements in", pkg))
	matchs := reg.FindSubmatch(data)
	if len(matchs) == 2 {
		cov, _ := strconv.ParseFloat(string(matchs[1]), 64)
		matchs[0] = bytes.ReplaceAll(matchs[0], []byte{'\t'}, nil)
		switch {
		case cov >=99:
			fmt.Printf("::notice::%s \r\n", matchs[0])
		case cov >= 90:
			fmt.Printf("::warning::%s \r\n", matchs[0])
		default:
			fmt.Printf("::error::%s \r\n", matchs[0])
		}
	}
}

func ChecktRace(data []byte) {
	reg := regexp.MustCompile("==================\\sWARNING: DATA RACE(?sU)(.*)==================")
	matchs := reg.FindAllSubmatch(data, -1)
	if len(matchs) > 0 {
		fmt.Printf("::error::DATA RACE\r\n")
		fmt.Print("::group::Data Race\r\n")
		defer fmt.Print("::endgroup::\r\n")
		for i := range matchs {
			fmt.Println(string(matchs[i][0]))
		}
	}
}
