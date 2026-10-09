package repository

import "errors"

var (
	ErrNotFound        = errors.New("data tidak ditemukan")
	ErrDuplicate       = errors.New("data sudah ada")
	ErrAlreadyEnrolled = errors.New("sudah pernah diambil pada tahun akademik ini")
	ErrKuotaPenuh      = errors.New("kuota mata kuliah penuh")
	ErrSKSExceed       = errors.New("total SKS melebihi batas")
)

func BatasSKS(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}
