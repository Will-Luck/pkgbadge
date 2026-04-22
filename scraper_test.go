package main

import (
	"fmt"
	"os"
	"testing"
)

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("load fixture %s: %v", name, err)
	}
	return string(data)
}

func TestParsePackagePage_TotalPulls(t *testing.T) {
	html := loadFixture(t, "package-page.html")
	stats, err := parsePackagePage(html, "Will-Luck", "docker-sentinel")
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalPulls != 433 {
		t.Errorf("TotalPulls = %d, want 433", stats.TotalPulls)
	}
}

func TestParsePackagePage_LatestVersion(t *testing.T) {
	html := loadFixture(t, "package-page.html")
	stats, err := parsePackagePage(html, "Will-Luck", "docker-sentinel")
	if err != nil {
		t.Fatal(err)
	}
	if stats.LatestVersion != "2.11.1" {
		t.Errorf("LatestVersion = %q, want %q", stats.LatestVersion, "2.11.1")
	}
}

func TestParsePackagePage_Architectures(t *testing.T) {
	html := loadFixture(t, "package-page.html")
	stats, err := parsePackagePage(html, "Will-Luck", "docker-sentinel")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"linux/amd64", "linux/arm64"}
	if len(stats.Architectures) != len(want) {
		t.Fatalf("Architectures = %v, want %v", stats.Architectures, want)
	}
	for i, a := range stats.Architectures {
		if a != want[i] {
			t.Errorf("Architectures[%d] = %q, want %q", i, a, want[i])
		}
	}
}

func TestParsePackagePage_NoData(t *testing.T) {
	_, err := parsePackagePage("<html><body>no data here</body></html>", "x", "y")
	if err == nil {
		t.Fatal("expected error for page with no extractable data")
	}
}

func TestParsePackagePage_AbbreviatedDownloads(t *testing.T) {
	// Once a package crosses ~1,000 downloads GitHub renders the text as
	// "1.32K" but keeps the exact integer in the title attribute.
	// See GiteaLN/pkgbadge scraper bug: 0-pull badges for iplayer-arr and
	// docker-sentinel even though both packages have four-figure pull counts.
	html := `<span class="text-normal h2 mr-1 color-fg-muted" >v1.2.3</span>` +
		`<span class="d-block color-fg-muted text-small tmp-mb-1">Total downloads</span>` +
		`<h3 title="1318">1.32K</h3>`
	stats, err := parsePackagePage(html, "Will-Luck", "iplayer-arr")
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalPulls != 1318 {
		t.Errorf("TotalPulls = %d, want 1318", stats.TotalPulls)
	}
}

func TestParsePackages(t *testing.T) {
	tests := []struct {
		input string
		want  int
		err   bool
	}{
		{"Will-Luck/docker-sentinel", 1, false},
		{"Will-Luck/Docker-Sentinel/docker-sentinel", 1, false},
		{"Will-Luck/docker-sentinel, Will-Luck/Docker-Guardian/docker-guardian", 2, false},
		{"", 0, false},
		{"a/b/c/d", 0, true},
	}
	for _, tt := range tests {
		refs, err := parsePackages(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("parsePackages(%q) error = %v, wantErr %v", tt.input, err, tt.err)
			continue
		}
		if len(refs) != tt.want {
			t.Errorf("parsePackages(%q) = %d refs, want %d", tt.input, len(refs), tt.want)
		}
	}
}

func TestParsePackages_ThreePart(t *testing.T) {
	refs, err := parsePackages("Will-Luck/Docker-Sentinel/docker-sentinel")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 {
		t.Fatalf("got %d refs, want 1", len(refs))
	}
	ref := refs[0]
	if ref.Owner != "Will-Luck" {
		t.Errorf("Owner = %q, want %q", ref.Owner, "Will-Luck")
	}
	if ref.Repo != "Docker-Sentinel" {
		t.Errorf("Repo = %q, want %q", ref.Repo, "Docker-Sentinel")
	}
	if ref.Package != "docker-sentinel" {
		t.Errorf("Package = %q, want %q", ref.Package, "docker-sentinel")
	}
}

func TestParseManifestSizes_Index(t *testing.T) {
	indexBody := loadFixture(t, "manifest-index.json")
	amd64Body := loadFixture(t, "manifest-amd64.json")
	arm64Body := loadFixture(t, "manifest-arm64.json")

	fetcher := func(digest string) ([]byte, error) {
		switch digest {
		case "sha256:amd64digest":
			return []byte(amd64Body), nil
		case "sha256:arm64digest":
			return []byte(arm64Body), nil
		default:
			return nil, fmt.Errorf("unexpected digest: %s", digest)
		}
	}

	sizes, err := parseManifestSizes([]byte(indexBody), fetcher)
	if err != nil {
		t.Fatal(err)
	}
	if len(sizes) != 2 {
		t.Fatalf("got %d platforms, want 2", len(sizes))
	}
	if sizes["linux/amd64"] != 10000000 {
		t.Errorf("amd64 = %d, want 10000000", sizes["linux/amd64"])
	}
	if sizes["linux/arm64"] != 9000000 {
		t.Errorf("arm64 = %d, want 9000000", sizes["linux/arm64"])
	}
}

func TestParseManifestSizes_Single(t *testing.T) {
	body := loadFixture(t, "manifest-single.json")

	sizes, err := parseManifestSizes([]byte(body), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sizes) != 1 {
		t.Fatalf("got %d entries, want 1", len(sizes))
	}
	if sizes[""] != 12000000 {
		t.Errorf("size = %d, want 12000000", sizes[""])
	}
}
