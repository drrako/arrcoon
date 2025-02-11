package arrs

import "regexp"

const AUTH_HEADER = "X-Api-Key"

// function that checks if string a valid torrent hash value
func isValidTorrentHash(hash string) bool {
	// SHA-1 hash (40-character hexadecimal)
	hexPattern := "^[0-9a-fA-F]{40}$"
	// Base32 torrent hash (32-character)
	base32Pattern := "^[A-Z2-7]{32}$"

	hexMatch, _ := regexp.MatchString(hexPattern, hash)
	base32Match, _ := regexp.MatchString(base32Pattern, hash)

	return hexMatch || base32Match
}
