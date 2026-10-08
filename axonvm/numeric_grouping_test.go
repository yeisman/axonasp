package axonvm

import "testing"

// FormatNumber output carries grouping commas; classic VBScript (en-US) still treats it as a number.
func TestVBScriptArithmeticAcceptsGroupedNumericStrings(t *testing.T) {
	tests := []struct{ source, want string }{
		{`<% p = FormatNumber(20000/13,0) : p = p+1 : Response.Write p %>`, "1539"},
		{`<% Response.Write "1,538" * 13 %>`, "19994"},
		{`<% Response.Write 20000 - "1,538" %>`, "18462"},
		{`<% Response.Write CDbl("1,234.5") + 1 %>`, "1235.5"},
		{`<% Response.Write CLng("12,345") %>`, "12345"},
		{`<% Response.Write "1,234.5" + 1 %>`, "1235.5"},
	}
	for _, tt := range tests {
		if got := runVBSAndGetOutput(t, tt.source); got != tt.want {
			t.Errorf("%s => %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestVBScriptAddNonNumericStringStillTypeMismatch(t *testing.T) {
	compiler := NewASPCompiler(`<% x = "abc" + 1 %>`)
	if err := compiler.Compile(); err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	vm := NewVMFromCompiler(compiler)
	vm.SetHost(NewMockHost())
	if err := vm.Run(); err == nil {
		t.Fatal("expected Type mismatch")
	}
}
