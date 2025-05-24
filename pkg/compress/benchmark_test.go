package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkWriter_Write_Compressed(b *testing.B) {
	sizes := []int{100, 1024, 1024 * 1024} // 100 байт, 1 КБ, 1 МБ
	for _, size := range sizes {
		b.Run(fmt.Sprintf("DataSize=%d", size), func(b *testing.B) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				cw := NewCompressWriter(w)
				cw.Header().Set("Content-Type", "text/html")
				cw.WriteHeader(http.StatusOK)
				_, err := cw.Write(data)
				if err != nil {
					b.Fatal(err)
				}
				err = cw.Close()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkWriter_Write_Uncompressed(b *testing.B) {
	sizes := []int{100, 1024, 1024 * 1024} // 100 байт, 1 КБ, 1 МБ
	for _, size := range sizes {
		b.Run(fmt.Sprintf("DataSize=%d", size), func(b *testing.B) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				cw := NewCompressWriter(w)
				cw.Header().Set("Content-Type", "image/png") // Тип, не поддерживающий сжатие
				cw.WriteHeader(http.StatusOK)
				_, err := cw.Write(data)
				if err != nil {
					b.Fatal(err)
				}
				err = cw.Close()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkReader_Read(b *testing.B) {
	sizes := []int{100, 1024, 1024 * 1024} // 100 байт, 1 КБ, 1 МБ
	for _, size := range sizes {
		b.Run(fmt.Sprintf("DataSize=%d", size), func(b *testing.B) {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}
			_, err := zw.Write(data)
			if err != nil {
				b.Fatal(err)
			}
			err = zw.Close()
			if err != nil {
				b.Fatal(err)
			}
			compressedData := buf.Bytes()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := io.NopCloser(bytes.NewReader(compressedData))
				cr, err := NewCompressReader(r)
				if err != nil {
					b.Fatal(err)
				}
				_, err = io.Copy(io.Discard, cr)
				if err != nil {
					b.Fatal(err)
				}
				err = cr.Close()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
