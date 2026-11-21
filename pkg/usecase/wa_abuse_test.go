package usecase

import "testing"

func TestContainsAbuse(t *testing.T) {
	blocked := []string{
		"you are a fucking idiot",
		"F*CK this",
		"fuuuuck",
		"f u c k you",
		"tu chutiya hai kya",
		"ch00tiya",
		"madarchod",
		"bhenchod kya bakwas",
		"bsdk",
		"gandu",
		"तू चूतिया है",
		"मादरचोद",
		"bitch",
		"m.a.d.a.r.c.h.o.d",
		"saali randi",
		"cunt",
	}
	for _, s := range blocked {
		if !containsAbuse(s, nil) {
			t.Errorf("expected abuse in %q", s)
		}
	}

	allowed := []string{
		"Is the 1.5 ton split AC available?",
		"chod do, kal baat karte hain",
		"mera kutta bimar hai",
		"bc I was busy, sorry",
		"Charles Dickens book available?",
		"class of 2026",
		"bahut accha, thank you bhai",
		"कल दुकान खुलेगी?",
		"price 1500 for 20 pcs",
		"assist me please",
		"a grandiose shop opening",
		"I live in Scunthorpe",
		"graduated magna cum laude",
		"Brandi will pick it up",
		"",
	}
	for _, s := range allowed {
		if containsAbuse(s, nil) {
			t.Errorf("false positive on %q", s)
		}
	}
}

func TestContainsAbuseExtraWords(t *testing.T) {
	if !containsAbuse("you are a nalayak", []string{"nalayak"}) {
		t.Error("extra word not blocked")
	}
	if containsAbuse("you are a nalayak", nil) {
		t.Error("extra word blocked without config")
	}
}
