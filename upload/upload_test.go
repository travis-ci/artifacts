package upload

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	testTmp, err = ioutil.TempDir("", "artifacts-test-upload")
)

func init() {
	os.Clearenv()

	if err != nil {
		log.Panicf("game over: %v\n", err)
	}
}

func setenvs(e map[string]string) error {
	for k, v := range e {
		err := os.Setenv(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
