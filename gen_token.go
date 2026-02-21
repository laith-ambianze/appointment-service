package main
import (
    "fmt"
    "github.com/laith-ambianze/appointment-service/pkg/auth"
    "github.com/google/uuid"
)
func main() {
    jwtManager := auth.NewJWTManager("your-jwt-secret-key-minimum-32-characters-change-in-production")
    token, _ := jwtManager.GenerateToken(
        uuid.MustParse("4bfd6e6a-de4e-49a5-b224-6eef9f807068"),
        "my-user-001",
        auth.RoleAdmin,
    )
    fmt.Println(token)
}
