package pkg

import (
	"go.uber.org/zap"
	"os"
)

func SanityCheck(logger *zap.Logger, variables []string) {
	for _, variable := range variables {
		if _, ok := os.LookupEnv(variable); !ok {
			logger.Fatal("environment variable not set", zap.String("variable", variable))
		}
	}
}
