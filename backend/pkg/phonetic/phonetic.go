package phonetic

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/vividvilla/metaphone"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var customReplacer = strings.NewReplacer(
	"Ł", "L",
	"ł", "l",
	"ñ", "n",
	"Ñ", "N",
)

func normalizeString(s string) (string, error) {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, err := transform.String(t, s)
	if err != nil {
		return "", err
	}

	normalized = customReplacer.Replace(normalized)

	var result strings.Builder
	for _, r := range normalized {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result.WriteRune(r)
		}
	}

	return result.String(), nil
}

// GenerateKeys normaliza una cadena y genera sus códigos Double Metaphone.
// Para frases de varias palabras, concatena los códigos de cada palabra.
func GenerateKeys(input string) (primary string, secondary string, err error) {
	cleanInput, err := normalizeString(input)
	if err != nil {
		return "", "", fmt.Errorf("error al normalizar la cadena: %w", err)
	}

	if cleanInput == "" {
		return "", "", nil
	}

	primary, secondary = metaphone.DoubleMetaphone(cleanInput)

	// Truncar para asegurar que quepa en el esquema de la BD (VARCHAR 12 o 24)
	const maxLen = 12
	if len(primary) > maxLen {
		primary = primary[:maxLen]
	}
	if len(secondary) > maxLen {
		secondary = secondary[:maxLen]
	}

	return primary, secondary, nil
}

// GenerateKeysForPhrase genera claves fonéticas para frases más largas,
// concatenando los códigos de las palabras individuales.
func GenerateKeysForPhrase(input string) (primary string, secondary string, err error) {
	normalizedInput, err := normalizeString(input)
	if err != nil {
		return "", "", fmt.Errorf("error al normalizar la cadena: %w", err)
	}

	words := strings.Fields(normalizedInput)
	if len(words) == 0 {
		return "", "", nil
	}

	var primaryKeys, secondaryKeys []string
	for _, word := range words {
		p, s := metaphone.DoubleMetaphone(word)
		if p != "" {
			primaryKeys = append(primaryKeys, p)
		}
		if s != "" {
			secondaryKeys = append(secondaryKeys, s)
		}
	}

	primary = strings.Join(primaryKeys, "")
	secondary = strings.Join(secondaryKeys, "")

	// Truncar para asegurar que quepa en el esquema de la BD (VARCHAR 24)
	const maxLen = 24
	if len(primary) > maxLen {
		primary = primary[:maxLen]
	}
	if len(secondary) > maxLen {
		secondary = secondary[:maxLen]
	}

	return primary, secondary, nil
}
