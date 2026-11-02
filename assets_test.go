package fayna

import (
	"io/fs"
	"os"
	"regexp"
	"testing"
)

// matrix.html loads these by URL, and CopyNamespacedAssets serves them from AssetsFS at
// boot. A script that is left out of the commit breaks the drawer without a build error.
func TestAssetsFS_HoldsEveryScriptTheMatrixLoads(t *testing.T) {
	const page = "domain/operation/outcome_matrix/templates/matrix.html"
	html, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	refs := regexp.MustCompile(`/assets/js/fayna/([a-z0-9-]+\.js)`).FindAllStringSubmatch(string(html), -1)
	seen := map[string]bool{}
	for _, m := range refs {
		seen[m[1]] = true
		if _, err := fs.Stat(AssetsFS, "assets/js/"+m[1]); err != nil {
			t.Errorf("%s loads %s, which is not embedded: %v", page, m[1], err)
		}
	}
	if !seen["outcome-matrix-note-pick.js"] {
		t.Errorf("%s does not load outcome-matrix-note-pick.js (found %v)", page, seen)
	}
}
