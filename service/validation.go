package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"api-siakad/app/model"
)

var (
	emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	nimRe   = regexp.MustCompile(`^\d{12}$`)
	taRe    = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)
)

func ValidateLogin(r model.LoginRequest) map[string]string {
	errs := map[string]string{}
	if !emailRe.MatchString(r.Email) {
		errs["email"] = "Format email tidak valid"
	}
	if len(r.Password) < 8 {
		errs["password"] = "Password minimal 8 karakter"
	}
	return errs
}

func ValidateCreateStudent(r model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}
	if !nimRe.MatchString(r.NIM) {
		errs["nim"] = "NIM harus 12 digit angka"
	}
	if strings.TrimSpace(r.Nama) == "" {
		errs["nama"] = "wajib diisi"
	}
	if !emailRe.MatchString(r.Email) {
		errs["email"] = "Format email tidak valid"
	}
	if strings.TrimSpace(r.Prodi) == "" {
		errs["prodi"] = "wajib diisi"
	}
	tahunNow := time.Now().Year()
	if r.Angkatan < 2000 || r.Angkatan > tahunNow {
		errs["angkatan"] = "Angkatan harus 4 digit dan <= tahun berjalan (" + strconv.Itoa(tahunNow) + ")"
	}
	if r.IPKTerakhir < 0 || r.IPKTerakhir > 4.0 {
		errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
	}
	return errs
}

func ValidateUpdateStudent(r model.UpdateStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(r.Nama) == "" {
		errs["nama"] = "wajib diisi"
	}
	if strings.TrimSpace(r.Prodi) == "" {
		errs["prodi"] = "wajib diisi"
	}
	tahunNow := time.Now().Year()
	if r.Angkatan < 2000 || r.Angkatan > tahunNow {
		errs["angkatan"] = "Angkatan harus 4 digit dan <= tahun berjalan (" + strconv.Itoa(tahunNow) + ")"
	}
	if r.IPKTerakhir < 0 || r.IPKTerakhir > 4.0 {
		errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
	}
	return errs
}

func ValidateCreateEnrollment(r model.CreateEnrollmentRequest) map[string]string {
	errs := map[string]string{}
	if r.CourseID <= 0 {
		errs["course_id"] = "wajib diisi dan harus > 0"
	}
	if !taRe.MatchString(r.TahunAkademik) {
		errs["tahun_akademik"] = `Format harus "YYYY/YYYY-Ganjil" atau "YYYY/YYYY-Genap"`
	}
	return errs
}
