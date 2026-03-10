package main

import (
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
