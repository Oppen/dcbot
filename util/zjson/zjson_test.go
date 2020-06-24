package zjson

import (
	"bytes"
	"testing"
)

type TestType struct {
	F1 int     `json:"f1"`
	F2 string  `json:"f2"`
	F3 float64 `json:"f3"`
}

func TestEncodeDecode(t *testing.T) {
	in := TestType{
		F1: 42,
		F2: "EN EL DOOM...",
		F3: 3.14,
	}
	var out TestType

	buf := &bytes.Buffer{}
	if err := Encode(buf, &in); err != nil {
		t.Fatalf("Encoding failed with: %v", err)
	}
	if err := Decode(buf, &out); err != nil {
		t.Fatalf("Decoding failed with: %v", err)
	}
	if in != out {
		t.Fatalf("%+v != %+v", in, out)
	}
}
