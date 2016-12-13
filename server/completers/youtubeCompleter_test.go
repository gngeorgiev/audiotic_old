package completers

import "testing"

var completer = youtubeCompleter{}

func TestFormatIsCorrect(t *testing.T) {
	d, err := completer.Complete("youtube")
	if err != nil {
		t.Fatal(err)
	}

	v1, isFirstElementString := d[0].(string)
	v2, isSecondElementString := d[1].(string)
	if !isFirstElementString || !isSecondElementString {
		t.Fatalf("Incorrect autocomplete values %s %s %s", v1, v2, d)
	}
}

func TestAutocompletesYoutube(t *testing.T) {
	d, err := completer.Complete("youtube")
	if err != nil {
		t.Fatal(err)
	}

	if len(d) != 10 {
		t.Fatalf("Incorrect autocomplete values %+v", d)
	}
}
