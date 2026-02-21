package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
)

// RequireRole creates a middleware that checks if the user has one of the required roles
// This should be used AFTER JWTAuth middleware
func RequireRole(roles ...auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetRoleFromContext(c)
		if !ok {
			abortWithForbidden(c, "role information not available")
			return
		}

		if !isRoleAllowed(role, roles) {
			abortWithForbidden(c, "insufficient permissions for this operation")
			return
		}

		c.Next()
	}
}

// RequireAdmin creates a middleware that allows only admin users
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(auth.RoleAdmin)
}

// RequireProvider creates a middleware that allows only provider users
func RequireProvider() gin.HandlerFunc {
	return RequireRole(auth.RoleProvider)
}

// RequireAdminOrProvider creates a middleware that allows admin or provider users
// This is commonly used for appointment management operations
func RequireAdminOrProvider() gin.HandlerFunc {
	return RequireRole(auth.RoleAdmin, auth.RoleProvider)
}

// RequireAny creates a middleware that allows any authenticated user
// This is essentially a no-op if JWTAuth has already passed, but provides explicit intent
func RequireAny() gin.HandlerFunc {
	return RequireRole(auth.RoleAdmin, auth.RoleUser, auth.RoleProvider)
}
