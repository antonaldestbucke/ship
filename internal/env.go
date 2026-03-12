package shipinternal

import "github.com/joho/godotenv"

func LoadDotEnv(path string) error {
	return godotenv.Overload(path)
}
