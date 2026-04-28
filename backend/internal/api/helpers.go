package api

import (
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

// hasDomainAccess проверяет, имеет ли администратор права на управление доменом.
// Если администратор является SuperAdmin, доступ всегда разрешен.
func hasDomainAccess(claims *auth.Claims, targetDomain string) bool {
	if claims.SuperAdmin {
		return true
	}
	var count int64
	db.DB.Model(&models.DomainAdmin{}).
		Where("username = ? AND domain = ?", claims.Username, targetDomain).
		Count(&count)
	return count > 0
}
