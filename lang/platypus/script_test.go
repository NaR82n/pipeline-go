package platypus

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	ret, retErr := NewScripts(map[string]string{
		"abc.p": "if true {}",
	}, lang.WithNS(constants.NSGitRepo),
		lang.WithCat(point.Logging))

	if len(retErr) > 0 {
		t.Fatal(retErr)
	}

	s := ret["abc.p"]

	if ng := s.Engine(); ng == nil {
		t.Fatalf("no engine")
	}
	plpt := ptinput.NewPlPt(point.Logging, "ng", nil, nil, time.Now())
	err := s.Run(plpt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, plpt.Fields(), map[string]interface{}{"status": constants.DefaultStatus})
	assert.Equal(t, 0, len(plpt.Tags()))
	assert.Equal(t, "abc.p", s.Name())
	assert.Equal(t, point.Logging, s.Category())
	assert.Equal(t, s.NS(), constants.NSGitRepo)

	//nolint:dogsled
	plpt = ptinput.NewPlPt(point.Logging, "ng", nil, nil, time.Now())
	err = s.Run(plpt, nil, &lang.LogOption{DisableAddStatusField: true})
	if err != nil {
		t.Fatal(err)
	}

	if len(plpt.Fields()) != 1 {
		t.Fatal(plpt.Fields())
	} else {
		if _, ok := plpt.Fields()["status"]; !ok {
			t.Fatal("without status")
		}
	}

	//nolint:dogsled
	plpt = ptinput.NewPlPt(point.Logging, "ng", nil, nil, time.Now())
	err = s.Run(plpt, nil, &lang.LogOption{
		DisableAddStatusField: false,
		IgnoreStatus:          []string{constants.DefaultStatus},
	})
	if err != nil {
		t.Fatal(err)
	}
	if plpt.Dropped() != true {
		t.Fatal("!drop")
	}
}

func TestDrop(t *testing.T) {
	ret, retErr := NewScripts(map[string]string{
		"abc.p": "add_key(a, \"a\"); add_key(status, \"debug\"); drop(); add_key(b, \"b\")"},
		lang.WithNS(constants.NSGitRepo),
		lang.WithCat(point.Logging))
	if len(retErr) > 0 {
		t.Fatal(retErr)
	}

	s := ret["abc.p"]

	plpt := ptinput.NewPlPt(point.Logging, "ng", nil, nil, time.Now())
	if err := s.Run(plpt, nil, nil); err != nil {
		t.Fatal(err)
	}

	if plpt.Dropped() != true {
		t.Error("drop != true")
	}
}
