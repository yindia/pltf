package config

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAccessLinksUnmarshal_DuplicateKeysAppend(t *testing.T) {
	input := `
readWrite: adminpltfrole
readWrite:
  - userpltfrole
read: auditrole
`

	var links AccessLinks
	if err := yaml.Unmarshal([]byte(input), &links); err != nil {
		t.Fatalf("unexpected error unmarshalling access links: %v", err)
	}

	expectedReadWrite := []string{"adminpltfrole", "userpltfrole"}
	if got := links["readWrite"]; !reflect.DeepEqual(got, expectedReadWrite) {
		t.Fatalf("readWrite links = %v, want %v", got, expectedReadWrite)
	}

	if got := links["read"]; !reflect.DeepEqual(got, []string{"auditrole"}) {
		t.Fatalf("read links = %v, want [auditrole]", got)
	}
}
