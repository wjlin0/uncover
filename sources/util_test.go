package sources

import "testing"

func TestMatchSubdomains(t *testing.T) {
	tests := "https://nuclei.projectdiscovery.io/templating-guide/protocols/http"
	domain := "projectdiscovery.io"
	submatch := MatchSubdomains(domain, tests, true)
	if len(submatch) == 0 {
		t.Error("MatchSubdomains failed")
	}
	t.Log(submatch)

}
