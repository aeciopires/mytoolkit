package jsontoon

import "testing"

func TestConvertTabularArray(t *testing.T) {
	input := `{"users":[{"id":1,"name":"Alice","role":"admin"},{"id":2,"name":"Bob","role":"user"}]}`
	want := "users[2]{id,name,role}:\n  1,Alice,admin\n  2,Bob,user\n"
	got, err := Convert([]byte(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Convert() = %q, want %q", got, want)
	}
}

func TestConvertSimpleObject(t *testing.T) {
	input := `{"id":123,"name":"Ada","active":true}`
	want := "id: 123\nname: Ada\nactive: true\n"
	got, err := Convert([]byte(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Convert() = %q, want %q", got, want)
	}
}

func TestConvertArrayOfArrays(t *testing.T) {
	input := `{"pairs":[[1,2],[3,4]]}`
	want := "pairs[2]:\n  - [2]: 1,2\n  - [2]: 3,4\n"
	got, err := Convert([]byte(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Convert() = %q, want %q", got, want)
	}
}

func TestConvertNonUniformObjectArray(t *testing.T) {
	input := `{"items":[{"a":1},{"b":2}]}`
	got, err := Convert([]byte(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	// Differing key sets must NOT use tabular {a,b} header form.
	if want := "items[2]{"; len(got) >= len(want) && got[:len(want)] == want {
		t.Errorf("Convert() used tabular form for non-uniform objects: %q", got)
	}
}

func TestStringQuoting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", `{"s":""}`, `s: ""` + "\n"},
		{"leading/trailing whitespace", `{"s":" hi "}`, `s: " hi "` + "\n"},
		{"literal true", `{"s":"true"}`, `s: "true"` + "\n"},
		{"literal false", `{"s":"false"}`, `s: "false"` + "\n"},
		{"literal null", `{"s":"null"}`, `s: "null"` + "\n"},
		{"numeric-looking", `{"s":"123"}`, `s: "123"` + "\n"},
		{"contains colon", `{"s":"a:b"}`, `s: "a:b"` + "\n"},
		{"contains bracket", `{"s":"a[b]"}`, `s: "a[b]"` + "\n"},
		{"starts with hyphen", `{"s":"-x"}`, `s: "-x"` + "\n"},
		{"plain word not quoted", `{"s":"hello"}`, "s: hello\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert([]byte(tt.input), Options{})
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Convert() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNumberCanonicalization(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"integer", `{"n":42}`, "n: 42\n"},
		{"float", `{"n":1.50}`, "n: 1.5\n"},
		{"negative zero", `{"n":-0}`, "n: 0\n"},
		{"large magnitude", `{"n":123}`, "n: 123\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert([]byte(tt.input), Options{})
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Convert() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDelimiterOptions(t *testing.T) {
	input := `{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}`

	tab, err := Convert([]byte(input), Options{Delimiter: "tab"})
	if err != nil {
		t.Fatal(err)
	}
	wantTab := "users[2~]{id\tname}:\n  1\tAlice\n  2\tBob\n"
	if tab != wantTab {
		t.Errorf("tab delimiter Convert() = %q, want %q", tab, wantTab)
	}

	pipe, err := Convert([]byte(input), Options{Delimiter: "pipe"})
	if err != nil {
		t.Fatal(err)
	}
	wantPipe := "users[2|]{id|name}:\n  1|Alice\n  2|Bob\n"
	if pipe != wantPipe {
		t.Errorf("pipe delimiter Convert() = %q, want %q", pipe, wantPipe)
	}
}

func TestIndentOption(t *testing.T) {
	input := `{"a":{"b":1}}`
	got, err := Convert([]byte(input), Options{IndentSize: 4})
	if err != nil {
		t.Fatal(err)
	}
	want := "a:\n    b: 1\n"
	if got != want {
		t.Errorf("Convert() = %q, want %q", got, want)
	}
}

func TestConvertErrors(t *testing.T) {
	if _, err := Convert([]byte(""), Options{}); err == nil {
		t.Error("expected error for empty input")
	}
	if _, err := Convert([]byte(`{"a":}`), Options{}); err == nil {
		t.Error("expected error for malformed JSON")
	}
	if _, err := Convert([]byte(`{"a":1}`), Options{Delimiter: "bogus"}); err == nil {
		t.Error("expected error for invalid delimiter")
	}
}

func TestConvertRootScalar(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"42", "42\n"},
		{`"hi"`, "hi\n"},
		{"true", "true\n"},
		{"null", "null\n"},
	}
	for _, tt := range tests {
		got, err := Convert([]byte(tt.input), Options{})
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("Convert(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
