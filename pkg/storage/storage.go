package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// UploadToSupabase mengunggah file form-data (multipart) ke Supabase Storage via REST API.
// Mengembalikan URL publik dari file yang diunggah atau error jika gagal.
func UploadToSupabase(fileHeader *multipart.FileHeader) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	supabaseBucket := os.Getenv("SUPABASE_BUCKET")

	if supabaseURL == "" || supabaseKey == "" || supabaseBucket == "" {
		return "", fmt.Errorf("konfigurasi supabase belum lengkap di .env")
	}

	// Buka file yang diunggah
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %w", err)
	}
	defer file.Close()

	// Baca isi file ke memory
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file: %w", err)
	}

	// Bersihkan trailing slash pada URL jika ada
	supabaseURL = strings.TrimSuffix(supabaseURL, "/")

	// Tentukan contentType berdasarkan MIME asli
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Buat nama file menjadi unik (tambahkan timestamp) agar tidak ditolak jika ada nama duplikat
	timestamp := time.Now().UnixMilli()
	uniqueFilename := fmt.Sprintf("%d_%s", timestamp, fileHeader.Filename)
	
	// Encode nama file agar aman dari spasi dan karakter spesial di URL
	safeFilename := url.PathEscape(uniqueFilename)

	// Buat request HTTP ke Supabase
	// Supabase API format: POST /storage/v1/object/[bucket_name]/[file_name]
	uploadEndpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, supabaseBucket, safeFilename)
	req, err := http.NewRequest("POST", uploadEndpoint, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("gagal membuat request HTTP: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", contentType)

	// Lakukan request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal mengeksekusi request ke supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gagal mengunggah file, status: %d, response: %s", resp.StatusCode, string(respBody))
	}

	// Buat URL publik
	// Supabase API format: GET /storage/v1/object/public/[bucket_name]/[file_name]
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, supabaseBucket, safeFilename)
	
	return publicURL, nil
}
