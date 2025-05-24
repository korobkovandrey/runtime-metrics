package sign

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
)

// BenchmarkMake тестирует производительность функции Make для разных размеров данных.
func BenchmarkMake(b *testing.B) {
	sizes := []int{100, 1024, 1024 * 1024} // 100B, 1KB, 1MB
	key := []byte("secret_key")
	for _, size := range sizes {
		data := make([]byte, size)
		_, err := rand.Read(data)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(fmt.Sprintf("DataSize=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Make(data, key)
			}
		})
	}
}

// BenchmarkEncodeToString тестирует производительность функции EncodeToString.
func BenchmarkEncodeToString(b *testing.B) {
	sizes := []int{32, 64, 128} // Размеры хешей в байтах
	for _, size := range sizes {
		data := make([]byte, size)
		_, err := rand.Read(data)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(fmt.Sprintf("HashSize=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = EncodeToString(data)
			}
		})
	}
}

// BenchmarkDecodeString тестирует производительность функции DecodeString.
func BenchmarkDecodeString(b *testing.B) {
	sizes := []int{64, 128, 256} // Длины строк в символах
	for _, size := range sizes {
		data := make([]byte, size/2) // 2 символа hex = 1 байт
		_, err := rand.Read(data)
		if err != nil {
			b.Fatal(err)
		}
		hexStr := hex.EncodeToString(data)
		b.Run(fmt.Sprintf("StringSize=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := DecodeString(hexStr)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMakeToString тестирует производительность функции MakeToString.
func BenchmarkMakeToString(b *testing.B) {
	sizes := []int{100, 1024, 1024 * 1024} // 100B, 1KB, 1MB
	key := []byte("secret_key")
	for _, size := range sizes {
		data := make([]byte, size)
		_, err := rand.Read(data)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(fmt.Sprintf("DataSize=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = MakeToString(data, key)
			}
		})
	}
}

// BenchmarkValidate тестирует производительность функции Validate.
func BenchmarkValidate(b *testing.B) {
	sizes := []int{100, 1024, 1024 * 1024} // 100B, 1KB, 1MB
	key := []byte("secret_key")
	for _, size := range sizes {
		data := make([]byte, size)
		_, err := rand.Read(data)
		if err != nil {
			b.Fatal(err)
		}
		hash := Make(data, key)
		// Тест с валидным хешем
		b.Run(fmt.Sprintf("DataSize=%d_Valid", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Validate(data, key, hash)
			}
		})
		// Тест с невалидным хешем
		invalidHash := append(hash, 0) // Изменяем хеш
		b.Run(fmt.Sprintf("DataSize=%d_Invalid", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Validate(data, key, invalidHash)
			}
		})
	}
}
