package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type StudentResponse struct {
	ID          int     `json:"id"`
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type CreateStudentRequest struct {
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Email       string  `json:"email"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type UpdateStudentRequest struct {
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type ListStudentQuery struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan int
	Search   string
	Sort     string
}

func (q ListStudentQuery) Offset() int { return (q.Page - 1) * q.PerPage }

type StudentDetailResponse struct {
	Student  StudentResponse  `json:"student"`
	Courses  []CourseResponse `json:"courses"`
	TotalSKS int              `json:"total_sks"`
	BatasSKS int              `json:"batas_sks"`
	SisaSKS  int              `json:"sisa_sks"`
}
