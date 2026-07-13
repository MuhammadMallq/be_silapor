package model

// Response adalah format standar semua response API
// Semua endpoint mengembalikan JSON dengan format ini
type Response struct {
	Message string      `json:"message" example:"Detail pesan dari server"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty" example:"Detail error jika ada"`
}

// Struct khusus untuk Response 200 OK (Swagger Documentations)
type Response200 struct {
	Message string      `json:"message" example:"Berhasil memproses permintaan"`
	Data    interface{} `json:"data,omitempty"`
}

// Struct khusus untuk Response 201 Created
type Response201 struct {
	Message string      `json:"message" example:"Data berhasil dibuat"`
	Data    interface{} `json:"data,omitempty"`
}

// Struct khusus untuk Response 400 Bad Request
type Response400 struct {
	Message string `json:"message" example:"Bad Request - Validasi gagal"`
	Error   string `json:"error,omitempty" example:"Field required"`
}

// Struct khusus untuk Response 401 Unauthorized
type Response401 struct {
	Message string `json:"message" example:"Unauthorized - Token JWT tidak valid atau tidak ditemukan"`
}

// Struct khusus untuk Response 403 Forbidden
type Response403 struct {
	Message string `json:"message" example:"Forbidden - Anda tidak memiliki akses ke resource ini"`
}

// Struct khusus untuk Response 404 Not Found
type Response404 struct {
	Message string `json:"message" example:"Not Found - Data tidak ditemukan"`
}

// Struct khusus untuk Response 500 Internal Server Error
type Response500 struct {
	Message string `json:"message" example:"Internal Server Error - Terjadi kesalahan pada server"`
	Error   string `json:"error,omitempty" example:"Database connection failed"`
}
