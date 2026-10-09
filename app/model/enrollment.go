package model

import "time"

type Enrollment struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id"`
	CourseID      int       `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateEnrollmentRequest struct {
	CourseID      int    `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}

type EnrollmentWithCourse struct {
	Enrollment Enrollment
	Course     Course
}
