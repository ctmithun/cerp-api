package psw_generator

import "math/rand"

const (
	charSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+=-"
)

func GeneratePsw() string {

	var password []byte

	for i := 0; i < 12; i++ {
		randNum := rand.Intn(len(charSource))
		password = append(password, charSource[randNum])
	}
	return string(password)
}
