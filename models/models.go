package models

// Word model
type Word struct {
	ID       uint   `gorm:"primaryKey"`
	Word     string `gorm:"uniqueIndex;not null"`
	Language string `gorm:"not null;check:language IN ('pl', 'en');index"`
}

// Translation model
type Translation struct {
	ID       uint `gorm:"primaryKey"`
	WordIDPl uint `gorm:"not null; index; uniqueIndex:pl_en_pair"` // Unique pair
	WordIDEn uint `gorm:"not null; index; uniqueIndex:pl_en_pair"` // Unique pair
	WordPl   Word `gorm:"foreignKey:WordIDPl;references:ID;constraint:OnDelete:CASCADE"`
	WordEn   Word `gorm:"foreignKey:WordIDEn;references:ID;constraint:OnDelete:CASCADE"`
}

// Example model
type Example struct {
	ID      uint   `gorm:"primaryKey"`
	WordID  uint   `gorm:"not null; uniqueIndex:wordid_example"`        // unique example - word pair
	Example string `gorm:"unique;not null; uniqueIndex:wordid_example"` // unique exaple - word pair
	Word    Word   `gorm:"foreignKey:WordID;references:ID;constraint:OnDelete:CASCADE"`
}
