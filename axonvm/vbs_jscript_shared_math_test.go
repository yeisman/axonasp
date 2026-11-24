package axonvm

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

import "testing"

// TestVBSUsesJScriptMathFromServerBlock checks that VBScript can call the JScript
// Math object when the page has a server-side JScript block, matching names
// without case (Math.Ceil), while other VBScript names are unaffected.
func TestVBSUsesJScriptMathFromServerBlock(t *testing.T) {
	src := `<script language="javascript" runat="server">
function jsTwice(x) { return x * 2; }
</script>
<%
Dim Number
Response.Write jsTwice(21) & "|" & Math.Ceil(25 / 10) & "|" & Math.floor(3.7) & "|" & Math.Max(3, 9) & "|"
Response.Write IsEmpty(Number) & "|" & UBound(Array(1, 2)) & "|" & String(2, "x")
%>`
	got := runVBSAndGetOutput(t, src)
	want := "42|3|3|9|True|1|xx"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestVBSMathWithoutJScriptBlockIsObjectRequired checks that a page without a
// server-side JScript block does not see the JScript Math object.
func TestVBSMathWithoutJScriptBlockIsObjectRequired(t *testing.T) {
	src := `<%
On Error Resume Next
x = Math.Ceil(2.5)
Response.Write "" & (Err.Number <> 0) & "|" & IsEmpty(x)
%>`
	got := runVBSAndGetOutput(t, src)
	want := "True|True"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
