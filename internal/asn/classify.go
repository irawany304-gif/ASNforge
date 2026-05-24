package asn

import "strings"

func ClassifyName(name, org string) (string, []string, int) {
	s := strings.ToLower(name + " " + org)
	switch {
	case strings.Contains(s, "google"):
		return TypeCloud, []string{"cloud", "search", "dns"}, 70
	case strings.Contains(s, "cloudflare"):
		return TypeCDN, []string{"cdn", "cloud", "dns", "security"}, 70
	case strings.Contains(s, "amazon") || strings.Contains(s, "aws"):
		return TypeCloud, []string{"cloud", "hosting"}, 70
	case strings.Contains(s, "microsoft") || strings.Contains(s, "azure"):
		return TypeCloud, []string{"cloud"}, 70
	case strings.Contains(s, "university") || strings.Contains(s, "college"):
		return TypeEducation, []string{"education"}, 65
	case strings.Contains(s, "government") || strings.Contains(s, "gov"):
		return TypeGovernment, []string{"government"}, 60
	case strings.Contains(s, "internet exchange") || strings.Contains(s, " ix"):
		return TypeIX, []string{"ix"}, 60
	default:
		return TypeUnknown, nil, 50
	}
}
