// Package classifier, bir dosya adına/uzantısına bakarak (içeriğini hiç
// okumadan) kaba bir kategori ve temel risk puanı atar. Hem aktif modda
// (indirilen nesneler için) hem de pasif modda (grayhatwarfare sonuçları
// için, içerik hiç indirilmediği için) kullanılır.
package classifier

import "strings"

// Category, bir dosyanın ait olduğu kaba risk kategorisidir.
type Category string

const (
	CategoryEnvConfig   Category = "env_config"
	CategoryPrivateKey  Category = "private_key"
	CategoryCredentials Category = "credentials"
	CategoryDatabase    Category = "database"
	CategoryBackup      Category = "backup"
	CategoryArchive     Category = "archive"
	CategoryOther       Category = "other"
)

type rule struct {
	patterns []string
	category Category
	baseRisk int
}

// rules, sırayla kontrol edilir; ilk eşleşen desen kazanır. Bu yüzden daha
// spesifik/riskli desenler listede daha önce yer alır.
var rules = []rule{
	{[]string{".env"}, CategoryEnvConfig, 70},
	{[]string{".pem", ".key", "id_rsa", "id_ed25519"}, CategoryPrivateKey, 85},
	{[]string{"credential", ".aws/credentials"}, CategoryCredentials, 80},
	{[]string{".sql", ".dump", ".db", ".sqlite"}, CategoryDatabase, 60},
	{[]string{"backup", ".bak", ".old"}, CategoryBackup, 55},
	{[]string{".zip", ".tar", ".gz", ".7z", ".rar"}, CategoryArchive, 30},
}

// Classify, dosya adına/uzantısına göre kategori ve 0-100 aralığında temel
// bir risk puanı döner. Hiçbir desen eşleşmezse CategoryOther ve düşük bir
// temel puan döner.
func Classify(filename string) (Category, int) {
	lower := strings.ToLower(filename)
	for _, r := range rules {
		for _, pattern := range r.patterns {
			if strings.Contains(lower, pattern) {
				return r.category, r.baseRisk
			}
		}
	}
	return CategoryOther, 10
}
