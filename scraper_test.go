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

func TestParsePackagePage_NoDownloads(t *testing.T) {
	stats, err := parsePackagePage("<html><body>no data here</body></html>", "x", "y")
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalPulls != 0 {
		t.Errorf("TotalPulls = %d, want 0", stats.TotalPulls)
	}
	if stats.LatestVersion != "" {
		t.Errorf("LatestVersion = %q, want empty", stats.LatestVersion)
	}
}
