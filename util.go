package yiap

import (
	"io/ioutil"
	"path"
	"runtime"
	"strings"
)

func getFixtureWithPath(_path string) string {
	_, filename, _, _ := runtime.Caller(1)
	estimateddAbsoluteFilePath := path.Join(path.Dir(filename))

	projectRootPathGuess := strings.Split(estimateddAbsoluteFilePath, "/")
	dotDotzCount := 0
	for _, e := range projectRootPathGuess {
		if e == "yiap" {
			break
		}
		dotDotzCount++
	}
	dotDotzCount = len(projectRootPathGuess) - dotDotzCount - 1

	dotString := ""
	for i := 0; i < dotDotzCount; i++ {
		dotString += "../"
	}

	data, err := ioutil.ReadFile(path.Join(path.Dir(filename), dotString+"/test_helpers/fixtures/"+_path))
	checkErr(err)
	return string(data)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
