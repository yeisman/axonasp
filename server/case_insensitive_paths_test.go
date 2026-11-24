/*
 * AxonASP Server
 * Copyright (C) 2026 G3pix Ltda. All rights reserved.
 *
 * Developed by Yuri Eisman (@yeisman)
 * Contact: https://g3pix.com.br
 * Project URL: https://g3pix.com.br/axonasp
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * Attribution Notice:
 * If this software is used in other projects, the name "AxonASP Server"
 * must be cited in the documentation or "About" section.
 *
 * Contribution Policy:
 * Modifications to the core source code of AxonASP Server must be
 * made available under this same license terms.
 */
package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupCaseInsensitiveRoot creates a web root with mixed-case names and points the
// server globals at it; the returned function restores the previous values.
func setupCaseInsensitiveRoot(t *testing.T, enabled bool) (string, func()) {
	t.Helper()
	root := t.TempDir()
	write := func(rel string, body string) {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	write("Scripts/appscripts.js", "var app = 1;")
	write("include/ServerScripts.js", "var server = 1;")
	write("Hidden/secret.txt", "secret")
	write("Private/data.config", "secret")

	origRoot := RootDir
	origVirtual := VirtualAppPath
	origWebConfig := activeWebConfig
	origCaseInsensitive := CaseInsensitivePaths
	origBlockedExt := BlockedExtensions
	origBlockedDirs := BlockedDirs
	origBlockedPrefixes := blockedDirPrefixes

	RootDir = root
	VirtualAppPath = ""
	activeWebConfig = nil
	CaseInsensitivePaths = enabled
	BlockedExtensions = []string{".config"}
	BlockedDirs = []string{filepath.Join(root, "Hidden")}
	blockedDirPrefixes = buildBlockedDirPrefixes(BlockedDirs)

	return root, func() {
		RootDir = origRoot
		VirtualAppPath = origVirtual
		activeWebConfig = origWebConfig
		CaseInsensitivePaths = origCaseInsensitive
		BlockedExtensions = origBlockedExt
		BlockedDirs = origBlockedDirs
		blockedDirPrefixes = origBlockedPrefixes
	}
}

// fileSystemIgnoresCase reports whether root lives on a case-insensitive file system
// (the macOS and Windows defaults), where the fallback is never reached.
func fileSystemIgnoresCase(t *testing.T, root string) bool {
	t.Helper()
	probe := filepath.Join(root, "CaseProbe")
	if err := os.WriteFile(probe, nil, 0o644); err != nil {
		t.Fatalf("write probe: %v", err)
	}
	defer os.Remove(probe)
	_, err := os.Stat(filepath.Join(root, "caseprobe"))
	return err == nil
}

func serveTestPath(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handleRequest(rec, req)
	return rec
}

func TestCaseInsensitivePathsServesMiscasedStaticFiles(t *testing.T) {
	_, restore := setupCaseInsensitiveRoot(t, true)
	defer restore()

	cases := map[string]string{
		"/Scripts/appscripts.js":    "var app = 1;",
		"/scripts/appScripts.js":    "var app = 1;",
		"/include/serverScripts.js": "var server = 1;",
		"/INCLUDE/SERVERSCRIPTS.JS": "var server = 1;",
	}
	for path, want := range cases {
		rec := serveTestPath(path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, rec.Code)
		}
		if got := rec.Body.String(); got != want {
			t.Fatalf("%s: expected body %q, got %q", path, want, got)
		}
	}

	if rec := serveTestPath("/scripts/missing.js"); rec.Code != http.StatusNotFound {
		t.Fatalf("missing file: expected 404, got %d", rec.Code)
	}
}

func TestCaseInsensitivePathsKeepsBlockedEntriesHidden(t *testing.T) {
	root, restore := setupCaseInsensitiveRoot(t, true)
	defer restore()
	if fileSystemIgnoresCase(t, root) {
		t.Skip("file system is case-insensitive; the fallback is not reached here")
	}

	for _, path := range []string{"/hidden/secret.txt", "/HIDDEN/Secret.TXT", "/private/DATA.CONFIG"} {
		rec := serveTestPath(path)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s: expected 404, got %d", path, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "secret") {
			t.Fatalf("%s: blocked content leaked: %q", path, rec.Body.String())
		}
	}
}

func TestCaseInsensitivePathsDisabledByDefault(t *testing.T) {
	root, restore := setupCaseInsensitiveRoot(t, false)
	defer restore()
	if fileSystemIgnoresCase(t, root) {
		t.Skip("file system is case-insensitive; the disabled behavior cannot be observed here")
	}
	if rec := serveTestPath("/scripts/appScripts.js"); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 with case_insensitive_paths disabled, got %d", rec.Code)
	}
}

func TestResolveRequestPathCaseInsensitivePrefersExactCasing(t *testing.T) {
	root := t.TempDir()
	if fileSystemIgnoresCase(t, root) {
		t.Skip("file system is case-insensitive; the fallback is not reached here")
	}
	if err := os.MkdirAll(filepath.Join(root, "Assets"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "Assets", "Logo.png"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, ok := resolveRequestPathCaseInsensitive(root, "assets/logo.PNG")
	if !ok {
		t.Fatalf("expected a match")
	}
	if want := filepath.Join(root, "Assets", "Logo.png"); got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
	if _, ok := resolveRequestPathCaseInsensitive(root, "assets/../../etc/passwd"); ok {
		t.Fatalf("expected traversal outside the root to be rejected")
	}
}
