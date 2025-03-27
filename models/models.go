package models

type Word struct {
	ID       uint   `gorm:"primaryKey"`
	Word     string `gorm:"unique;not null"`
	Language string `gorm:"not null;check:language IN ('pl', 'en')"`
}

type Translation struct {
	ID       uint `gorm:"primaryKey"`
	WordIDPl uint `gorm:"not null"`
	WordIDEn uint `gorm:"not null"`
	WordPl   Word `gorm:"foreignKey:WordIDPl;constraint:OnDelete:CASCADE"`
	WordEn   Word `gorm:"foreignKey:WordIDEn;constraint:OnDelete:CASCADE"`
}

type Example struct {
	ID      uint   `gorm:"primaryKey"`
	WordID  uint   `gorm:"not null"`
	Example string `gorm:"unique;not null"`
	Word    Word   `gorm:"foreignKey:WordID;constraint:OnDelete:CASCADE"`
}
