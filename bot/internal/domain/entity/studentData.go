package entity

type StudentData struct {
	Login string `gorm:"login"`
	Fio   string `gorm:"fio"`
}
