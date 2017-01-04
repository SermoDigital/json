package json

import "testing"

func TestIssue279(t *testing.T) {
	var foo struct {
		F struct {
			M RawMessage
		}
	}
	b, err := Marshal(&foo)
	if err != nil {
		t.Fatal(err)
	}
	sb := string(b)
	const want = "{\"F\":{\"M\":null}}\n"
	if sb != want {
		t.Fatalf("wanted %s, got %s", want, sb)
	}
}
