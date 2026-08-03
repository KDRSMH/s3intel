package cmd

import "testing"

func TestRootCommand_HasActiveAndPassiveSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		names[c.Name()] = true
	}
	if !names["active"] {
		t.Error("root komutunda 'active' alt komutu bulunamadı")
	}
	if !names["passive"] {
		t.Error("root komutunda 'passive' alt komutu bulunamadı")
	}
}

func TestActiveCommand_RequiresBucketFlag(t *testing.T) {
	activeBucket = ""
	if err := activeCmd.RunE(activeCmd, nil); err == nil {
		t.Error("--bucket verilmeden active komutu hata döndürmeliydi")
	}
}

func TestPassiveCommand_RequiresKeywordFlag(t *testing.T) {
	passiveKeyword = ""
	if err := passiveCmd.RunE(passiveCmd, nil); err == nil {
		t.Error("--keyword verilmeden passive komutu hata döndürmeliydi")
	}
}
