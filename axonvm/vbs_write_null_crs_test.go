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

// Writing Null outputs nothing in IIS VBScript; it must not render the text "Null".
func TestVBSWriteNullOutputsNothing(t *testing.T) {
	cases := map[string]string{
		`<% x = Null %>[<%=x%>]`:            "[]",
		`<% x = Null %>[<%=x%><br/><%=x%>]`: "[<br/>]",
		`[<%=Null%>]`:                       "[]",
		`<% x = Null : Response.Write "[" : Response.Write x : Response.Write "]" %>`: "[]",
		`<% x = Null : Response.Write "[" & x & "]" %>`:                               "[]",
		`<% x = Null : Response.Write("[" & x & "]") %>`:                              "[]",
		`<% x = Null : y = "[" & x & "]" : Response.Write y %>`:                       "[]",
		`<% Response.Write TypeName(Null) & IsNull(Null) %>`:                          "NullTrue",
	}
	for src, want := range cases {
		if got := runVBSAndGetOutput(t, src); got != want {
			t.Errorf("%s => %q, want %q", src, got, want)
		}
	}
}
