package version

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
	"github.com/sirupsen/logrus"
)

type dFormat struct {
}

func (dFormat) Format(e *logrus.Entry) ([]byte, error) {
	return []byte(e.Message), nil
}

func TestLogBuildInformation(t *testing.T) {

	n := time.Now()
	BuildTime = n.String()
	buf := bytes.NewBufferString("")
	logrus.SetFormatter(dFormat{})
	logrus.SetOutput(buf)
	LogBuildInformation("hello")
	assert.Equal(t, buf.String(), fmt.Sprintf("------------------------------------------------------------------------App [hello], Version [unknown], BuildSHA [unknown], BuildTS [%s]------------------------------------------------------------------------", n.String()))
}
