package models

type Word struct {
	ID       uint   `gorm:"primaryKey"`
	Word     string `gorm:"uniqueIndex;not null"`
	Language string `gorm:"not null;check:language IN ('pl', 'en');index"`
}

type Translation struct {
	ID       uint `gorm:"primaryKey"`
	WordIDPl uint `gorm:"not null; index"`
	WordIDEn uint `gorm:"not null; index"`
	WordPl   Word `gorm:"foreignKey:WordIDPl;references:ID;constraint:OnDelete:CASCADE"`
	WordEn   Word `gorm:"foreignKey:WordIDEn;references:ID;constraint:OnDelete:CASCADE"`
}

type Example struct {
	ID      uint   `gorm:"primaryKey"`
	WordID  uint   `gorm:"not null"`
	Example string `gorm:"unique;not null"`
	Word    Word   `gorm:"foreignKey:WordID;references:ID;constraint:OnDelete:CASCADE"`
}
