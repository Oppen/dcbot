package zjson

// Marshal and unmarshal zlib-compressed JSON files.
// To avoid file corruption on writes it creates a temporary file and then moves it.
// Files should end with the extension '.zz' so pigz can use them.

import (
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func Decode(r io.Reader, obj interface{}) error {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return err
	}
	defer zr.Close()

	return json.NewDecoder(zr).Decode(obj)
}

func Encode(w io.Writer, obj interface{}) error {
	zw, err := zlib.NewWriterLevel(w, zlib.BestCompression)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(zw).Encode(obj); err != nil {
		return err
	}

	return zw.Close()
}

func Store(path string, obj interface{}) error {
	tmpf, err := ioutil.TempFile("", path + ".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpf.Name())

	if err = Encode(tmpf, obj); err != nil {
		return err
	}

	if err = tmpf.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmpf.Name(), path); err != nil {
		return err
	}

	return nil
}

func Load(path string, obj interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%s: file open: %w", path, err)
	}
	defer f.Close()

	return Decode(f, obj)
}
