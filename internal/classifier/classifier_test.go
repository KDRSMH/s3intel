package classifier

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		name         string
		filename     string
		wantCategory Category
		minRisk      int
	}{
		{"env dosyası", "config/.env", CategoryEnvConfig, 50},
		{"private key", "keys/id_rsa.pem", CategoryPrivateKey, 50},
		{"credentials dosyası", "aws/credentials", CategoryCredentials, 50},
		{"sql dump", "backups/dump.sql", CategoryDatabase, 40},
		{"backup dosyası", "site-backup.zip", CategoryBackup, 30},
		{"arşiv dosyası", "release.tar.gz", CategoryArchive, 20},
		{"bilinmeyen dosya", "readme.txt", CategoryOther, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cat, risk := Classify(tc.filename)
			if cat != tc.wantCategory {
				t.Errorf("Classify(%q) kategori = %v, beklenen %v", tc.filename, cat, tc.wantCategory)
			}
			if risk < tc.minRisk {
				t.Errorf("Classify(%q) risk = %d, en az %d bekleniyordu", tc.filename, risk, tc.minRisk)
			}
		})
	}
}
