package store

import (
	"crypto/rand"
	"math/big"
)

var avatarIcons = []string{
	"Zap", "Star", "Flame", "Shield", "Crown", "Trophy", "Target", "Rocket",
	"Ghost", "Cat", "Dog", "Bird", "Leaf", "Sun", "Moon", "Snowflake",
	"Mountain", "Waves", "Music", "Heart", "Smile", "Fish", "Swords", "Dumbbell",
	"Bike", "Footprints",
}

func randomAvatarIcon() string {
	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(avatarIcons))))
	return avatarIcons[idx.Int64()]
}

func randomAvatarColor() string {
	return "forest"
}

// codeAlpha is uppercase + digits with visually ambiguous chars removed
// (0, 1, I, O) so codes are easy to read aloud and type.
const codeAlpha = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func newID() string {
	b := make([]byte, 4)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlpha))))
		b[i] = codeAlpha[idx.Int64()]
	}
	return string(b)
}

func newAdminToken() string {
	return "tok_" + randString(32)
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b[i] = alphabet[idx.Int64()]
	}
	return string(b)
}
