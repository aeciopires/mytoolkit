package hashgen

import "testing"

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		algo    string
		want    string
		wantErr bool
	}{
		{"md5 hello", "hello", "md5", "5d41402abc4b2a76b9719d911017c592", false},
		{"sha1 hello", "hello", "sha1", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d", false},
		{"sha256 hello", "hello", "sha256", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", false},
		{"sha512 hello", "hello", "sha512", "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043", false},
		{"empty md5", "", "md5", "d41d8cd98f00b204e9800998ecf8427e", false},
		{"default algorithm is sha256", "hello", "", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", false},
		{"unsupported algorithm", "hello", "sha1024", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Generate([]byte(tt.input), Options{Algorithm: tt.algo})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("Generate() = %q, want %q", got, tt.want)
			}
		})
	}
}
