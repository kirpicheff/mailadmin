package api

import (
	"reflect"
	"testing"
)

func TestParseSieveCode(t *testing.T) {
	code1 := `require ["fileinto"];
# rule:[SPAM]
if header :contains "subject" "*** SPAM ***"
{
        fileinto "Junk";
}
`

	expected1 := []Filter{
		{
			Name:     "SPAM",
			MatchAll: true,
			Active:   true,
			Conditions: []Condition{
				{Field: "Subject", Operator: "contains", Value: "*** SPAM ***"},
			},
			Actions: []Action{
				{Type: "fileinto", Target: "Junk"},
			},
		},
	}

	filters1 := parseSieveCode(code1)
	if !reflect.DeepEqual(filters1, expected1) {
		t.Errorf("Test 1 failed. Expected %+v, got %+v", expected1, filters1)
	}

	code2 := ` require ["fileinto"];
# Move spam to spam folder
if header :contains "X-Spam-Status" ["YES"] {
  fileinto "Junk";
  stop;
}
`

	expected2 := []Filter{
		{
			Name:     "Move spam to spam folder",
			MatchAll: true,
			Active:   true,
			Conditions: []Condition{
				{Field: "X-Spam-Flag", Operator: "contains", Value: "YES"},
			},
			Actions: []Action{
				{Type: "fileinto", Target: "Junk"},
			},
		},
	}

	filters2 := parseSieveCode(code2)
	if !reflect.DeepEqual(filters2, expected2) {
		t.Errorf("Test 2 failed. Expected %+v, got %+v", expected2, filters2)
	}
}
