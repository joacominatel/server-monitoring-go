package logger

import (
	"fmt"
)

// formatArgs formatea argumentos variables a una cadena
func formatArgs(args ...interface{}) string {
	return fmt.Sprint(args...)
}

// formatf formatea un string con argumentos usando fmt.Sprintf
func formatf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
} 