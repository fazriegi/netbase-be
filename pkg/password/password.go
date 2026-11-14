package password

import "golang.org/x/crypto/bcrypt"

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy_password_for_timing_mitigation"), bcrypt.DefaultCost)

func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func Check(password, hash string) bool {
	if len(password) > 72 {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// DummyCheck performs a constant-work bcrypt comparison against a precomputed hash
// to mitigate username enumeration via timing side-channels when a user does not exist.
func DummyCheck(password string) {
	if len(password) > 72 {
		password = password[:72]
	}
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
}
